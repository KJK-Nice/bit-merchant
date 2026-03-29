package http

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"sync"
	"time"

	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	authInfra "bitmerchant/internal/infrastructure/auth"
	"bitmerchant/internal/interfaces/templates"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type pendingRegistration struct {
	user            *domain.User
	restaurantName  string
	invitationToken string
}

type beginRegistrationRequest struct {
	DisplayName     string `json:"displayName"`
	RestaurantName  string `json:"restaurantName"`
	InvitationToken string `json:"invitationToken"`
}

type AuthHandler struct {
	webauthnSvc      *authInfra.WebAuthnService
	userRepo         domain.UserRepository
	membershipRepo   domain.MembershipRepository
	invitationRepo   domain.InvitationRepository
	sessionRepo      domain.SessionRepository
	createRestaurant *restaurant.CreateRestaurantUseCase

	mu      sync.Mutex
	pending map[string]pendingRegistration
}

func NewAuthHandler(
	webauthnSvc *authInfra.WebAuthnService,
	userRepo domain.UserRepository,
	membershipRepo domain.MembershipRepository,
	invitationRepo domain.InvitationRepository,
	sessionRepo domain.SessionRepository,
	createRestaurant *restaurant.CreateRestaurantUseCase,
) *AuthHandler {
	return &AuthHandler{
		webauthnSvc:      webauthnSvc,
		userRepo:         userRepo,
		membershipRepo:   membershipRepo,
		invitationRepo:   invitationRepo,
		sessionRepo:      sessionRepo,
		createRestaurant: createRestaurant,
		pending:          make(map[string]pendingRegistration),
	}
}

func (h *AuthHandler) GetSignup(c echo.Context) error {
	return templates.AuthSignup(getCSRFToken(c)).Render(c.Request().Context(), c.Response())
}

func (h *AuthHandler) GetLogin(c echo.Context) error {
	return templates.AuthLogin(getCSRFToken(c)).Render(c.Request().Context(), c.Response())
}

func (h *AuthHandler) GetInvite(c echo.Context) error {
	token := c.Param("token")
	invitation, err := h.invitationRepo.FindByToken(token)
	if err != nil {
		return c.String(http.StatusNotFound, "Invitation not found")
	}
	if invitation.IsUsed() || invitation.IsExpired(time.Now()) {
		return c.String(http.StatusBadRequest, "Invitation is no longer valid")
	}
	return templates.AuthInvite(getCSRFToken(c), token, string(invitation.Role)).Render(c.Request().Context(), c.Response())
}

func (h *AuthHandler) BeginRegistration(c echo.Context) error {
	var req beginRegistrationRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid payload")
	}
	if req.DisplayName == "" {
		return c.String(http.StatusBadRequest, "displayName is required")
	}

	if req.InvitationToken == "" && req.RestaurantName == "" {
		return c.String(http.StatusBadRequest, "restaurantName is required for owner signup")
	}

	sessionID, _ := c.Get("sessionID").(string)
	if sessionID == "" {
		return c.String(http.StatusUnauthorized, "session not found")
	}

	user, err := domain.NewUser(domain.UserID(uuid.NewString()), req.DisplayName)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	options, err := h.webauthnSvc.BeginRegistration(sessionID, user)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to start registration: "+err.Error())
	}

	h.mu.Lock()
	h.pending[sessionID] = pendingRegistration{
		user:            user,
		restaurantName:  req.RestaurantName,
		invitationToken: req.InvitationToken,
	}
	h.mu.Unlock()

	return c.JSON(http.StatusOK, options)
}

