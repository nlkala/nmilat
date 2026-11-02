package nip86

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ohstr/nmilat/nip01"
)

const (
	adminPrivKey = "0acd12cbf0fb87cd13b17bc9b57dffd11b3870b407984cec5a4ce2a69b90268c"
	otherPrivKey = "0000000000000000000000000000000000000000000000000000000000000001"
)

func pubkeyOf(t *testing.T, privKey string) string {
	t.Helper()
	ev := &nip01.Event{}
	if err := ev.Sign(privKey); err != nil {
		t.Fatalf("sign: %v", err)
	}
	return ev.PubKey
}

// authHeader builds a NIP-98 Authorization header over body, as NIP-86 requires.
func authHeader(t *testing.T, privKey, url, method string, body []byte) string {
	t.Helper()
	sum := sha256.Sum256(body)
	ev := &nip01.Event{
		Kind:      27235,
		CreatedAt: uint64(time.Now().Unix()),
		Tags: [][]string{
			{"u", url},
			{"method", method},
			{"payload", hex.EncodeToString(sum[:])},
		},
	}
	if err := ev.Sign(privKey); err != nil {
		t.Fatalf("sign: %v", err)
	}
	raw, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return "Nostr " + base64.StdEncoding.EncodeToString(raw)
}

func testHandler(t *testing.T, admins []string, origins []string) *Handler {
	t.Helper()
	router := NewRouter()
	router.Handle(MethodListAllowedPubkeys, func(context.Context, string, Request) (any, error) {
		return []string{"alice"}, nil
	})
	return NewHandler(Config{Router: router, AllowedPubkeys: admins, AllowedOrigins: origins})
}

// tamperedSigAuthHeader builds a NIP-98 header whose event is well-formed but
// whose signature does not verify, so Verify rejects it at the signature step.
func tamperedSigAuthHeader(t *testing.T, privKey, url, method string, body []byte) string {
	t.Helper()
	sum := sha256.Sum256(body)
	ev := &nip01.Event{
		Kind:      27235,
		CreatedAt: uint64(time.Now().Unix()),
		Tags: [][]string{
			{"u", url},
			{"method", method},
			{"payload", hex.EncodeToString(sum[:])},
		},
	}
	if err := ev.Sign(privKey); err != nil {
		t.Fatalf("sign: %v", err)
	}
	// Flip the first hex digit of the signature: still 128 hex chars, so it
	// parses, but it no longer verifies against the event.
	if ev.Sig[0] == '0' {
		ev.Sig = "1" + ev.Sig[1:]
	} else {
		ev.Sig = "0" + ev.Sig[1:]
	}
	raw, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return "Nostr " + base64.StdEncoding.EncodeToString(raw)
}

func postRequest(t *testing.T, body []byte) *http.Request {
	t.Helper()
	req := httptest.NewRequest("POST", "http://example.com/", bytes.NewReader(body))
	req.Header.Set("Content-Type", ContentType)
	return req
}

