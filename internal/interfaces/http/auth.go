package http

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	authInfra "bitmerchant/internal/infrastructure/auth"
	"bitmerchant/internal/interfaces/http/middleware"
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
	restaurantRepo   domain.RestaurantRepository
	createRestaurant *restaurant.CreateRestaurantUseCase
	logger           *slog.Logger
	sessionOpts      middleware.SessionOptions

	mu      sync.Mutex
	pending map[string]pendingRegistration
}

func NewAuthHandler(
	webauthnSvc *authInfra.WebAuthnService,
	userRepo domain.UserRepository,
	membershipRepo domain.MembershipRepository,
	invitationRepo domain.InvitationRepository,
	sessionRepo domain.SessionRepository,
	restaurantRepo domain.RestaurantRepository,
	createRestaurant *restaurant.CreateRestaurantUseCase,
	logger *slog.Logger,
	sessionOpts middleware.SessionOptions,
) *AuthHandler {
	if logger == nil {
		logger = slog.Default()
	}

	return &AuthHandler{
		webauthnSvc:      webauthnSvc,
		userRepo:         userRepo,
		membershipRepo:   membershipRepo,
		invitationRepo:   invitationRepo,
		sessionRepo:      sessionRepo,
		restaurantRepo:   restaurantRepo,
		createRestaurant: createRestaurant,
		logger:           logger,
		sessionOpts:      sessionOpts,
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
		h.logger.Warn("BeginRegistration payload bind failed", "error", err)
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
		h.logger.Warn("BeginRegistration invalid user data", "error", err)
		return c.String(http.StatusBadRequest, "Invalid signup input")
	}

	options, err := h.webauthnSvc.BeginRegistration(sessionID, user)
	if err != nil {
		h.logger.Error("BeginRegistration failed", "error", err, "sessionID", sessionID)
		return c.String(http.StatusInternalServerError, "Failed to start registration")
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
	oldSessionID, _ := c.Get("sessionID").(string)
	if oldSessionID == "" {
		return c.String(http.StatusUnauthorized, "session not found")
	}

	pending, err := h.getPending(oldSessionID)
	if err != nil {
		h.logger.Warn("FinishRegistration pending state missing", "error", err, "sessionID", oldSessionID)
		return c.String(http.StatusBadRequest, err.Error())
	}

	credential, err := h.webauthnSvc.FinishRegistration(oldSessionID, pending.user, c.Request())
	if err != nil {
		h.logger.Warn("FinishRegistration ceremony verification failed", "error", err, "sessionID", oldSessionID)
		return c.String(http.StatusBadRequest, "Registration failed")
	}

	pending.user.AddCredential(*credential)
	if err := h.userRepo.Save(pending.user); err != nil {
		h.logger.Error("FinishRegistration save user failed", "error", err, "userID", pending.user.ID)
		return c.String(http.StatusInternalServerError, "Failed to save user")
	}

	var restaurantID *domain.RestaurantID
	redirect := "/dashboard"

	if pending.invitationToken != "" {
		invitation, invErr := h.invitationRepo.FindByToken(pending.invitationToken)
		if invErr != nil {
			h.logger.Warn("FinishRegistration invitation not found", "error", invErr, "token", pending.invitationToken)
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
			h.logger.Error("FinishRegistration save invited membership failed", "error", memErr, "userID", pending.user.ID, "restaurantID", invitation.RestaurantID)
			return c.String(http.StatusInternalServerError, "Failed to save membership")
		}

		now := time.Now()
		invitation.MarkUsed(pending.user.ID, now)
		if updateErr := h.invitationRepo.Update(invitation); updateErr != nil {
			h.logger.Error("FinishRegistration mark invitation used failed", "error", updateErr, "token", pending.invitationToken)
			return c.String(http.StatusInternalServerError, "Failed to finalize invitation")
		}

		restaurantID = &invitation.RestaurantID
		if invitation.Role == domain.RoleKitchenStaff {
			redirect = "/kitchen"
		}
	} else if pending.restaurantName != "" {
		rest, restErr := h.createRestaurant.Execute(c.Request().Context(), restaurant.CreateRestaurantRequest{
			Name: pending.restaurantName,
		})
		if restErr != nil {
			h.logger.Error("FinishRegistration create restaurant failed", "error", restErr, "userID", pending.user.ID)
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
			h.logger.Error("FinishRegistration save owner membership failed", "error", memErr, "userID", pending.user.ID, "restaurantID", rest.ID)
			return c.String(http.StatusInternalServerError, "Failed to save membership")
		}
		restaurantID = &rest.ID
	}

	authSession := &domain.Session{
		ID:           middleware.NewSessionID(),
		UserID:       &pending.user.ID,
		RestaurantID: restaurantID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(h.sessionOpts.WithDefaults().TTL),
	}
	if err := h.sessionRepo.Save(authSession); err != nil {
		h.logger.Error("FinishRegistration create auth session failed", "error", err, "userID", pending.user.ID)
		return c.String(http.StatusInternalServerError, "Failed to create session")
	}
	if oldSessionID != authSession.ID {
		_ = h.sessionRepo.Delete(oldSessionID)
	}
	c.SetCookie(middleware.NewSessionCookie(authSession.ID, h.sessionOpts))

	setAuthenticatedContext(c, pending.user, authSession)
	h.clearPending(oldSessionID)
	h.logger.Info("User registered successfully", "userID", pending.user.ID, "restaurantID", restaurantID)

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
		h.logger.Error("BeginLogin failed", "error", err, "sessionID", sessionID)
		return c.String(http.StatusInternalServerError, "Failed to start login")
	}
	return c.JSON(http.StatusOK, assertion)
}

