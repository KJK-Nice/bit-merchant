package http

import (
	authapp "bitmerchant/internal/auth/app"
	authcommand "bitmerchant/internal/auth/app/command"
	authquery "bitmerchant/internal/auth/app/query"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/common/http/middleware"

	authInfra "bitmerchant/internal/infrastructure/auth"
	"bitmerchant/internal/interfaces/templates"
	"context"
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
	app         *authapp.Application
	webauthnSvc *authInfra.WebAuthnService
	logger      *slog.Logger
	sessionOpts middleware.SessionOptions

	mu      sync.Mutex
	pending map[string]pendingRegistration
}

func NewAuthHandler(
	webauthnSvc *authInfra.WebAuthnService,
	app *authapp.Application,
	logger *slog.Logger,
	sessionOpts middleware.SessionOptions,
) *AuthHandler {
	if logger == nil {
		logger = slog.Default()
	}
	if app == nil {
		panic("nil auth app")
	}

	return &AuthHandler{
		app:         app,
		webauthnSvc: webauthnSvc,
		logger:      logger,
		sessionOpts: sessionOpts,
		pending:     make(map[string]pendingRegistration),
	}
}

func (h *AuthHandler) GetSignup(c echo.Context) error {
	return templates.AuthSignup(commonhttp.CSRFToken(c)).Render(c.Request().Context(), c.Response())
}

func (h *AuthHandler) GetLogin(c echo.Context) error {
	return templates.AuthLogin(commonhttp.CSRFToken(c)).Render(c.Request().Context(), c.Response())
}

