// Package authapp maintains the web based api for auth access.
package authapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/foundation/web"
)

type app struct {
	auth *auth.Auth
}

func newApp(ath *auth.Auth) *app {
	return &app{
		auth: ath,
	}
}

func (a *app) token(ctx context.Context, r *http.Request) web.Encoder {
	kid := web.Param(r, "kid")
	if kid == "" {
		return errs.NewFieldErrors("kid", errors.New("missing kid"))
	}

	// The BearerBasic middleware function generates the claims.
	claims := mid.GetClaims(ctx)

	tkn, err := a.auth.GenerateToken(kid, claims)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	return token{Token: tkn}
}
