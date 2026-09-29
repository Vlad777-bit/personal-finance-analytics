package config

import "os"

type Config struct {
	GatewayHTTPAddr string
}

func Load() Config {
	addr := os.Getenv("GATEWAY_HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return Config{GatewayHTTPAddr: addr}
}
