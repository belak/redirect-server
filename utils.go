package main

import (
	"encoding/json"
	"net"
	"os"
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
	return &c, err
}
