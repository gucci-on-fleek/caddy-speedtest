// Caddy Speedtest
// https://maxchernoff.ca/tools/speedtest
// SPDX-License-Identifier: Apache-2.0+
// SPDX-FileCopyrightText: 2025 Max Chernoff

//////////////////////
/// Initialization ///
//////////////////////

package speedtest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddytest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/////////////////
/// Constants ///
/////////////////

const urlAuthority = "http://localhost:8080"

////////////////////////
/// Helper Functions ///
////////////////////////

// “testSetup” initializes a Caddy test server with the “speedtest” module
// enabled.
func testSetup(t testing.TB) *caddytest.Tester {
	// Silence the caddytest logs
	zap.RedirectStdLogAt(caddy.Log(), zapcore.DebugLevel)

	tester := caddytest.NewTester(t)
	tester.InitServer(
		fmt.Sprintf(
			`{
				skip_install_trust
				admin localhost:%d

				log {
					level ERROR
					format console
				}
			}

			%s {
				respond / "404 Not Found" 404
				speedtest /speedtest
				handle_errors {
					header Content-Type "text/plain; charset=utf-8"
					respond "{err.status_code} {err.status_text}: {err.message}"
				}
			}`,
			caddytest.Default.AdminPort, urlAuthority,
		),
		"caddyfile",
	)
	return tester
}

// “getRandBytes” returns a byte slice of the specified size filled with
// random data.
func getRandBytes(t testing.TB, size int64) []byte {
	rand := newRandReadSeeker(size)
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatalf("failed to read from rand: %v", err)
	}
	return data
}

/////////////////////////
/// Testing Functions ///
/////////////////////////

// “TestDownloadOk” tests valid download requests.
func TestDownloadOk(t *testing.T) {
	tester := testSetup(t)

	for _, tc := range []struct {
		sizeStr string
		sizeInt int64
	}{
		{"1", 1},
		{"1000", 1000},
		{"1k", 1000},
		{"1kB", 1000},
		{"1024B", 1024},
		{"1kiB", 1024},
		{"1MB", 1000 * 1000},
		{"0.1Gi", 1024 * 1024 * 1024 / 10},
	} {
		t.Run("Normal "+tc.sizeStr, func(t *testing.T) {
			t.Parallel()

			expectedBody := getRandBytes(t, tc.sizeInt)

			_, measuredBody := tester.AssertGetResponse(
				fmt.Sprint(urlAuthority, "/speedtest?bytes=", tc.sizeStr),
				200,
				string(expectedBody),
			)

			if len(measuredBody) != int(tc.sizeInt) {
				t.Fatalf("expected body size %d, got %d", tc.sizeInt, len(measuredBody))
			}
		})
	}

	t.Run("Range at beginning", func(t *testing.T) {
		t.Parallel()

		url, err := url.Parse(fmt.Sprint(urlAuthority, "/speedtest?bytes=1EB"))
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		tester.AssertResponse(
			&http.Request{
				Method: "GET",
				URL:    url,
				Header: http.Header{
					"Range": []string{"bytes=0-1000"},
				},
			},
			206,
			string(getRandBytes(t, 1001)),
		)
	})
}

// “TestUploadOk” tests valid upload requests.
func TestUploadOk(t *testing.T) {
	tester := testSetup(t)

	for _, tc := range []struct {
		sizeStr string
		sizeInt int64
	}{
		{"1 B", 1},
		{"1.0 kB", 1000},
		{"1.0 kB", 1024},
		{"1.0 MB", 1000 * 1000},
		{"100 MB", 100 * 1000 * 1000},
	} {
		t.Run(tc.sizeStr, func(t *testing.T) {
			t.Parallel()

			postBody := getRandBytes(t, tc.sizeInt)
			if int64(len(postBody)) != tc.sizeInt {
				t.Fatalf("expected post body size %d, got %d", tc.sizeInt, len(postBody))
			}

			_, _ = tester.AssertPostResponseBody(
				fmt.Sprint(urlAuthority, "/speedtest"),
				[]string{
					"Content-Type: application/octet-stream",
				},
				bytes.NewBuffer(postBody),
				200,
				fmt.Sprintf("Received %s.\n", tc.sizeStr),
			)
		})
	}
}