func (h *AuthHandler) GetInvite(c echo.Context) error {
	token := c.Param("token")
	inv, err := h.app.Queries.InvitationForToken.Handle(c.Request().Context(), authquery.InvitationForToken{Token: token})
	if err != nil {
		return c.String(http.StatusNotFound, "Invitation not found")
	}
	if inv.IsUsed() || inv.IsExpired(time.Now()) {
		return c.String(http.StatusBadRequest, "Invitation is no longer valid")
	}
	return templates.AuthInvite(commonhttp.CSRFToken(c), token, string(inv.Role)).Render(c.Request().Context(), c.Response())
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
	if err := h.app.User.Save(pending.user); err != nil {
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
	if err := h.app.Session.Save(authSession); err != nil {
		h.logger.Error("FinishRegistration create auth session failed", "error", err, "userID", pending.user.ID)
		return c.String(http.StatusInternalServerError, "Failed to create session")
	}
	if oldSessionID != authSession.ID {
		_ = h.app.Session.Delete(oldSessionID)
	}
	c.SetCookie(middleware.NewSessionCookie(authSession.ID, h.sessionOpts))

	commonhttp.SetAuthenticatedContext(c, pending.user, authSession)
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
		foundUser, _, findErr := h.app.User.FindByCredentialID(rawID)
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
	memberships, memErr := h.app.Queries.MembershipsForUser.Handle(c.Request().Context(), authquery.MembershipsForUser{UserID: domainUser.ID})
	if memErr == nil {
		restaurantID, redirect = authquery.PostLoginRestaurantContext(memberships)
	}

	authSession := &session.Session{
		ID:           middleware.NewSessionID(),
		UserID:       &domainUser.ID,
		RestaurantID: restaurantID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(h.sessionOpts.WithDefaults().TTL),
	}
	if err := h.app.Session.Save(authSession); err != nil {
		h.logger.Error("FinishLogin save session failed", "error", err, "userID", domainUser.ID)
		return c.String(http.StatusInternalServerError, "Failed to save session")
	}
	if oldSessionID != authSession.ID {
		_ = h.app.Session.Delete(oldSessionID)
	}
	c.SetCookie(middleware.NewSessionCookie(authSession.ID, h.sessionOpts))

	commonhttp.SetAuthenticatedContext(c, domainUser, authSession)
	h.logger.Info("User logged in successfully", "userID", domainUser.ID, "restaurantID", restaurantID)
	return c.JSON(http.StatusOK, map[string]string{
		"redirect": redirect,
	})
}

func (h *AuthHandler) restaurantOptionsForMemberships(ctx context.Context, memberships []*membership.Membership) []templates.RestaurantOption {
	return commonhttp.RestaurantSwitchOptionsFromMemberships(ctx, memberships, h.app.Restaurant)
}

func (h *AuthHandler) GetSelectRestaurant(c echo.Context) error {
	user, ok := commonhttp.GetAuthenticatedUser(c)
	if !ok || user == nil {
		return c.String(http.StatusUnauthorized, "unauthorized")
	}

	memberships, err := h.app.Queries.MembershipsForUser.Handle(c.Request().Context(), authquery.MembershipsForUser{UserID: user.ID})
	if err != nil || len(memberships) == 0 {
		return c.String(http.StatusForbidden, "no restaurant memberships found")
	}

	var currentRestaurantID string
	if sess, sessionOK := commonhttp.GetSession(c); sessionOK && sess != nil && sess.RestaurantID != nil {
		currentRestaurantID = string(*sess.RestaurantID)
	}

	options := h.restaurantOptionsForMemberships(c.Request().Context(), memberships)
	return templates.AuthSelectRestaurant(commonhttp.CSRFToken(c), currentRestaurantID, options).Render(c.Request().Context(), c.Response())
}

func (h *AuthHandler) GetProfile(c echo.Context) error {
	user, ok := commonhttp.GetAuthenticatedUser(c)
	if !ok || user == nil {
		return c.Redirect(http.StatusFound, "/auth/login")
	}

	memberships, err := h.app.Queries.MembershipsForUser.Handle(c.Request().Context(), authquery.MembershipsForUser{UserID: user.ID})
	if err != nil {
		h.logger.Error("GetProfile memberships failed", "error", err, "userID", user.ID)
		return c.String(http.StatusInternalServerError, "failed to load profile")
	}
	opts := commonhttp.RestaurantSwitchOptionsFromMemberships(c.Request().Context(), memberships, h.app.Restaurant)
	var activeRID string
	var label string
	var activeRole string
	if rid, rerr := commonhttp.RestaurantIDFromContext(c); rerr == nil {
		activeRID = string(rid)
		label = commonhttp.ActiveRestaurantLabel(c.Request().Context(), rid, h.app.Restaurant)
		activeRole = commonhttp.ActiveRestaurantRoleForMemberships(activeRID, memberships)
	}

	canCreateRestaurant := activeRole == string(common.RoleOwner)

	dn, st, ini := commonhttp.LayoutUserStrings(user)
	return templates.AuthProfilePage(
		commonhttp.CSRFToken(c),
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
	user, ok := commonhttp.GetAuthenticatedUser(c)
	if !ok || user == nil {
		return c.Redirect(http.StatusFound, "/auth/login")
	}

	restaurantID, err := commonhttp.RestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusForbidden, "restaurant context required")
	}

	membership, err := h.app.Membership.FindByUserAndRestaurant(user.ID, restaurantID)
	if err != nil || membership == nil || membership.Role != common.RoleOwner {
		return c.String(http.StatusForbidden, "only owners can create restaurants")
	}

	memberships, err := h.app.Queries.MembershipsForUser.Handle(c.Request().Context(), authquery.MembershipsForUser{UserID: user.ID})
	if err != nil {
		h.logger.Error("GetNewRestaurant memberships failed", "error", err, "userID", user.ID)
		return c.String(http.StatusInternalServerError, "failed to load profile")
	}
	opts := commonhttp.RestaurantSwitchOptionsFromMemberships(c.Request().Context(), memberships, h.app.Restaurant)
	activeRole := commonhttp.ActiveRestaurantRoleForMemberships(string(restaurantID), memberships)
	canCreateRestaurant := activeRole == string(common.RoleOwner)

	dn, st, ini := commonhttp.LayoutUserStrings(user)
	label := commonhttp.ActiveRestaurantLabel(c.Request().Context(), restaurantID, h.app.Restaurant)
	return templates.AuthNewRestaurantPage(
		commonhttp.CSRFToken(c),
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

	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return c.String(http.StatusBadRequest, "restaurant name is required")
	}
	currencyCode := strings.TrimSpace(c.FormValue("baseCurrency"))

	rest, err := h.app.Commands.CreateRestaurantUnderOwner.Handle(c.Request().Context(), authcommand.CreateRestaurantUnderOwner{
		OwnerUserID:              user.ID,
		OwnerContextRestaurantID: restaurantID,
		Name:                     name,
		CurrencyCode:             currencyCode,
	})
	if err != nil {
		if errors.Is(err, authcommand.ErrNotRestaurantOwner) {
			return c.String(http.StatusForbidden, err.Error())
		}
		h.logger.Error("PostNewRestaurant create failed", "error", err, "userID", user.ID)
		return c.String(http.StatusInternalServerError, "failed to create restaurant")
	}

	session, err := h.getOrInitSession(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, "session not found")
	}
	session.UserID = &user.ID
	rid := rest.ID
	session.RestaurantID = &rid
	session.ExpiresAt = time.Now().Add(h.sessionOpts.WithDefaults().TTL)
	if err := h.app.Session.Save(session); err != nil {
		h.logger.Error("PostNewRestaurant save session failed", "error", err, "userID", user.ID)
		return c.String(http.StatusInternalServerError, "failed to save session")
	}

	commonhttp.SetAuthenticatedContext(c, user, session)
	h.logger.Info("Restaurant created", "userID", user.ID, "restaurantID", rest.ID)
	return c.Redirect(http.StatusFound, "/dashboard")
}

