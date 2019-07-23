package auth

import (
	"context"

	"flamingo.me/flamingo/v3/framework/web"
)

type (
	// RequestIdentifier resolves an identity from a web request
	RequestIdentifier interface {
		Identify(ctx context.Context, request *web.Request) Identity
	}

	// WebIdentityService calls one or more identifier to get all possible identities of a user
	WebIdentityService struct {
		identityProviders []RequestIdentifier
	}
)

// Identify the user, if any identity is found
func (s *WebIdentityService) Identify(ctx context.Context, request *web.Request) Identity {
	for _, provider := range s.identityProviders {
		if identity := provider.Identify(ctx, request); identity != nil {
			return identity
		}
	}

	return nil
}

// IdentifyAll collects all possible user identites, in case multiple are available
func (s *WebIdentityService) IdentifyAll(ctx context.Context, request *web.Request) []Identity {
	var identities []Identity

	for _, provider := range s.identityProviders {
		if identity := provider.Identify(ctx, request); identity != nil {
			identities = append(identities, identity)
		}
	}

	return identities
}