func (h *AuthHandler) FinishRegistration(c echo.Context) error {
	sessionID, _ := c.Get("sessionID").(string)
	if sessionID == "" {
		return c.String(http.StatusUnauthorized, "session not found")
	}

	pending, err := h.getPending(sessionID)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	credential, err := h.webauthnSvc.FinishRegistration(sessionID, pending.user, c.Request())
	if err != nil {
		return c.String(http.StatusBadRequest, "Registration failed: "+err.Error())
	}

	pending.user.AddCredential(*credential)
	if err := h.userRepo.Save(pending.user); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save user")
	}

	var restaurantID *domain.RestaurantID
	redirect := "/dashboard"

	if pending.invitationToken != "" {
		invitation, invErr := h.invitationRepo.FindByToken(pending.invitationToken)
		if invErr != nil {
			return c.String(http.StatusBadRequest, "Invitation not found")
		}
		if invitation.IsExpired(time.Now()) || invitation.IsUsed() {
			return c.String(http.StatusBadRequest, "Invitation expired or already used")
		}

		membership, memErr := domain.NewMembership(
			domain.MembershipID(uuid.NewString()),
			pending.user.ID,
			invitation.RestaurantID,
			invitation.Role,
		)
		if memErr != nil {
			return c.String(http.StatusBadRequest, memErr.Error())
		}
		if memErr = h.membershipRepo.Save(membership); memErr != nil {
			return c.String(http.StatusInternalServerError, "Failed to save membership")
		}

		now := time.Now()
		invitation.MarkUsed(pending.user.ID, now)
		_ = h.invitationRepo.Update(invitation)

		restaurantID = &invitation.RestaurantID
		if invitation.Role == domain.RoleKitchenStaff {
			redirect = "/kitchen"
		}
	} else if pending.restaurantName != "" {
		rest, restErr := h.createRestaurant.Execute(c.Request().Context(), restaurant.CreateRestaurantRequest{
			Name: pending.restaurantName,
		})
		if restErr != nil {
			return c.String(http.StatusInternalServerError, "Failed to create restaurant")
		}

		membership, memErr := domain.NewMembership(
			domain.MembershipID(uuid.NewString()),
			pending.user.ID,
			rest.ID,
			domain.RoleOwner,
		)
		if memErr != nil {
			return c.String(http.StatusBadRequest, memErr.Error())
		}
		if memErr = h.membershipRepo.Save(membership); memErr != nil {
			return c.String(http.StatusInternalServerError, "Failed to save membership")
		}
		restaurantID = &rest.ID
	}

	authSession := &domain.Session{
		ID:           sessionID,
		UserID:       &pending.user.ID,
		RestaurantID: restaurantID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	if err := h.sessionRepo.Save(authSession); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create session")
	}

	setAuthenticatedContext(c, pending.user, authSession)
	h.clearPending(sessionID)

	return c.JSON(http.StatusOK, map[string]string{
		"redirect": redirect,
	})
}

func (h *AuthHandler) BeginLogin(c echo.Context) error {
	sessionID, _ := c.Get("sessionID").(string)
	if sessionID == "" {
		return c.String(http.StatusUnauthorized, "session not found")
	}

	assertion, err := h.webauthnSvc.BeginPasskeyLogin(sessionID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to start login: "+err.Error())
	}
	return c.JSON(http.StatusOK, assertion)
}

func (h *AuthHandler) FinishLogin(c echo.Context) error {
	sessionID, _ := c.Get("sessionID").(string)
	if sessionID == "" {
		return c.String(http.StatusUnauthorized, "session not found")
	}

	user, credential, err := h.webauthnSvc.FinishPasskeyLogin(sessionID, c.Request(), func(rawID, _ []byte) (webauthn.User, error) {
		foundUser, _, findErr := h.userRepo.FindByCredentialID(rawID)
		if findErr != nil {
			return nil, findErr
		}
		return foundUser, nil
	})
	if err != nil {
		return c.String(http.StatusBadRequest, "Login failed: "+err.Error())
	}

	domainUser, ok := user.(*domain.User)
	if !ok {
		return c.String(http.StatusInternalServerError, "Invalid user type")
	}

	if credential != nil {
		domainUser.UpdateCredential(*credential)
		_ = h.userRepo.Update(domainUser)
	}

	var restaurantID *domain.RestaurantID
	memberships, memErr := h.membershipRepo.FindByUserID(domainUser.ID)
	if memErr == nil && len(memberships) > 0 {
		restaurantID = &memberships[0].RestaurantID
	}

	authSession := &domain.Session{
		ID:           sessionID,
		UserID:       &domainUser.ID,
		RestaurantID: restaurantID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	if err := h.sessionRepo.Save(authSession); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save session")
	}

	setAuthenticatedContext(c, domainUser, authSession)
	return c.JSON(http.StatusOK, map[string]string{
		"redirect": "/dashboard",
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	sessionID, _ := c.Get("sessionID").(string)
	if sessionID != "" {
		_ = h.sessionRepo.Delete(sessionID)
	}
	return c.Redirect(http.StatusFound, "/menu")
}

func (h *AuthHandler) CreateInvitation(c echo.Context) error {
	user, ok := getAuthenticatedUser(c)
	if !ok || user == nil {
		return c.String(http.StatusUnauthorized, "unauthorized")
	}

	restaurantID, err := getRestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	membership, err := h.membershipRepo.FindByUserAndRestaurant(user.ID, restaurantID)
	if err != nil || membership.Role != domain.RoleOwner {
		return c.String(http.StatusForbidden, "only owners can invite")
	}

	token, err := newInviteToken()
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to create token")
	}

	invitation, err := domain.NewInvitation(
		domain.InvitationID(uuid.NewString()),
		restaurantID,
		domain.RoleKitchenStaff,
		token,
		time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if err := h.invitationRepo.Save(invitation); err != nil {
		return c.String(http.StatusInternalServerError, "failed to save invitation")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"inviteURL": "/auth/invite/" + token,
	})
}

func newInviteToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (h *AuthHandler) getPending(sessionID string) (pendingRegistration, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	pending, ok := h.pending[sessionID]
	if !ok {
		return pendingRegistration{}, errors.New("registration session not found")
	}
	return pending, nil
}

func (h *AuthHandler) clearPending(sessionID string) {
	h.mu.Lock()
	delete(h.pending, sessionID)
	h.mu.Unlock()
}
