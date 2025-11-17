// Caddy Speedtest
// https://github.com/gucci-on-fleek/caddy-speedtest
// SPDX-License-Identifier: Apache-2.0+
// SPDX-FileCopyrightText: 2025 Max Chernoff
package speedtest

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/dustin/go-humanize"
)

func init() {
	caddy.RegisterModule(Speedtest{})
	httpcaddyfile.RegisterHandlerDirective("speedtest", parseCaddyfile)
	httpcaddyfile.RegisterDirectiveOrder(
		"speedtest", httpcaddyfile.Before, "file_server",
	)
}

// Speedtest implements an HTTP handler that performs speed tests.
type Speedtest struct{}

// CaddyModule returns the Caddy module information.
func (Speedtest) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.speedtest",
		New: func() caddy.Module { return new(Speedtest) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Speedtest) Provision(ctx caddy.Context) error {
	return nil
}

// Validate implements caddy.Validator.
func (m *Speedtest) Validate() error {
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Speedtest) ServeHTTP(w http.ResponseWriter, r *http.Request, _ caddyhttp.Handler) error {
	switch r.Method {
	case http.MethodGet:
		return m.handleGet(w, r)
	case http.MethodPost:
		return m.handlePost(w, r)
	default:
		return caddyhttp.Error(http.StatusMethodNotAllowed, nil)
	}
}

// randReadSeeker is an io.ReadSeeker that generates random data.
type randReadSeeker struct {
	rng  *rand.ChaCha8
	size int64
}

var _ io.ReadSeeker = (*randReadSeeker)(nil)

func newRandReadSeeker(size int64) *randReadSeeker {
	rng := rand.NewChaCha8([32]byte{})
	return &randReadSeeker{
		rng:  rng,
		size: size,
	}
}

func (r *randReadSeeker) Seek(offset int64, whence int) (int64, error) {
	if offset != 0 {
		return 0, fmt.Errorf("seeking not supported")
	}
	switch whence {
	case io.SeekStart:
		return 0, nil
	case io.SeekEnd:
		return r.size, nil
	default:
		return 0, fmt.Errorf("seeking not supported")
	}
}

func (r *randReadSeeker) Read(p []byte) (n int, err error) {
	return r.rng.Read(p)
}

// handleGet handles GET requests for the speedtest.
func (m Speedtest) handleGet(w http.ResponseWriter, r *http.Request) error {
	bytes, err := humanize.ParseBytes(r.URL.Query().Get("bytes"))
	if err != nil || bytes == 0 {
		return caddyhttp.Error(
			http.StatusBadRequest,
			fmt.Errorf("invalid or missing 'bytes' query parameter"),
		)
	}

	rng := newRandReadSeeker(int64(bytes))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, "speedtest.dat", time.Time{}, rng)

	return nil
}

// handlePost handles POST requests for the speedtest.
func (m Speedtest) handlePost(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusContinue)

	size, err := io.Copy(io.Discard, r.Body)
	if err != nil {
		return caddyhttp.Error(
			http.StatusInternalServerError,
			fmt.Errorf("failed to read request body: %v", err),
		)
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "Received", humanize.Bytes(uint64(size)), "bytes.")
	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *Speedtest) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if !d.Next() { // "speedtest"
		return fmt.Errorf(`"speedtest" directive is missing`) // Impossible
	}
	if d.Next() { // Any arguments?
		return fmt.Errorf(`"speedtest" takes no arguments`)
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
