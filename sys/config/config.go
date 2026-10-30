// Package config reads the configuration of the app from the environment.
package config

import (
	"crypto/rand"
	"os"
)

type Configuration struct {
	Version   string
	ListenURL string
	// Secret encrypts cookies and CSRF tokens. If OWLDEMO_SECRET is not set, a
	// random one is made at startup (the cookies of a previous run stop working).
	Secret []byte
}

// CreateAndLoad reads OWLDEMO_ADDR (default :8080) and OWLDEMO_SECRET.
func CreateAndLoad() Configuration {
	conf := Configuration{Version: "0.1.0", ListenURL: ":8080"}
	if addr := os.Getenv("OWLDEMO_ADDR"); addr != "" {
		conf.ListenURL = addr
	}
	if secret := os.Getenv("OWLDEMO_SECRET"); secret != "" {
		conf.Secret = []byte(secret)
	}
	if len(conf.Secret) != 32 {
		conf.Secret = make([]byte, 32)
		if _, err := rand.Read(conf.Secret); err != nil {
			panic(err)
		}
	}
	return conf
}
