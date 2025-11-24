// Caddy Speedtest
// https://maxchernoff.ca/tools/speedtest
// SPDX-License-Identifier: Apache-2.0+
// SPDX-FileCopyrightText: 2025 Max Chernoff

//////////////////////
/// Initialization ///
//////////////////////

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

// Register the module and Caddyfile directive.
func init() {
	caddy.RegisterModule(Speedtest{})
	httpcaddyfile.RegisterHandlerDirective("speedtest", parseCaddyfile)
	httpcaddyfile.RegisterDirectiveOrder(
		"speedtest", httpcaddyfile.Before, "file_server",
	)
}

// [Speedtest] implements an HTTP handler that performs speed tests.
type Speedtest struct{}

// “CaddyModule” returns the Caddy module information.
func (Speedtest) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.speedtest",
		New: func() caddy.Module { return new(Speedtest) },
	}
}

/////////////////////////
/// Caddyfile Parsing ///
/////////////////////////

// “parseCaddyfile” initializes the [Speedtest] module from Caddyfile tokens.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Speedtest
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// “UnmarshalCaddyfile” implements [caddyfile.Unmarshaler], and parse the
// Caddyfile tokens for the “speedtest” directive.
func (m *Speedtest) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if !d.Next() { // "speedtest"
		return fmt.Errorf(`"speedtest" directive is missing`) // Impossible
	}
	if d.Next() { // Any arguments?
		return fmt.Errorf(`"speedtest" takes no arguments`)
	}
	return nil
}

// “Provision” implements [caddy.Provisioner]. We don't have any setup to do,
// so this is a no-op.
func (m *Speedtest) Provision(ctx caddy.Context) error {
	return nil
}

// “Validate” implements [caddy.Validator]. We don't have any configuration
// to validate, so this is a no-op.
func (m *Speedtest) Validate() error {
	return nil
}

// Interface guards
var (
	_ caddy.Provisioner     = (*Speedtest)(nil)
	_ caddy.Validator       = (*Speedtest)(nil)
	_ caddyfile.Unmarshaler = (*Speedtest)(nil)
)

//////////////////////////////
/// Random Bytes Generator ///
//////////////////////////////

// “randReadSeeker” implements [io.ReadSeeker]. We use this to generate
// pseudo-random data with a fixed seed such that it can be efficiently served
// via [http.ServeContent].
type randReadSeeker struct {
	rng  *rand.ChaCha8 // Pseudo-random number generator
	size int64         // Size in bytes of the data to be generated
}

// “newRandReadSeeker” creates a new [randReadSeeker] of the given size.
func newRandReadSeeker(size int64) *randReadSeeker {
	return &randReadSeeker{
		rng:  rand.NewChaCha8([32]byte{}),
		size: size,
	}
}

// “Seek” implements [io.Seeker]. We only implement the bare minimum required
// by [http.ServeContent] for non-“Range” requests.
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

// “Read” implements [io.Reader]. This is just a wrapper around the RNG's
// [rand.ChaCha8.Read] method.
func (r *randReadSeeker) Read(p []byte) (n int, err error) {
	return r.rng.Read(p)
}

// Interface guards
var (
	_ io.ReadSeeker = (*randReadSeeker)(nil)
)

////////////////////
/// HTTP Handler ///
////////////////////

// “ServeHTTP” implements [caddyhttp.MiddlewareHandler]. This is the function
// that responds to each HTTP request. Since this is a “responder” handler, the
// “next” handler is ignored and no further processing is done after this
// handler.
func (m Speedtest) ServeHTTP(w http.ResponseWriter, r *http.Request, _ caddyhttp.Handler) error {
	switch r.Method {
	case http.MethodGet:
		return m.handleGet(w, r)
	case http.MethodPost:
		return m.handlePost(w, r)
	default:
		return caddyhttp.Error(
			http.StatusMethodNotAllowed,
			fmt.Errorf("only GET and POST methods are allowed"),
		)
	}
}

// “handleGet” handles “GET” requests for the speedtest by serving pseudo-random
// data of the requested size.
func (m Speedtest) handleGet(w http.ResponseWriter, r *http.Request) error {
	// Parse the "bytes" query parameter.
	bytes, err := humanize.ParseBytes(r.URL.Query().Get("bytes"))
	if err != nil || bytes == 0 {
		return caddyhttp.Error(
			http.StatusBadRequest,
			fmt.Errorf(`invalid or missing "bytes" query parameter`),
		)
	}

	// Serve pseudo-random data of the requested size.
	rng := newRandReadSeeker(int64(bytes))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, "", time.Time{}, rng)

	return nil
}

// “handlePost” handles “POST” requests for the speedtest by reading and
// discarding the request body and reporting the number of bytes received.
func (m Speedtest) handlePost(w http.ResponseWriter, r *http.Request) error {
	// Unconditionally send a `100 Continue` response
	w.WriteHeader(http.StatusContinue)

	// Read and discard the request body
	size, err := io.Copy(io.Discard, r.Body)
	if err != nil {
		return caddyhttp.Error(
			http.StatusInternalServerError,
			fmt.Errorf("failed to read request body: %v", err),
		)
	}

	if size == 0 {
		return caddyhttp.Error(
			http.StatusBadRequest,
			fmt.Errorf("request body is empty"),
		)
	}

	// Respond with the number of bytes received
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "Received", humanize.Bytes(uint64(size)), ".")

	return nil
}

// Interface guards
var (
	_ caddyhttp.MiddlewareHandler = (*Speedtest)(nil)
)
