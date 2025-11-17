// Caddy Speedtest
// https://github.com/gucci-on-fleek/caddy-speedtest
// SPDX-License-Identifier: Apache-2.0+
// SPDX-FileCopyrightText: 2025 Max Chernoff
package speedtest

import (
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(Speedtest{})
	httpcaddyfile.RegisterHandlerDirective("speedtest", parseCaddyfile)
	httpcaddyfile.RegisterDirectiveOrder("speedtest", httpcaddyfile.Before, "file_server")

}

// Speedtest implements an HTTP handler that performs speed tests.
type Speedtest struct {
}

// CaddyModule returns the Caddy module information.
func (Speedtest) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.speedtest",
		New: func() caddy.Module { return new(Speedtest) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Speedtest) Provision(ctx caddy.Context) error {
	// switch m.Output {
	// case "stdout":
	// 	m.w = os.Stdout
	// case "stderr":
	// 	m.w = os.Stderr
	// default:
	// 	return fmt.Errorf("an output stream is required")
	// }
	return nil
}

// Validate implements caddy.Validator.
func (m *Speedtest) Validate() error {
	// if m.w == nil {
	// 	return fmt.Errorf("no writer")
	// }
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Speedtest) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	_, err := w.Write([]byte("speedtest module works!\n"))
	return err
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *Speedtest) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			// case "strict":
			// 	Speedtest.Strict = true
			// case "mime":
			// 	Speedtest.MIMETypes = d.RemainingArgs()
			// 	if len(Speedtest.MIMETypes) == 0 {
			// 		return d.ArgErr()
			// 	}
			// case "platform":
			// 	if !d.Args(&Speedtest.Platform) {
			// 		return d.ArgErr()
			// 	}
			default:
				return fmt.Errorf("unknown subdirective: %q", d.Val())
			}
		}
	}
	return nil

}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Speedtest
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Speedtest)(nil)
	_ caddy.Validator             = (*Speedtest)(nil)
	_ caddyhttp.MiddlewareHandler = (*Speedtest)(nil)
	_ caddyfile.Unmarshaler       = (*Speedtest)(nil)
)
