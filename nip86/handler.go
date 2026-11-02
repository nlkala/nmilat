package nip86

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/ohstr/nmilat/nip98"
)

// MaxBodyBytes caps a management request body. The methods take pubkeys, codes
// and small role objects, so anything larger is a mistake or an attack.
const MaxBodyBytes = 1 << 20

// Config configures a Handler.
type Config struct {
	// Router holds the methods to serve. Required.
	Router *Router

	// AllowedPubkeys may call the API. Empty refuses everyone: an operator who
	// has configured no admin gets a closed endpoint, not an open one.
	AllowedPubkeys []string

	// Authorize, when set, admits a verified key that AllowedPubkeys does not
	// list. NIP-98 still checks the signature, payload and URL for everyone --
	// this only widens which verified keys are allowed to reach the Router. A
	// key that Authorize admits but that is not an admin is still filtered
	// per-method by Router.Visible downstream, so this grants reachability, not
	// privilege. Nil leaves authorization exactly as AllowedPubkeys alone: only
	// listed keys pass, and an empty list still refuses everyone.
	Authorize func(pubkey string) bool

	// AllowedOrigins is the CORS allowlist. Empty answers any origin, which is
	// safe here because authorization is a signed header rather than a cookie,
	// so a hostile page gains nothing by being allowed to send the request.
	AllowedOrigins []string
}

// Handler serves the NIP-86 management API over HTTP.
type Handler struct {
	cfg Config
}

// NewHandler returns a Handler serving cfg.
func NewHandler(cfg Config) *Handler {
	return &Handler{cfg: cfg}
}

// ServeHTTP answers a management request.
//
// A browser cannot reach this endpoint without the preflight below: the
// application/nostr+json+rpc content type is not CORS-safelisted and neither is
// Authorization, so the browser sends OPTIONS first and refuses to send the POST
// at all unless it is answered.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	allowed, ok := h.allowOrigin(origin)
	if ok {
		w.Header().Set("Access-Control-Allow-Origin", allowed)
		if allowed != "*" {
			w.Header().Add("Vary", "Origin")
		}
	}

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
	if err != nil {
		http.Error(w, "request body too large or unreadable", http.StatusRequestEntityTooLarge)
		return
	}

	caller, err := h.authorize(r, body)
	if err != nil {
		// 401 is the only status code the spec gives this API; every other
		// failure is a 200 carrying an error field.
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	req, err := ParseRequest(body)
	if err != nil {
		writeResponse(w, Response{Error: err.Error()})
		return
	}

	writeResponse(w, h.cfg.Router.Dispatch(r.Context(), caller, req))
}

// authorize runs NIP-98 over the request and decides whether the verified key
// may reach the Router. The signature, payload, URL and freshness checks are
// identical for everyone; only which verified keys are admitted differs.
func (h *Handler) authorize(r *http.Request, body []byte) (string, error) {
	// NIP-86 requires the payload tag, and NormalizeRootPath spares a client
	// that signed the bare relay URL a 401 it could not diagnose.
	opts := nip98.Options{
		AllowedPubkeys:    h.cfg.AllowedPubkeys,
		Body:              body,
		RequirePayload:    true,
		NormalizeRootPath: true,
	}

	// With no Authorize hook, authorization is exactly the AllowedPubkeys gate
	// nip98.Verify already applies -- unchanged, down to the error it returns,
	// and an empty list still refuses everyone.
	if h.cfg.Authorize == nil {
		return nip98.Verify(r, opts)
	}

	// With one, Verify's checks still run for everyone: VerifyAnyPubkey applies
	// the identical signature/payload/URL/freshness checks and only drops the
	// AllowedPubkeys gate, so a bad signature is still a 401 before Authorize is
	// ever consulted. A verified key then reaches the Router if it is an admin
	// OR Authorize admits it; a key that is neither still gets a 401. What a
	// non-admin may actually call is decided downstream by Router.Visible, not
	// here.
	caller, err := nip98.VerifyAnyPubkey(r, opts)
	if err != nil {
		return "", err
	}
	if h.isAdmin(caller) || h.cfg.Authorize(caller) {
		return caller, nil
	}
	return "", nip98.ErrWrongPubkey
}

// isAdmin reports whether pubkey is in the AllowedPubkeys list.
func (h *Handler) isAdmin(pubkey string) bool {
	for _, p := range h.cfg.AllowedPubkeys {
		if p == pubkey {
			return true
		}
	}
	return false
}

// allowOrigin resolves the Access-Control-Allow-Origin value for origin,
// reporting whether one should be sent at all.
func (h *Handler) allowOrigin(origin string) (string, bool) {
	if len(h.cfg.AllowedOrigins) == 0 {
		return "*", true
	}
	for _, candidate := range h.cfg.AllowedOrigins {
		if candidate == "*" {
			return "*", true
		}
		if origin != "" && strings.EqualFold(candidate, origin) {
			return origin, true
		}
	}
	return "", false
}

func writeResponse(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/nostr+json+rpc")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
