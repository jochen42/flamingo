package auth

type (
	// Identity donates an authentication object which at least identifies the authenticated subject
	Identity interface {
		Subject() string
	}

	// HasRoles adds the ability to provide roles to an identity
	HasRoles interface {
		Roles() []string
	}

	// OAuth2Identity defines an identity which is can be used to get an AccessToken vor OAuth2 flows
	OAuth2Identity interface {
		Identity
		HasRoles
		AccessToken() string
	}

	// OpenIDIdentity is an extension of OAuth2Identity which provides an IDToken on top of OAuth2
	OpenIDIdentity interface {
		OAuth2Identity
		IDToken() string
	}
)