func (h *AuthHandler) FinishLogin(c echo.Context) error {
	oldSessionID, _ := c.Get("sessionID").(string)
	if oldSessionID == "" {
		return c.String(http.StatusUnauthorized, "session not found")
	}

	user, credential, err := h.webauthnSvc.FinishPasskeyLogin(oldSessionID, c.Request(), func(rawID, _ []byte) (webauthn.User, error) {
		foundUser, _, findErr := h.userRepo.FindByCredentialID(rawID)
		if findErr != nil {
			return nil, findErr
		}
		return foundUser, nil
	})
	if err != nil {
		h.logger.Warn("FinishLogin verification failed", "error", err, "sessionID", oldSessionID)
		return c.String(http.StatusBadRequest, "Login failed")
	}

	domainUser, ok := user.(*domain.User)
	if !ok {
		h.logger.Error("FinishLogin resolved invalid user type")
		return c.String(http.StatusInternalServerError, "Invalid user type")
	}

	if credential != nil {
		domainUser.UpdateCredential(*credential)
		if updateErr := h.userRepo.Update(domainUser); updateErr != nil {
			h.logger.Error("FinishLogin user credential update failed", "error", updateErr, "userID", domainUser.ID)
		}
	}

	var (
		restaurantID *domain.RestaurantID
		redirect     = "/dashboard"
	)
	memberships, memErr := h.membershipRepo.FindByUserID(domainUser.ID)
	if memErr == nil && len(memberships) == 1 {
		restaurantID = &memberships[0].RestaurantID
		if memberships[0].Role == domain.RoleKitchenStaff {
			redirect = "/kitchen"
		}
	}
	if memErr == nil && len(memberships) > 1 {
		redirect = "/auth/select-restaurant"
	}

	authSession := &domain.Session{
		ID:           middleware.NewSessionID(),
		UserID:       &domainUser.ID,
		RestaurantID: restaurantID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(h.sessionOpts.WithDefaults().TTL),
	}
	if err := h.sessionRepo.Save(authSession); err != nil {
		h.logger.Error("FinishLogin save session failed", "error", err, "userID", domainUser.ID)
		return c.String(http.StatusInternalServerError, "Failed to save session")
	}
	if oldSessionID != authSession.ID {
		_ = h.sessionRepo.Delete(oldSessionID)
	}
	c.SetCookie(middleware.NewSessionCookie(authSession.ID, h.sessionOpts))

	setAuthenticatedContext(c, domainUser, authSession)
	h.logger.Info("User logged in successfully", "userID", domainUser.ID, "restaurantID", restaurantID)
	return c.JSON(http.StatusOK, map[string]string{
		"redirect": redirect,
	})
}

