package web

import (
	"crypto/tls"
	"net/http"

	"github.com/caddyserver/certmagic"
	"github.com/hashicorp/go-hclog"
)

type TLS struct {
	cfg *certmagic.Config
}

func NewTLS(L hclog.Logger, path, email string, test bool) (*TLS, error) {
	certmagic.DefaultACME.Agreed = true
	certmagic.DefaultACME.DisableHTTPChallenge = true

	if test {
		certmagic.DefaultACME.CA = certmagic.DefaultACME.TestCA
	}

	certmagic.DefaultACME.Email = email

	cfg := certmagic.NewDefault()
	cfg.Storage = &certmagic.FileStorage{Path: path}
	cfg.OnDemand = &certmagic.OnDemandConfig{
		DecisionFunc: func(name string) error {
			L.Info("tls decision calculation", "name", name)
			return nil
		},
	}

	return &TLS{
		cfg: cfg,
	}, nil
}

func (t *TLS) ListenAndServe(addr string, h http.Handler) error {
	listener, err := tls.Listen("tcp", addr, t.cfg.TLSConfig())
	if err != nil {
		return err
	}

	return http.Serve(listener, h)
}
