package main

import (
	"encoding/json"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func extractHostname(rawHost string) string {
	host, _, err := net.SplitHostPort(rawHost)
	if err != nil {
		return rawHost
	}

	return host
}

func readConfig() (*Config, error) {
	var c Config

	data, err := os.ReadFile("redirects.json")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &c)
	return &c, err
}

func hostnameHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := hlog.FromRequest(r)
			log.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str(fieldKey, extractHostname(r.Host))
			})
			next.ServeHTTP(w, r)
		})
	}
}
