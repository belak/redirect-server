package main

import (
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/justinas/alice"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

var logger zerolog.Logger

type RedirectType int

const (
	RedirectTypeTemporary = iota
	RedirectTypePermanent
)

type Config struct {
	Domains map[string]*DomainConfig
}

type DomainConfig struct {
	Routes map[string]*RouteConfig
}

type RouteConfig struct {
	Target string
	Type   RedirectType
}

func main() {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		//logger = zerolog.New(os.Stdout)
		logger = zerolog.New(zerolog.NewConsoleWriter())
	} else {
		logger = zerolog.New(os.Stdout)
	}

	// Attach a timestamp to the logger to be more useful
	logger = logger.With().
		Timestamp().
		Logger()

	config, err := readConfig()
	if err != nil {
		logger.Panic().Err(err).Msg("Failed to read config")
	}

	logger.Info().Msg("Starting up")

	// Create a new middleware chain
	c := alice.New()
	c = c.Append(hlog.NewHandler(logger))
	c = c.Append(hlog.URLHandler("url"))
	c = c.Append(hlog.MethodHandler("method"))
	c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("Handled request")
	}))

	/*
		c = c.Append(hlog.RemoteAddrHandler("ip"))
		c = c.Append(hlog.UserAgentHandler("user_agent"))
		c = c.Append(hlog.RefererHandler("referer"))
	*/

	handleDomain := func(w http.ResponseWriter, r *http.Request, domain *DomainConfig, urlPath string) {
		route := domain.Routes[urlPath]
		if route == nil {
			http.NotFound(w, r)
			return
		}

		hlog.FromRequest(r).Info().
			Str("redirect_target", route.Target).
			Bool("redirect_permanent", route.Type == RedirectTypePermanent).
			Msg("Found redirect")

		status := http.StatusTemporaryRedirect
		if route.Type == RedirectTypePermanent {
			status = http.StatusPermanentRedirect
		}

		http.Redirect(w, r, route.Target, status)
	}

	handler := c.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		urlHost := extractHostname(r.Host)
		urlPath := strings.Trim(path.Clean(r.URL.Path), "/")

		hlog.FromRequest(r).UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("hostname", urlHost).Str("path", urlPath)
		})

		// First check for this domain and try to find an exact match.
		if domain := config.Domains[urlHost]; domain != nil {
			handleDomain(w, r, domain, urlPath)
			return
		}

		// If we didn't have an exact match, use the fallback domain.
		if starDomain := config.Domains["*"]; starDomain != nil {
			handleDomain(w, r, starDomain, urlPath)
			return
		}

		http.NotFound(w, r)
	})

	err = http.ListenAndServe(":8080", handler)

	if err != nil {
		logger.Panic().Err(err).Msg("Failed to bind to address")
	}
}
