package auth

import (
	"context"

	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3/framework/config"
	"flamingo.me/flamingo/v3/framework/web"
)

// BindHTTPBasicAuthIdentifier helper for dingo bindings in projects
func BindHTTPBasicAuthIdentifier(injector *dingo.Injector) {
	injector.BindMulti(new(RequestIdentifier)).To(new(HTTPBasicAuthIdentifier))
}

// HTTPBasicAuthIdentifier identifies users based on HTTP Basic Authentication header
type HTTPBasicAuthIdentifier struct {
	users map[string]string
}

// Inject user configuration
func (i *HTTPBasicAuthIdentifier) Inject(config *struct {
	Users config.Map `inject:"config:core.auth.httpbasicusers"`
}) *HTTPBasicAuthIdentifier {
	config.Users.MapInto(&i.users)
	return i
}

// HTTPBasicAuthIdentity transports a user identity, currently just the username
type HTTPBasicAuthIdentity struct {
	User string
}

// Subject is the http basic auth user
func (i *HTTPBasicAuthIdentity) Subject() string {
	return i.User
}

// Identify a user and match against the configured users
func (i *HTTPBasicAuthIdentifier) Identify(ctx context.Context, request *web.Request) Identity {
	user, pass, ok := request.Request().BasicAuth()
	if !ok {
		return nil
	}

	if userpass, ok := i.users[user]; ok && pass == userpass {
		return &HTTPBasicAuthIdentity{User: user}
	}

	return nil
}
