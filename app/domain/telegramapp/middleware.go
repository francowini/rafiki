package telegramapp

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/foundation/web"
)

const telegramSecretHeader = "X-Telegram-Bot-Api-Secret-Token"

// verifyTelegramSignature validates the webhook signature from Telegram.
func verifyTelegramSignature(webhookSecret string) web.MidFunc {
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(ctx context.Context, r *http.Request) web.Encoder {
			receivedToken := r.Header.Get(telegramSecretHeader)

			if receivedToken == "" {
				return errs.New(errs.Unauthenticated, errors.New("missing telegram signature"))
			}

			// Constant-time comparison to prevent timing attacks
			if subtle.ConstantTimeCompare([]byte(receivedToken), []byte(webhookSecret)) != 1 {
				return errs.New(errs.Unauthenticated, errors.New("invalid telegram signature"))
			}

			return next(ctx, r)
		}
	}
}
