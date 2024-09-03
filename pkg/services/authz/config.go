package authz

import (
	"fmt"

	"github.com/grafana/grafana/pkg/setting"
)

type Mode string

func (s Mode) IsValid() bool {
	switch s {
	case ModeGRPC, ModeInProc:
		return true
	}
	return false
}

const (
	ModeGRPC   Mode = "grpc"
	ModeInProc Mode = "inproc"
)

type Cfg struct {
	remoteAddress string
	listen        bool
	mode          Mode

	token            string
	tokenExchangeURL string
	tokenNamespace   string

	allowInsecure bool
}

func ReadCfg(cfg *setting.Cfg) (*Cfg, error) {
	authorizationSection := cfg.SectionWithEnvOverrides("authorization")
	grpcClientAuthSection := cfg.SectionWithEnvOverrides("grpc_client_authentication")

	mode := Mode(authorizationSection.Key("mode").MustString(string(ModeInProc)))
	if !mode.IsValid() {
		return nil, fmt.Errorf("authorization: invalid mode %q", mode)
	}

	token := grpcClientAuthSection.Key("token").MustString("")
	tokenExchangeURL := grpcClientAuthSection.Key("token_exchange_url").MustString("")
	tokenNamespace := grpcClientAuthSection.Key("token_namespace").MustString("stack-" + cfg.StackID)

	// When running in cloud mode, the token and tokenExchangeURL are required.
	if cfg.StackID != "" && token == "" && tokenExchangeURL == "" {
		return nil, fmt.Errorf("authorization: missing token or tokenExchangeUrl")
	}

	return &Cfg{
		remoteAddress:    authorizationSection.Key("remote_address").MustString(""),
		listen:           authorizationSection.Key("listen").MustBool(false),
		mode:             mode,
		token:            token,
		tokenExchangeURL: tokenExchangeURL,
		tokenNamespace:   tokenNamespace,
		allowInsecure:    cfg.Env == setting.Dev,
	}, nil
}