// Without this a browser never sends the POST at all: neither the content type
// nor the Authorization header is CORS-safelisted.
func TestHandler_Preflight(t *testing.T) {
	h := testHandler(t, []string{pubkeyOf(t, adminPrivKey)}, nil)

	req := httptest.NewRequest("OPTIONS", "http://example.com/", nil)
	req.Header.Set("Origin", "https://app.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Allow-Origin = %q, want *", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "POST, OPTIONS" {
		t.Fatalf("Allow-Methods = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Authorization, Content-Type" {
		t.Fatalf("Allow-Headers = %q, must cover Authorization and Content-Type", got)
	}
}

func TestHandler_Dispatches(t *testing.T) {
	admin := pubkeyOf(t, adminPrivKey)
	h := testHandler(t, []string{admin}, nil)

	body := []byte(`{"method":"listallowedpubkeys","params":[]}`)
	req := postRequest(t, body)
	req.Header.Set("Authorization", authHeader(t, adminPrivKey, "http://example.com/", "POST", body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		Result []string `json:"result"`
		Error  string   `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error != "" {
		t.Fatalf("error = %q", resp.Error)
	}
	if len(resp.Result) != 1 || resp.Result[0] != "alice" {
		t.Fatalf("result = %v", resp.Result)
	}
}

// A client that signs the bare relay URL must still reach a handler at "/".
func TestHandler_AcceptsURLWithoutTrailingSlash(t *testing.T) {
	admin := pubkeyOf(t, adminPrivKey)
	h := testHandler(t, []string{admin}, nil)

	body := []byte(`{"method":"listallowedpubkeys","params":[]}`)
	req := postRequest(t, body)
	req.Header.Set("Authorization", authHeader(t, adminPrivKey, "http://example.com", "POST", body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestHandler_Unauthorized(t *testing.T) {
	admin := pubkeyOf(t, adminPrivKey)
	body := []byte(`{"method":"listallowedpubkeys","params":[]}`)

	tests := []struct {
		name   string
		admins []string
		mutate func(t *testing.T, req *http.Request)
	}{
		{
			name:   "no authorization header",
			admins: []string{admin},
			mutate: func(*testing.T, *http.Request) {},
		},
		{
			name:   "not an admin",
			admins: []string{admin},
			mutate: func(t *testing.T, req *http.Request) {
				req.Header.Set("Authorization", authHeader(t, otherPrivKey, "http://example.com/", "POST", body))
			},
		},
		{
			name:   "no admins configured fails closed",
			admins: nil,
			mutate: func(t *testing.T, req *http.Request) {
				req.Header.Set("Authorization", authHeader(t, adminPrivKey, "http://example.com/", "POST", body))
			},
		},
		{
			name:   "payload tag missing",
			admins: []string{admin},
			mutate: func(t *testing.T, req *http.Request) {
				ev := &nip01.Event{
					Kind:      27235,
					CreatedAt: uint64(time.Now().Unix()),
					Tags:      [][]string{{"u", "http://example.com/"}, {"method", "POST"}},
				}
				if err := ev.Sign(adminPrivKey); err != nil {
					t.Fatalf("sign: %v", err)
				}
				raw, _ := json.Marshal(ev)
				req.Header.Set("Authorization", "Nostr "+base64.StdEncoding.EncodeToString(raw))
			},
		},
		{
			name:   "header signed for a different body",
			admins: []string{admin},
			mutate: func(t *testing.T, req *http.Request) {
				other := []byte(`{"method":"allowpubkey","params":["attacker"]}`)
				req.Header.Set("Authorization", authHeader(t, adminPrivKey, "http://example.com/", "POST", other))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := testHandler(t, tc.admins, nil)
			req := postRequest(t, body)
			tc.mutate(t, req)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401 (body: %s)", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandler_UnknownMethodIsA200WithError(t *testing.T) {
	admin := pubkeyOf(t, adminPrivKey)
	h := testHandler(t, []string{admin}, nil)

	body := []byte(`{"method":"nosuchmethod","params":[]}`)
	req := postRequest(t, body)
	req.Header.Set("Authorization", authHeader(t, adminPrivKey, "http://example.com/", "POST", body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error == "" {
		t.Fatal("want an error field")
	}
}

// Authorize widens which verified keys reach the router without touching the
// signature/payload/URL checks nip98.Verify runs for everyone.
func TestHandler_Authorize(t *testing.T) {
	admin := pubkeyOf(t, adminPrivKey)
	member := pubkeyOf(t, otherPrivKey)
	const url = "http://example.com/"
	body := []byte(`{"method":"listallowedpubkeys","params":[]}`)

	// newHandler builds a handler whose only admin is adminPrivKey and whose
	// Authorize hook is authorize (nil leaves behaviour exactly as today).
	newHandler := func(authorize func(string) bool) *Handler {
		router := NewRouter()
		router.Handle(MethodListAllowedPubkeys, func(context.Context, string, Request) (any, error) {
			return []string{"alice"}, nil
		})
		return NewHandler(Config{Router: router, AllowedPubkeys: []string{admin}, Authorize: authorize})
	}

	serve := func(h *Handler, auth string) *httptest.ResponseRecorder {
		req := postRequest(t, body)
		req.Header.Set("Authorization", auth)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	t.Run("nil Authorize leaves a non-listed key a 401", func(t *testing.T) {
		rec := serve(newHandler(nil), authHeader(t, otherPrivKey, url, "POST", body))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body: %s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("Authorize true lets a non-listed verified key reach the router", func(t *testing.T) {
		var asked string
		h := newHandler(func(pk string) bool { asked = pk; return true })
		rec := serve(h, authHeader(t, otherPrivKey, url, "POST", body))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
		}
		if asked != member {
			t.Fatalf("Authorize was asked about %q, want the member key %q", asked, member)
		}
		var resp Response
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Error != "" {
			t.Fatalf("reached the router but got error %q", resp.Error)
		}
	})

	t.Run("Authorize false keeps a non-listed key a 401", func(t *testing.T) {
		rec := serve(newHandler(func(string) bool { return false }), authHeader(t, otherPrivKey, url, "POST", body))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body: %s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("an admin still passes even when Authorize would refuse", func(t *testing.T) {
		rec := serve(newHandler(func(string) bool { return false }), authHeader(t, adminPrivKey, url, "POST", body))
		if rec.Code != http.StatusOK {
			t.Fatalf("admin status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("a bad signature is still 401 even if Authorize would admit the key", func(t *testing.T) {
		called := false
		h := newHandler(func(string) bool { called = true; return true })
		rec := serve(h, tamperedSigAuthHeader(t, otherPrivKey, url, "POST", body))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body: %s)", rec.Code, rec.Body.String())
		}
		if called {
			t.Fatal("Authorize was consulted despite an invalid signature; Verify must run first")
		}
	})
}

// An error reply must serialize result as an explicit null, so a client can
// read it unconditionally (change E).
func TestHandler_ErrorReplyCarriesNullResult(t *testing.T) {
	admin := pubkeyOf(t, adminPrivKey)
	router := NewRouter()
	router.Handle(MethodDeleteClaim, func(context.Context, string, Request) (any, error) {
		return nil, errors.New("no such claim")
	})
	h := NewHandler(Config{Router: router, AllowedPubkeys: []string{admin}})

	body := []byte(`{"method":"deleteclaim","params":["missing"]}`)
	req := postRequest(t, body)
	req.Header.Set("Authorization", authHeader(t, adminPrivKey, "http://example.com/", "POST", body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	got := strings.TrimSpace(rec.Body.String())
	if want := `{"result":null,"error":"no such claim"}`; got != want {
		t.Fatalf("error reply = %s, want %s", got, want)
	}
}

func TestHandler_RejectsGet(t *testing.T) {
	h := testHandler(t, []string{pubkeyOf(t, adminPrivKey)}, nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest("GET", "http://example.com/", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestHandler_OriginAllowlist(t *testing.T) {
	h := testHandler(t, []string{pubkeyOf(t, adminPrivKey)}, []string{"https://app.example"})

	matching := httptest.NewRequest("OPTIONS", "http://example.com/", nil)
	matching.Header.Set("Origin", "https://app.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, matching)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example" {
		t.Fatalf("Allow-Origin = %q, want the echoed origin", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("Vary = %q, want Origin when echoing", got)
	}

	foreign := httptest.NewRequest("OPTIONS", "http://example.com/", nil)
	foreign.Header.Set("Origin", "https://evil.example")
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, foreign)
	if got := rec2.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Allow-Origin = %q, want none for a foreign origin", got)
	}
}