func (h *AuthHandler) PostSelectRestaurant(c echo.Context) error {
	user, ok := commonhttp.GetAuthenticatedUser(c)
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

	currentSession, err := h.app.Commands.SwitchActiveRestaurant.Handle(c.Request().Context(), authcommand.SwitchActiveRestaurant{
		SessionID:    sessionID,
		UserID:       user.ID,
		RestaurantID: restaurantID,
		SessionTTL:   h.sessionOpts.WithDefaults().TTL,
	})
	if err != nil {
		if errors.Is(err, authcommand.ErrMembershipNotFound) {
			return c.String(http.StatusForbidden, "membership not found")
		}
		h.logger.Error("PostSelectRestaurant save session failed", "error", err, "userID", user.ID, "restaurantID", restaurantID)
		return c.String(http.StatusInternalServerError, "failed to save session")
	}

	mem, err := h.app.Membership.FindByUserAndRestaurant(user.ID, restaurantID)
	if err != nil || mem == nil {
		return c.String(http.StatusForbidden, "membership not found")
	}

	commonhttp.SetAuthenticatedContext(c, user, currentSession)
	return c.Redirect(http.StatusFound, h.redirectByRole(mem.Role))
}

func (h *AuthHandler) Logout(c echo.Context) error {
	sessionID, _ := c.Get("sessionID").(string)
	_ = h.app.Commands.EndCustomerSession.Handle(c.Request().Context(), authcommand.EndCustomerSession{SessionID: sessionID})
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
	user, ok := commonhttp.GetAuthenticatedUser(c)
	if !ok || user == nil {
		return c.String(http.StatusUnauthorized, "unauthorized")
	}

	restaurantID, err := commonhttp.RestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	res, err := h.app.Commands.IssueKitchenStaffInvitation.Handle(c.Request().Context(), authcommand.IssueKitchenStaffInvitation{
		OwnerUserID:  user.ID,
		RestaurantID: restaurantID,
	})
	if err != nil {
		if errors.Is(err, authcommand.ErrNotRestaurantOwner) {
			return c.String(http.StatusForbidden, "only owners can invite")
		}
		h.logger.Error("CreateInvitation save failed", "error", err, "userID", user.ID, "restaurantID", restaurantID)
		return c.String(http.StatusInternalServerError, "failed to save invitation")
	}

	h.logger.Info("Invitation created", "userID", user.ID, "restaurantID", restaurantID, "role", common.RoleKitchenStaff)
	return c.JSON(http.StatusOK, map[string]string{
		"inviteURL": "/auth/invite/" + res.Token,
	})
}

type passwordLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type passwordRegisterRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	DisplayName     string `json:"displayName"`
	RestaurantName  string `json:"restaurantName"`
	InvitationToken string `json:"invitationToken"`
}

func (h *AuthHandler) PostRegisterPassword(c echo.Context) error {
	var req passwordRegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "invalid payload")
	}

	outcome, err := h.app.Commands.RegisterWithPassword.Handle(c.Request().Context(), authcommand.RegisterWithPassword{
		Email:           req.Email,
		Password:        req.Password,
		DisplayName:     req.DisplayName,
		RestaurantName:  req.RestaurantName,
		InvitationToken: req.InvitationToken,
	})
	if err != nil {
		switch {
		case errors.Is(err, authcommand.ErrEmailAlreadyTaken):
			return c.String(http.StatusConflict, "email already registered")
		case errors.Is(err, authcommand.ErrInvitationNotFound):
			return c.String(http.StatusBadRequest, "invitation not found")
		case errors.Is(err, authcommand.ErrInvitationNotUsable):
			return c.String(http.StatusBadRequest, "invitation expired or already used")
		default:
			h.logger.Error("PostRegisterPassword failed", "error", err)
			return c.String(http.StatusBadRequest, err.Error())
		}
	}

	u, err := h.app.User.FindByEmail(req.Email)
	if err != nil {
		h.logger.Error("PostRegisterPassword user lookup after save failed", "error", err)
		return c.String(http.StatusInternalServerError, "failed to load user")
	}

	redirect, err := h.issueAuthSession(c, u, outcome.RestaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to create session")
	}
	if outcome.Redirect != "" {
		redirect = outcome.Redirect
	}
	return c.JSON(http.StatusOK, map[string]string{"redirect": redirect})
}

