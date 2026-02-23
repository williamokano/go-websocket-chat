package webauthn

import (
	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"

	"github.com/williamokano/example-websocket-chat/internal/user"
)

// UserAdapter wraps a user.User and their WebAuthn credentials to satisfy
// the webauthn.User interface required by the go-webauthn library.
type UserAdapter struct {
	User        *user.User
	Credentials []gowebauthn.Credential
}

func (u *UserAdapter) WebAuthnID() []byte {
	return []byte(u.User.ID)
}

func (u *UserAdapter) WebAuthnName() string {
	return u.User.Username
}

func (u *UserAdapter) WebAuthnDisplayName() string {
	return u.User.Username
}

func (u *UserAdapter) WebAuthnCredentials() []gowebauthn.Credential {
	return u.Credentials
}

// credentialDescriptors returns credential descriptors for exclusion during registration.
func (u *UserAdapter) credentialDescriptors() []protocol.CredentialDescriptor {
	descriptors := make([]protocol.CredentialDescriptor, len(u.Credentials))
	for i, cred := range u.Credentials {
		descriptors[i] = cred.Descriptor()
	}
	return descriptors
}

// toLibCredential converts a stored Credential to the go-webauthn library format.
func toLibCredential(c *Credential) gowebauthn.Credential {
	var transports []protocol.AuthenticatorTransport
	for _, t := range c.Transports {
		transports = append(transports, protocol.AuthenticatorTransport(t))
	}

	return gowebauthn.Credential{
		ID:              c.ID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Transport:       transports,
		Authenticator: gowebauthn.Authenticator{
			AAGUID:    c.AAGUID,
			SignCount: c.SignCount,
		},
	}
}

// toLibCredentials converts a slice of stored credentials to the library format.
func toLibCredentials(creds []*Credential) []gowebauthn.Credential {
	result := make([]gowebauthn.Credential, len(creds))
	for i, c := range creds {
		result[i] = toLibCredential(c)
	}
	return result
}
