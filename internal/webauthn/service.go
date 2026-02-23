package webauthn

import (
	"context"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"

	"github.com/williamokano/example-websocket-chat/internal/auth"
	"github.com/williamokano/example-websocket-chat/internal/user"
)

// Service orchestrates WebAuthn ceremonies, credential storage, and JWT issuance.
type Service struct {
	webAuthn     *gowebauthn.WebAuthn
	credRepo     Repo
	userRepo     user.Repo
	sessionStore *SessionStore
	jwtService   *auth.JWTService
}

func NewService(
	webAuthn *gowebauthn.WebAuthn,
	credRepo Repo,
	userRepo user.Repo,
	sessionStore *SessionStore,
	jwtService *auth.JWTService,
) *Service {
	return &Service{
		webAuthn:     webAuthn,
		credRepo:     credRepo,
		userRepo:     userRepo,
		sessionStore: sessionStore,
		jwtService:   jwtService,
	}
}

// BeginRegistration starts the WebAuthn registration ceremony for a new passwordless account.
func (s *Service) BeginRegistration(ctx context.Context, username string) (*protocol.CredentialCreation, string, error) {
	if len(username) < 3 || len(username) > 50 {
		return nil, "", fmt.Errorf("username must be between 3 and 50 characters")
	}

	// Create user with empty password (passwordless)
	u, err := s.userRepo.Create(ctx, username, "")
	if err != nil {
		return nil, "", fmt.Errorf("creating user: %w", err)
	}

	adapter := &UserAdapter{User: u}
	options, session, err := s.webAuthn.BeginRegistration(adapter,
		gowebauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred),
	)
	if err != nil {
		return nil, "", fmt.Errorf("beginning registration: %w", err)
	}

	sessionID, err := s.sessionStore.Save(ctx, session)
	if err != nil {
		return nil, "", fmt.Errorf("saving session: %w", err)
	}

	return options, sessionID, nil
}

// FinishRegistration completes the registration ceremony and returns an auth token.
func (s *Service) FinishRegistration(ctx context.Context, sessionID string, r *protocol.ParsedCredentialCreationData) (*auth.AuthResponse, error) {
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("retrieving session: %w", err)
	}

	userID := string(session.UserID)
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}

	existingCreds, err := s.credRepo.FindByUserID(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("finding credentials: %w", err)
	}

	adapter := &UserAdapter{User: u, Credentials: toLibCredentials(existingCreds)}
	credential, err := s.webAuthn.CreateCredential(adapter, *session, r)
	if err != nil {
		return nil, fmt.Errorf("creating credential: %w", err)
	}

	if err := s.storeCredential(ctx, u.ID, credential, ""); err != nil {
		return nil, err
	}

	return s.issueToken(u)
}

// BeginLogin starts the WebAuthn login ceremony (discoverable credentials / passkey).
func (s *Service) BeginLogin(ctx context.Context) (*protocol.CredentialAssertion, string, error) {
	options, session, err := s.webAuthn.BeginDiscoverableLogin()
	if err != nil {
		return nil, "", fmt.Errorf("beginning login: %w", err)
	}

	sessionID, err := s.sessionStore.Save(ctx, session)
	if err != nil {
		return nil, "", fmt.Errorf("saving session: %w", err)
	}

	return options, sessionID, nil
}

// FinishLogin completes the login ceremony and returns an auth token.
func (s *Service) FinishLogin(ctx context.Context, sessionID string, r *protocol.ParsedCredentialAssertionData) (*auth.AuthResponse, error) {
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("retrieving session: %w", err)
	}

	handler := func(rawID, userHandle []byte) (gowebauthn.User, error) {
		userID := string(userHandle)
		u, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("finding user: %w", err)
		}
		creds, err := s.credRepo.FindByUserID(ctx, u.ID)
		if err != nil {
			return nil, fmt.Errorf("finding credentials: %w", err)
		}
		return &UserAdapter{User: u, Credentials: toLibCredentials(creds)}, nil
	}

	_, credential, err := s.webAuthn.ValidatePasskeyLogin(handler, *session, r)
	if err != nil {
		return nil, fmt.Errorf("validating login: %w", err)
	}

	// Update sign count
	if err := s.credRepo.UpdateSignCount(ctx, credential.ID, credential.Authenticator.SignCount); err != nil {
		return nil, fmt.Errorf("updating sign count: %w", err)
	}

	// Resolve user from the credential
	storedCred, err := s.credRepo.FindByCredentialID(ctx, credential.ID)
	if err != nil {
		return nil, fmt.Errorf("finding stored credential: %w", err)
	}

	u, err := s.userRepo.FindByID(ctx, storedCred.UserID)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}

	return s.issueToken(u)
}

