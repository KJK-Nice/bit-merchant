package http

import (
	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"

	authInfra "bitmerchant/internal/infrastructure/auth"
	"bitmerchant/internal/interfaces/http/middleware"
	"bitmerchant/internal/interfaces/templates"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

type pendingRegistration struct {
	user            *user.User
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
	userRepo         user.Repository
	membershipRepo   membership.Repository
	invitationRepo   invitation.Repository
	sessionRepo      session.Repository
	restaurantRepo   restaurant.Repository
	createRestaurant *restaurantCmd.CreateRestaurantUseCase
	logger           *slog.Logger
	sessionOpts      middleware.SessionOptions

	mu      sync.Mutex
	pending map[string]pendingRegistration
}

func NewAuthHandler(
	webauthnSvc *authInfra.WebAuthnService,
	userRepo user.Repository,
	membershipRepo membership.Repository,
	invitationRepo invitation.Repository,
	sessionRepo session.Repository,
	restaurantRepo restaurant.Repository,
	createRestaurant *restaurantCmd.CreateRestaurantUseCase,
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

	user, err := user.NewUser(common.UserID(uuid.NewString()), req.DisplayName)
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

	outcome, err := h.resolveRegistrationOutcome(c, c.Request().Context(), pending)
	if err != nil {
		return nil // response already written by helper
	}

	authSession := &session.Session{
		ID:           middleware.NewSessionID(),
		UserID:       &pending.user.ID,
		RestaurantID: outcome.restaurantID,
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
	h.logger.Info("User registered successfully", "userID", pending.user.ID, "restaurantID", outcome.restaurantID)

	return c.JSON(http.StatusOK, map[string]string{
		"redirect": outcome.redirect,
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

	authnUser, credential, err := h.webauthnSvc.FinishPasskeyLogin(oldSessionID, c.Request(), func(rawID, _ []byte) (webauthn.User, error) {
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

	domainUser, ok := authnUser.(*user.User)
	if !ok {
		h.logger.Error("FinishLogin resolved invalid user type")
		return c.String(http.StatusInternalServerError, "Invalid user type")
	}

	h.refreshCredential(domainUser, credential)

	var restaurantID *common.RestaurantID
	redirect := "/dashboard"
	memberships, memErr := h.membershipRepo.FindByUserID(domainUser.ID)
	if memErr == nil {
		restaurantID, redirect = resolveRedirectFromMemberships(memberships)
	}

	authSession := &session.Session{
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

func (h *AuthHandler) restaurantOptionsForMemberships(ctx context.Context, memberships []*membership.Membership) []templates.RestaurantOption {
	return RestaurantSwitchOptionsFromMemberships(ctx, memberships, h.restaurantRepo)
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

	options := h.restaurantOptionsForMemberships(c.Request().Context(), memberships)
	return templates.AuthSelectRestaurant(getCSRFToken(c), currentRestaurantID, options).Render(c.Request().Context(), c.Response())
}

func (h *AuthHandler) GetProfile(c echo.Context) error {
	user, ok := getAuthenticatedUser(c)
	if !ok || user == nil {
		return c.Redirect(http.StatusFound, "/auth/login")
	}

	memberships, err := h.membershipRepo.FindByUserID(user.ID)
	if err != nil {
		h.logger.Error("GetProfile memberships failed", "error", err, "userID", user.ID)
		return c.String(http.StatusInternalServerError, "failed to load profile")
	}
	opts := RestaurantSwitchOptionsFromMemberships(c.Request().Context(), memberships, h.restaurantRepo)
	var activeRID string
	var label string
	var activeRole string
	if rid, rerr := getRestaurantIDFromContext(c); rerr == nil {
		activeRID = string(rid)
		label = ActiveRestaurantLabel(c.Request().Context(), rid, h.restaurantRepo)
		activeRole = ActiveRestaurantRoleForMemberships(activeRID, memberships)
	}

	canCreateRestaurant := activeRole == string(common.RoleOwner)

	dn, st, ini := LayoutUserStrings(user)
	return templates.AuthProfilePage(
		getCSRFToken(c),
		"/auth/profile",
		label,
		dn, st, ini,
		user,
		opts,
		activeRole,
		canCreateRestaurant,
	).Render(c.Request().Context(), c.Response())
}

func (h *AuthHandler) GetNewRestaurant(c echo.Context) error {
	user, ok := getAuthenticatedUser(c)
	if !ok || user == nil {
		return c.Redirect(http.StatusFound, "/auth/login")
	}

	restaurantID, err := getRestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusForbidden, "restaurant context required")
	}

	membership, err := h.membershipRepo.FindByUserAndRestaurant(user.ID, restaurantID)
	if err != nil || membership == nil || membership.Role != common.RoleOwner {
		return c.String(http.StatusForbidden, "only owners can create restaurants")
	}

	memberships, err := h.membershipRepo.FindByUserID(user.ID)
	if err != nil {
		h.logger.Error("GetNewRestaurant memberships failed", "error", err, "userID", user.ID)
		return c.String(http.StatusInternalServerError, "failed to load profile")
	}
	opts := RestaurantSwitchOptionsFromMemberships(c.Request().Context(), memberships, h.restaurantRepo)
	activeRole := ActiveRestaurantRoleForMemberships(string(restaurantID), memberships)
	canCreateRestaurant := activeRole == string(common.RoleOwner)

	dn, st, ini := LayoutUserStrings(user)
	label := ActiveRestaurantLabel(c.Request().Context(), restaurantID, h.restaurantRepo)
	return templates.AuthNewRestaurantPage(
		getCSRFToken(c),
		"/auth/restaurants/new",
		label,
		dn, st, ini,
		opts,
		activeRole,
		canCreateRestaurant,
	).Render(c.Request().Context(), c.Response())
}

func (h *AuthHandler) PostNewRestaurant(c echo.Context) error {
	user, restaurantID, err := requireAuthAndRestaurant(c)
	if err != nil {
		return err
	}

	if err := h.assertOwner(user.ID, restaurantID); err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	if h.createRestaurant == nil {
		return c.String(http.StatusInternalServerError, "restaurant creation unavailable")
	}

	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return c.String(http.StatusBadRequest, "restaurant name is required")
	}

	rest, err := h.createRestaurant.Execute(c.Request().Context(), restaurantCmd.CreateRestaurantRequest{Name: name})
	if err != nil {
		h.logger.Error("PostNewRestaurant create failed", "error", err, "userID", user.ID)
		return c.String(http.StatusInternalServerError, "failed to create restaurant")
	}

	if err := h.saveOwnerMembership(user.ID, rest.ID); err != nil {
		h.logger.Error("PostNewRestaurant save membership failed", "error", err, "userID", user.ID, "restaurantID", rest.ID)
		return c.String(http.StatusInternalServerError, "failed to save membership")
	}

	session, err := h.getOrInitSession(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, "session not found")
	}
	session.UserID = &user.ID
	rid := rest.ID
	session.RestaurantID = &rid
	session.ExpiresAt = time.Now().Add(h.sessionOpts.WithDefaults().TTL)
	if err := h.sessionRepo.Save(session); err != nil {
		h.logger.Error("PostNewRestaurant save session failed", "error", err, "userID", user.ID)
		return c.String(http.StatusInternalServerError, "failed to save session")
	}

	setAuthenticatedContext(c, user, session)
	h.logger.Info("Restaurant created", "userID", user.ID, "restaurantID", rest.ID)
	return c.Redirect(http.StatusFound, "/dashboard")
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

	restaurantID := common.RestaurantID(c.FormValue("restaurantID"))
	if restaurantID == "" {
		return c.String(http.StatusBadRequest, "restaurantID is required")
	}

	membership, err := h.membershipRepo.FindByUserAndRestaurant(user.ID, restaurantID)
	if err != nil {
		return c.String(http.StatusForbidden, "membership not found")
	}

	currentSession, err := h.sessionRepo.Get(sessionID)
	if err != nil || currentSession == nil {
		currentSession = &session.Session{
			ID:        sessionID,
			CreatedAt: time.Now(),
		}
	}
	currentSession.UserID = &user.ID
	currentSession.RestaurantID = &restaurantID
	currentSession.ExpiresAt = time.Now().Add(h.sessionOpts.WithDefaults().TTL)
	if err := h.sessionRepo.Save(currentSession); err != nil {
		h.logger.Error("PostSelectRestaurant save session failed", "error", err, "userID", user.ID, "restaurantID", restaurantID)
		return c.String(http.StatusInternalServerError, "failed to save session")
	}

	setAuthenticatedContext(c, user, currentSession)
	return c.Redirect(http.StatusFound, h.redirectByRole(membership.Role))
}

func (h *AuthHandler) Logout(c echo.Context) error {
	sessionID, _ := c.Get("sessionID").(string)
	if sessionID != "" {
		_ = h.sessionRepo.Delete(sessionID)
	}
	opts := h.sessionOpts.WithDefaults()
	c.SetCookie(&http.Cookie{
		Name:     opts.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
	if opts.CookieName != opts.LegacyCookieName {
		c.SetCookie(&http.Cookie{
			Name:     opts.LegacyCookieName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
		})
	}
	return c.Redirect(http.StatusFound, "/")
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
	if err != nil || membership.Role != common.RoleOwner {
		return c.String(http.StatusForbidden, "only owners can invite")
	}

	token, err := newInviteToken()
	if err != nil {
		h.logger.Error("CreateInvitation token generation failed", "error", err, "userID", user.ID)
		return c.String(http.StatusInternalServerError, "failed to create token")
	}

	invitation, err := invitation.NewInvitation(
		common.InvitationID(uuid.NewString()),
		restaurantID,
		common.RoleKitchenStaff,
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

	h.logger.Info("Invitation created", "userID", user.ID, "restaurantID", restaurantID, "role", common.RoleKitchenStaff)
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

func (h *AuthHandler) redirectByRole(role common.MemberRole) string {
	if role == common.RoleKitchenStaff {
		return "/kitchen"
	}
	return "/dashboard"
}

// registrationOutcome holds the post-registration redirect target and restaurant context.
type registrationOutcome struct {
	restaurantID *common.RestaurantID
	redirect     string
}

func (h *AuthHandler) resolveRegistrationOutcome(c echo.Context, ctx context.Context, pending pendingRegistration) (registrationOutcome, error) {
	if pending.invitationToken != "" {
		return h.applyInvitationRegistration(c, ctx, pending)
	}
	if pending.restaurantName != "" {
		return h.applyNewRestaurantRegistration(c, ctx, pending)
	}
	return registrationOutcome{redirect: "/dashboard"}, nil
}

func (h *AuthHandler) applyInvitationRegistration(c echo.Context, ctx context.Context, pending pendingRegistration) (registrationOutcome, error) {
	invitation, err := h.invitationRepo.FindByToken(pending.invitationToken)
	if err != nil {
		h.logger.Warn("FinishRegistration invitation not found", "error", err, "token", pending.invitationToken)
		_ = c.String(http.StatusBadRequest, "Invitation not found")
		return registrationOutcome{}, err
	}
	if invitation.IsExpired(time.Now()) || invitation.IsUsed() {
		_ = c.String(http.StatusBadRequest, "Invitation expired or already used")
		return registrationOutcome{}, errors.New("invitation expired or already used")
	}
	membership, err := membership.NewMembership(
		common.MembershipID(uuid.NewString()),
		pending.user.ID,
		invitation.RestaurantID,
		invitation.Role,
	)
	if err != nil {
		_ = c.String(http.StatusBadRequest, err.Error())
		return registrationOutcome{}, err
	}
	if err = h.membershipRepo.Save(membership); err != nil {
		h.logger.Error("FinishRegistration save invited membership failed", "error", err, "userID", pending.user.ID, "restaurantID", invitation.RestaurantID)
		_ = c.String(http.StatusInternalServerError, "Failed to save membership")
		return registrationOutcome{}, err
	}
	invitation.MarkUsed(pending.user.ID, time.Now())
	if err := h.invitationRepo.Update(invitation); err != nil {
		h.logger.Error("FinishRegistration mark invitation used failed", "error", err, "token", pending.invitationToken)
		_ = c.String(http.StatusInternalServerError, "Failed to finalize invitation")
		return registrationOutcome{}, err
	}
	redirect := "/dashboard"
	if invitation.Role == common.RoleKitchenStaff {
		redirect = "/kitchen"
	}
	return registrationOutcome{restaurantID: &invitation.RestaurantID, redirect: redirect}, nil
}

func (h *AuthHandler) applyNewRestaurantRegistration(c echo.Context, ctx context.Context, pending pendingRegistration) (registrationOutcome, error) {
	rest, err := h.createRestaurant.Execute(ctx, restaurantCmd.CreateRestaurantRequest{Name: pending.restaurantName})
	if err != nil {
		h.logger.Error("FinishRegistration create restaurant failed", "error", err, "userID", pending.user.ID)
		_ = c.String(http.StatusInternalServerError, "Failed to create restaurant")
		return registrationOutcome{}, err
	}
	membership, err := membership.NewMembership(
		common.MembershipID(uuid.NewString()),
		pending.user.ID,
		rest.ID,
		common.RoleOwner,
	)
	if err != nil {
		_ = c.String(http.StatusBadRequest, err.Error())
		return registrationOutcome{}, err
	}
	if err = h.membershipRepo.Save(membership); err != nil {
		h.logger.Error("FinishRegistration save owner membership failed", "error", err, "userID", pending.user.ID, "restaurantID", rest.ID)
		_ = c.String(http.StatusInternalServerError, "Failed to save membership")
		return registrationOutcome{}, err
	}
	return registrationOutcome{restaurantID: &rest.ID, redirect: "/dashboard"}, nil
}

func (h *AuthHandler) refreshCredential(user *user.User, credential *webauthn.Credential) {
	if credential == nil {
		return
	}
	user.UpdateCredential(*credential)
	if err := h.userRepo.Update(user); err != nil {
		h.logger.Error("FinishLogin user credential update failed", "error", err, "userID", user.ID)
	}
}

func resolveRedirectFromMemberships(memberships []*membership.Membership) (*common.RestaurantID, string) {
	if len(memberships) == 1 {
		redirect := "/dashboard"
		if memberships[0].Role == common.RoleKitchenStaff {
			redirect = "/kitchen"
		}
		return &memberships[0].RestaurantID, redirect
	}
	if len(memberships) > 1 {
		return nil, "/auth/select-restaurant"
	}
	return nil, "/dashboard"
}

func requireAuthAndRestaurant(c echo.Context) (*user.User, common.RestaurantID, error) {
	user, ok := getAuthenticatedUser(c)
	if !ok || user == nil {
		return nil, "", echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	restaurantID, err := getRestaurantIDFromContext(c)
	if err != nil {
		return nil, "", echo.NewHTTPError(http.StatusForbidden, "restaurant context required")
	}
	return user, restaurantID, nil
}

func (h *AuthHandler) assertOwner(userID common.UserID, restaurantID common.RestaurantID) error {
	membership, err := h.membershipRepo.FindByUserAndRestaurant(userID, restaurantID)
	if err != nil || membership == nil || membership.Role != common.RoleOwner {
		return errors.New("only owners can create restaurants")
	}
	return nil
}

func (h *AuthHandler) saveOwnerMembership(userID common.UserID, restaurantID common.RestaurantID) error {
	membership, err := membership.NewMembership(
		common.MembershipID(uuid.NewString()),
		userID,
		restaurantID,
		common.RoleOwner,
	)
	if err != nil {
		return err
	}
	return h.membershipRepo.Save(membership)
}

func (h *AuthHandler) getOrInitSession(c echo.Context) (*session.Session, error) {
	sessionID, _ := c.Get("sessionID").(string)
	if sessionID == "" {
		return nil, errors.New("session not found")
	}
	currentSession, err := h.sessionRepo.Get(sessionID)
	if err != nil || currentSession == nil {
		return &session.Session{
			ID:        sessionID,
			CreatedAt: time.Now(),
		}, nil
	}
	return currentSession, nil
}