// “TestDownloadBadRequest” tests invalid download requests.
func TestDownloadBad(t *testing.T) {
	tester := testSetup(t)

	for _, tc := range []struct {
		name         string
		url          string
		statusCode   int
		responseBody string
	}{
		{
			"unhandled page",
			fmt.Sprint(urlAuthority, "/"),
			404,
			"404 Not Found",
		},
		{
			"zero bytes",
			fmt.Sprint(urlAuthority, "/speedtest?bytes=0"),
			400,
			`400 Bad Request: invalid or missing "bytes" query parameter`,
		},
		{
			"invalid bytes",
			fmt.Sprint(urlAuthority, "/speedtest?bytes=invalid"),
			400,
			`400 Bad Request: invalid or missing "bytes" query parameter`,
		},
		{
			"too large",
			fmt.Sprintf("%s/speedtest?bytes=%d", urlAuthority, uint64(1<<63)),
			500,
			"negative content size computed\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tester.AssertGetResponse(
				tc.url,
				tc.statusCode,
				tc.responseBody,
			)
		})
	}

	t.Run("Range at end", func(t *testing.T) {
		t.Parallel()

		url, err := url.Parse(fmt.Sprint(urlAuthority, "/speedtest?bytes=1EB"))
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		tester.AssertResponse(
			&http.Request{
				Method: "GET",
				URL:    url,
				Header: http.Header{
					"Range": []string{"bytes=1-1000"},
				},
			},
			416,
			"seeking not supported\n",
		)
	})
}

// “TestUploadBadRequest” tests invalid upload requests.
func TestUploadBad(t *testing.T) {
	tester := testSetup(t)

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		_, _ = tester.AssertPostResponseBody(
			fmt.Sprint(urlAuthority, "/speedtest"),
			[]string{
				"Content-Type: application/octet-stream",
			},
			bytes.NewBuffer([]byte{}),
			400,
			"400 Bad Request: request body is empty",
		)
	})

	t.Run("unhandled page", func(t *testing.T) {
		t.Parallel()

		_, _ = tester.AssertPostResponseBody(
			fmt.Sprint(urlAuthority, "/"),
			[]string{
				"Content-Type: application/octet-stream",
			},
			bytes.NewBuffer([]byte("test")),
			404,
			"404 Not Found",
		)
	})
}

// “BenchmarkDownload“ benchmarks download performance.
func BenchmarkDownload(b *testing.B) {
	tester := testSetup(b)

	// Warm up the server
	for range 3 {
		resp, err := tester.Client.Get(
			fmt.Sprint(urlAuthority, "/speedtest?bytes=1MB"),
		)
		if err != nil || resp.StatusCode != 200 {
			b.Fatalf("failed to warm up server: %v", err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	// Benchmark different download sizes
	for _, bc := range []struct {
		sizeStr string
		sizeInt int64
	}{
		{"1MB", 1 * 1000 * 1000},
		{"10MB", 10 * 1000 * 1000},
		{"100MB", 100 * 1000 * 1000},
		{"1GB", 1 * 1000 * 1000 * 1000},
	} {
		b.Run(bc.sizeStr, func(b *testing.B) {
			b.SetBytes(bc.sizeInt)
			for b.Loop() {
				resp, err := tester.Client.Get(
					fmt.Sprint(urlAuthority, "/speedtest?bytes=", bc.sizeStr),
				)
				if err != nil || resp.StatusCode != 200 {
					b.Fatalf("failed to download data: %v", err)
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			b.ReportMetric(0, "ns/op") // Discard the ns/op metric
		})
	}
}

// “BenchmarkUpload“ benchmarks upload performance.
func BenchmarkUpload(b *testing.B) {
	tester := testSetup(b)

	// Warm up the server
	for range 3 {
		postBody := bytes.NewBuffer(getRandBytes(b, 1*1000*1000))
		resp, err := tester.Client.Post(
			fmt.Sprint(urlAuthority, "/speedtest"),
			"application/octet-stream",
			postBody,
		)
		if err != nil || resp.StatusCode != 200 {
			b.Fatalf("failed to warm up server: %v", err)
		}
		resp.Body.Close()
	}

	// Benchmark different upload sizes
	for _, bc := range []struct {
		sizeStr string
		sizeInt int64
	}{
		{"1MB", 1 * 1000 * 1000},
		{"10MB", 10 * 1000 * 1000},
		{"100MB", 100 * 1000 * 1000},
		{"1GB", 1 * 1000 * 1000 * 1000},
	} {
		b.Run(bc.sizeStr, func(b *testing.B) {
			var postBody bytes.Buffer
			randBytes := getRandBytes(b, bc.sizeInt)
			postBody.Grow(int(bc.sizeInt))

			b.SetBytes(bc.sizeInt)

			for b.Loop() {
				postBody.Reset()
				postBody.Write(randBytes)
				resp, err := tester.Client.Post(
					fmt.Sprint(urlAuthority, "/speedtest"),
					"application/octet-stream",
					&postBody,
				)
				if err != nil || resp.StatusCode != 200 {
					b.Fatalf("failed to upload data: %v", err)
				}
				resp.Body.Close()
			}
			b.ReportMetric(0, "ns/op") // Discard the ns/op metric
		})
	}
}
