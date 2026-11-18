// Package httpapi menyediakan REST API dan SPA VibeLog.
package httpapi

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/enpidi26/vibelog/internal/config"
	"github.com/enpidi26/vibelog/internal/email"
	"github.com/enpidi26/vibelog/internal/ratelimit"
)

// Deps adalah dependensi yang dibutuhkan HTTP layer.
type Deps struct {
	Config config.Config
	DB     *sql.DB
	Mailer email.Mailer
}

type Server struct {
	cfg       config.Config
	db        *sql.DB
	mailer    email.Mailer
	authLimit *ratelimit.Limiter
	apiLimit  *ratelimit.Limiter
	static    http.Handler
}

func New(d Deps) *Server {
	return &Server{
		cfg:       d.Config,
		db:        d.DB,
		mailer:    d.Mailer,
		authLimit: ratelimit.New(d.Config.AuthRateLimit, time.Minute, d.Config.TrustedProxies),
		apiLimit:  ratelimit.New(d.Config.APIRateLimit, time.Minute, d.Config.TrustedProxies),
		static:    spaHandler(),
	}
}