func (h *AuthHandler) PostLoginPassword(c echo.Context) error {
	var req passwordLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "invalid payload")
	}

	result, err := h.app.Commands.LoginWithPassword.Handle(c.Request().Context(), authcommand.LoginWithPassword{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return c.String(http.StatusUnauthorized, "invalid email or password")
	}

	var restaurantID *common.RestaurantID
	redirect := "/dashboard"
	memberships, memErr := h.app.Queries.MembershipsForUser.Handle(c.Request().Context(), authquery.MembershipsForUser{UserID: result.User.ID})
	if memErr == nil {
		restaurantID, redirect = authquery.PostLoginRestaurantContext(memberships)
	}

	if _, err := h.issueAuthSession(c, result.User, restaurantID); err != nil {
		return c.String(http.StatusInternalServerError, "failed to create session")
	}
	return c.JSON(http.StatusOK, map[string]string{"redirect": redirect})
}

// issueAuthSession creates a new auth session, sets the cookie, and returns the redirect path.
func (h *AuthHandler) issueAuthSession(c echo.Context, u *user.User, restaurantID *common.RestaurantID) (string, error) {
	oldSessionID, _ := c.Get("sessionID").(string)
	authSession := &session.Session{
		ID:           middleware.NewSessionID(),
		UserID:       &u.ID,
		RestaurantID: restaurantID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(h.sessionOpts.WithDefaults().TTL),
	}
	if err := h.app.Session.Save(authSession); err != nil {
		h.logger.Error("issueAuthSession save failed", "error", err, "userID", u.ID)
		return "", err
	}
	if oldSessionID != "" && oldSessionID != authSession.ID {
		_ = h.app.Session.Delete(oldSessionID)
	}
	c.SetCookie(middleware.NewSessionCookie(authSession.ID, h.sessionOpts))
	commonhttp.SetAuthenticatedContext(c, u, authSession)
	return "/dashboard", nil
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
		out, err := h.app.Commands.AcceptInvitation.Handle(ctx, authcommand.AcceptInvitationForUser{
			NewUserID:       pending.user.ID,
			InvitationToken: pending.invitationToken,
		})
		if err != nil {
			switch {
			case errors.Is(err, authcommand.ErrInvitationNotFound):
				_ = c.String(http.StatusBadRequest, "Invitation not found")
			case errors.Is(err, authcommand.ErrInvitationNotUsable):
				_ = c.String(http.StatusBadRequest, "Invitation expired or already used")
			default:
				_ = c.String(http.StatusBadRequest, err.Error())
			}
			return registrationOutcome{}, err
		}
		return registrationOutcome{restaurantID: out.RestaurantID, redirect: out.Redirect}, nil
	}
	if pending.restaurantName != "" {
		out, err := h.app.Commands.CompleteSignupNewRestaurant.Handle(ctx, authcommand.CompleteSignupNewRestaurant{
			OwnerUserID:    pending.user.ID,
			RestaurantName: pending.restaurantName,
		})
		if err != nil {
			h.logger.Error("FinishRegistration create restaurant failed", "error", err, "userID", pending.user.ID)
			_ = c.String(http.StatusInternalServerError, "Failed to create restaurant")
			return registrationOutcome{}, err
		}
		return registrationOutcome{restaurantID: out.RestaurantID, redirect: out.Redirect}, nil
	}
	return registrationOutcome{redirect: "/dashboard"}, nil
}

func (h *AuthHandler) refreshCredential(user *user.User, credential *webauthn.Credential) {
	if credential == nil {
		return
	}
	user.UpdateCredential(*credential)
	if err := h.app.User.Update(user); err != nil {
		h.logger.Error("FinishLogin user credential update failed", "error", err, "userID", user.ID)
	}
}

func requireAuthAndRestaurant(c echo.Context) (*user.User, common.RestaurantID, error) {
	user, ok := commonhttp.GetAuthenticatedUser(c)
	if !ok || user == nil {
		return nil, "", echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	restaurantID, err := commonhttp.RestaurantIDFromContext(c)
	if err != nil {
		return nil, "", echo.NewHTTPError(http.StatusForbidden, "restaurant context required")
	}
	return user, restaurantID, nil
}

func (h *AuthHandler) getOrInitSession(c echo.Context) (*session.Session, error) {
	sessionID, _ := c.Get("sessionID").(string)
	if sessionID == "" {
		return nil, errors.New("session not found")
	}
	currentSession, err := h.app.Session.Get(sessionID)
	if err != nil || currentSession == nil {
		return &session.Session{
			ID:        sessionID,
			CreatedAt: time.Now(),
		}, nil
	}
	return currentSession, nil
}