func (h *AuthHandler) GetSelectRestaurant(c echo.Context) error {
	user, ok := getAuthenticatedUser(c)
	if !ok || user == nil {
		return c.String(http.StatusUnauthorized, "unauthorized")
	}

	memberships, err := h.membershipRepo.FindByUserID(user.ID)
	if err != nil || len(memberships) == 0 {
		return c.String(http.StatusForbidden, "no restaurant memberships found")
	}

	var currentRestaurantID string
	if session, sessionOK := getSession(c); sessionOK && session != nil && session.RestaurantID != nil {
		currentRestaurantID = string(*session.RestaurantID)
	}

	options := make([]templates.RestaurantOption, 0, len(memberships))
	for _, membership := range memberships {
		option := templates.RestaurantOption{
			RestaurantID: string(membership.RestaurantID),
			Role:         string(membership.Role),
			DisplayName:  string(membership.RestaurantID),
		}
		if h.restaurantRepo != nil {
			if rest, restErr := h.restaurantRepo.FindByID(membership.RestaurantID); restErr == nil && rest != nil && rest.Name != "" {
				option.DisplayName = rest.Name
			}
		}
		options = append(options, option)
	}

	return templates.AuthSelectRestaurant(getCSRFToken(c), currentRestaurantID, options).Render(c.Request().Context(), c.Response())
}

func (h *AuthHandler) PostSelectRestaurant(c echo.Context) error {
	user, ok := getAuthenticatedUser(c)
	if !ok || user == nil {
		return c.String(http.StatusUnauthorized, "unauthorized")
	}

	sessionID, _ := c.Get("sessionID").(string)
	if sessionID == "" {
		return c.String(http.StatusUnauthorized, "session not found")
	}

	restaurantID := domain.RestaurantID(c.FormValue("restaurantID"))
	if restaurantID == "" {
		return c.String(http.StatusBadRequest, "restaurantID is required")
	}

	membership, err := h.membershipRepo.FindByUserAndRestaurant(user.ID, restaurantID)
	if err != nil {
		return c.String(http.StatusForbidden, "membership not found")
	}

	session, err := h.sessionRepo.Get(sessionID)
	if err != nil || session == nil {
		session = &domain.Session{
			ID:        sessionID,
			CreatedAt: time.Now(),
		}
	}
	session.UserID = &user.ID
	session.RestaurantID = &restaurantID
	session.ExpiresAt = time.Now().Add(h.sessionOpts.WithDefaults().TTL)
	if err := h.sessionRepo.Save(session); err != nil {
		h.logger.Error("PostSelectRestaurant save session failed", "error", err, "userID", user.ID, "restaurantID", restaurantID)
		return c.String(http.StatusInternalServerError, "failed to save session")
	}

	setAuthenticatedContext(c, user, session)
	return c.Redirect(http.StatusFound, h.redirectByRole(membership.Role))
}

func (h *AuthHandler) Logout(c echo.Context) error {
	sessionID, _ := c.Get("sessionID").(string)
	if sessionID != "" {
		_ = h.sessionRepo.Delete(sessionID)
	}
	c.SetCookie(&http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
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
		h.logger.Error("CreateInvitation token generation failed", "error", err, "userID", user.ID)
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
		h.logger.Error("CreateInvitation save failed", "error", err, "userID", user.ID, "restaurantID", restaurantID)
		return c.String(http.StatusInternalServerError, "failed to save invitation")
	}

	h.logger.Info("Invitation created", "userID", user.ID, "restaurantID", restaurantID, "role", domain.RoleKitchenStaff)
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

func (h *AuthHandler) redirectByRole(role domain.MemberRole) string {
	if role == domain.RoleKitchenStaff {
		return "/kitchen"
	}
	return "/dashboard"
}
