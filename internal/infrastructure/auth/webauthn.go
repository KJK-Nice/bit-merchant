package auth

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"bitmerchant/internal/domain"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type ceremonyType string

const (
	ceremonyRegistration ceremonyType = "registration"
	ceremonyLogin        ceremonyType = "login"
)

type ceremonyState struct {
	session   webauthn.SessionData
	typ       ceremonyType
	createdAt time.Time
}

// WebAuthnService wraps go-webauthn and manages temporary ceremony sessions.
type WebAuthnService struct {
	wa       *webauthn.WebAuthn
	mu       sync.RWMutex
	ceremony map[string]ceremonyState
}

const defaultCeremonyTTL = 5 * time.Minute

// NewWebAuthnService creates a configured WebAuthn wrapper.
func NewWebAuthnService(rpID, rpDisplayName string, rpOrigins []string) (*WebAuthnService, error) {
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: rpDisplayName,
		RPOrigins:     rpOrigins,
	})
	if err != nil {
		return nil, err
	}

	return &WebAuthnService{
		wa:       wa,
		ceremony: make(map[string]ceremonyState),
	}, nil
}

// BeginRegistration starts a passkey registration ceremony.
func (s *WebAuthnService) BeginRegistration(sessionKey string, user *domain.User) (*protocol.CredentialCreation, error) {
	if sessionKey == "" {
		return nil, errors.New("session key is required")
	}
	options, data, err := s.wa.BeginRegistration(user)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.ceremony[sessionKey] = ceremonyState{session: *data, typ: ceremonyRegistration, createdAt: time.Now()}
	s.mu.Unlock()

	return options, nil
}

// FinishRegistration verifies the attestation and creates a credential.
func (s *WebAuthnService) FinishRegistration(sessionKey string, user *domain.User, req *http.Request) (*webauthn.Credential, error) {
	state, err := s.loadCeremony(sessionKey, ceremonyRegistration)
	if err != nil {
		return nil, err
	}

	credential, err := s.wa.FinishRegistration(user, state.session, req)
	if err != nil {
		return nil, err
	}

	s.deleteCeremony(sessionKey)
	return credential, nil
}

// BeginPasskeyLogin starts a discoverable login ceremony.
func (s *WebAuthnService) BeginPasskeyLogin(sessionKey string) (*protocol.CredentialAssertion, error) {
	if sessionKey == "" {
		return nil, errors.New("session key is required")
	}
	assertion, data, err := s.wa.BeginDiscoverableLogin()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.ceremony[sessionKey] = ceremonyState{session: *data, typ: ceremonyLogin, createdAt: time.Now()}
	s.mu.Unlock()

	return assertion, nil
}

// FinishPasskeyLogin verifies assertion and resolves the authenticated user.
func (s *WebAuthnService) FinishPasskeyLogin(sessionKey string, req *http.Request, resolver webauthn.DiscoverableUserHandler) (webauthn.User, *webauthn.Credential, error) {
	state, err := s.loadCeremony(sessionKey, ceremonyLogin)
	if err != nil {
		return nil, nil, err
	}

	user, credential, err := s.wa.FinishPasskeyLogin(resolver, state.session, req)
	if err != nil {
		return nil, nil, err
	}

	s.deleteCeremony(sessionKey)
	return user, credential, nil
}

func (s *WebAuthnService) loadCeremony(sessionKey string, expected ceremonyType) (ceremonyState, error) {
	if sessionKey == "" {
		return ceremonyState{}, errors.New("session key is required")
	}

	s.mu.RLock()
	state, exists := s.ceremony[sessionKey]
	s.mu.RUnlock()
	if !exists {
		return ceremonyState{}, errors.New("auth ceremony not found")
	}
	if state.typ != expected {
		return ceremonyState{}, errors.New("invalid auth ceremony type")
	}
	if !state.session.Expires.IsZero() && state.session.Expires.Before(time.Now()) {
		s.deleteCeremony(sessionKey)
		return ceremonyState{}, errors.New("auth ceremony expired")
	}
	if state.createdAt.Add(defaultCeremonyTTL).Before(time.Now()) {
		s.deleteCeremony(sessionKey)
		return ceremonyState{}, errors.New("auth ceremony expired")
	}
	return state, nil
}

func (s *WebAuthnService) deleteCeremony(sessionKey string) {
	s.mu.Lock()
	delete(s.ceremony, sessionKey)
	s.mu.Unlock()
}
