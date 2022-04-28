package main

import (
	"encoding/json"
	"net"
	"os"
	"strings"
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

	filename := os.Getenv("REDIRECTS_CONFIG")
	if filename == "" {
		filename = "redirects.json"
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &c)

	ret := Config{
		Domains: make(map[string]*DomainConfig),
	}
	for k, v := range ret.Domains {
		domainConfig := &DomainConfig{
			Routes: make(map[string]*RouteConfig),
		}
		for k2, v2 := range v.Routes {
			domainConfig.Routes[strings.ToLower(k2)] = v2
		}
		ret.Domains[strings.ToLower(k)] = domainConfig
	}

	return &ret, err
}