// BeginAddCredential starts adding a passkey to an existing authenticated user.
func (s *Service) BeginAddCredential(ctx context.Context, userID string) (*protocol.CredentialCreation, string, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("finding user: %w", err)
	}

	existingCreds, err := s.credRepo.FindByUserID(ctx, u.ID)
	if err != nil {
		return nil, "", fmt.Errorf("finding credentials: %w", err)
	}

	adapter := &UserAdapter{User: u, Credentials: toLibCredentials(existingCreds)}
	options, session, err := s.webAuthn.BeginRegistration(adapter,
		gowebauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred),
		gowebauthn.WithExclusions(adapter.credentialDescriptors()),
	)
	if err != nil {
		return nil, "", fmt.Errorf("beginning registration: %w", err)
	}

	sessionID, err := s.sessionStore.Save(ctx, session)
	if err != nil {
		return nil, "", fmt.Errorf("saving session: %w", err)
	}

	return options, sessionID, nil
}

// FinishAddCredential completes adding a passkey to an existing authenticated user.
func (s *Service) FinishAddCredential(ctx context.Context, userID, sessionID, friendlyName string, r *protocol.ParsedCredentialCreationData) (*CredentialInfo, error) {
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("retrieving session: %w", err)
	}

	// Verify the session belongs to this user
	if string(session.UserID) != userID {
		return nil, fmt.Errorf("session does not belong to this user")
	}

	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}

	existingCreds, err := s.credRepo.FindByUserID(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("finding credentials: %w", err)
	}

	adapter := &UserAdapter{User: u, Credentials: toLibCredentials(existingCreds)}
	credential, err := s.webAuthn.CreateCredential(adapter, *session, r)
	if err != nil {
		return nil, fmt.Errorf("creating credential: %w", err)
	}

	if err := s.storeCredential(ctx, u.ID, credential, friendlyName); err != nil {
		return nil, err
	}

	return &CredentialInfo{
		ID:           encodeBase64URL(credential.ID),
		FriendlyName: friendlyName,
		CreatedAt:    time.Now(),
	}, nil
}

// ListCredentials returns all passkeys for a user.
func (s *Service) ListCredentials(ctx context.Context, userID string) ([]*CredentialInfo, error) {
	creds, err := s.credRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding credentials: %w", err)
	}

	infos := make([]*CredentialInfo, len(creds))
	for i, c := range creds {
		infos[i] = &CredentialInfo{
			ID:           encodeBase64URL(c.ID),
			FriendlyName: c.FriendlyName,
			CreatedAt:    c.CreatedAt,
			LastUsedAt:   c.LastUsedAt,
		}
	}
	return infos, nil
}

// DeleteCredential removes a passkey.
func (s *Service) DeleteCredential(ctx context.Context, userID, credentialIDBase64 string) error {
	credID, err := decodeBase64URL(credentialIDBase64)
	if err != nil {
		return fmt.Errorf("invalid credential ID: %w", err)
	}
	return s.credRepo.Delete(ctx, credID, userID)
}

func (s *Service) storeCredential(ctx context.Context, userID string, cred *gowebauthn.Credential, friendlyName string) error {
	var transports []string
	for _, t := range cred.Transport {
		transports = append(transports, string(t))
	}

	stored := &Credential{
		ID:              cred.ID,
		UserID:          userID,
		PublicKey:       cred.PublicKey,
		AttestationType: cred.AttestationType,
		AAGUID:          cred.Authenticator.AAGUID,
		SignCount:       cred.Authenticator.SignCount,
		Transports:      transports,
		FriendlyName:    friendlyName,
		CreatedAt:       time.Now(),
	}

	if err := s.credRepo.Create(ctx, stored); err != nil {
		return fmt.Errorf("storing credential: %w", err)
	}
	return nil
}

func (s *Service) issueToken(u *user.User) (*auth.AuthResponse, error) {
	token, err := s.jwtService.GenerateToken(u.ID, u.Username)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	u.Password = ""
	return &auth.AuthResponse{Token: token, User: *u}, nil
}
