// Package nip86 implements NIP-86: the Relay Management API, a JSON-RPC-ish
// POST served on the relay's own URL and authenticated with NIP-98.
//
// This package is the protocol only -- the request and response shapes, the
// method names, and a router that dispatches them. It holds no relay state and
// does not import relay/, so the binding to a particular membership store lives
// in whatever composes the relay's HTTP handler.
//
// A request is selected by its Content-Type, not its path, so the management
// API shares the relay URL with the WebSocket upgrade and the NIP-11 document:
//
//	POST / HTTP/1.1
//	Content-Type: application/nostr+json+rpc
//	Authorization: Nostr <base64 kind-27235 event>
//
//	{"method":"allowpubkey","params":["<hex pubkey>"]}
//
// NIP-86 requires the NIP-98 event to carry a payload tag hashing the body;
// nip98.Verify checks that when given Options.Body.
package nip86

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
)

// ContentType selects the management API on a request to the relay URL.
const ContentType = "application/nostr+json+rpc"

// Method names. Those below are the ones a members-only relay needs; NIP-86
// defines further event, kind and IP methods a relay may also answer.
const (
	MethodSupportedMethods   = "supportedmethods"
	MethodAllowPubkey        = "allowpubkey"
	MethodUnallowPubkey      = "unallowpubkey"
	MethodListAllowedPubkeys = "listallowedpubkeys"
	MethodBanPubkey          = "banpubkey"
	MethodUnbanPubkey        = "unbanpubkey"
	MethodListBannedPubkeys  = "listbannedpubkeys"
	MethodCreateClaim        = "createclaim"
	MethodListClaims         = "listclaims"
	MethodDeleteClaim        = "deleteclaim"
	MethodCreateRole         = "createrole"
	MethodEditRole           = "editrole"
	MethodDeleteRole         = "deleterole"
	MethodAssignRole         = "assignrole"
	MethodUnassignRole       = "unassignrole"
)

var (
	ErrNotManagementRequest = errors.New("nip86: not a relay management request")
	ErrMalformedRequest     = errors.New("nip86: malformed request body")
	ErrMissingMethod        = errors.New("nip86: request names no method")
	ErrUnknownMethod        = errors.New("nip86: unknown method")
	ErrMissingParam         = errors.New("nip86: missing parameter")
	ErrWrongParamType       = errors.New("nip86: parameter has the wrong type")
)

// Request is one management call. Params stay raw because they are
// heterogeneous across methods -- a pubkey string here, a role object there.
type Request struct {
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

// Response is the reply. Per the spec a failure is still HTTP 200 with error
// set; only an authentication failure is a status code (401).
//
// Result has no omitempty, so an error serializes as {"result":null,"error":
// "<message>"} rather than dropping the field: a client can read result
// unconditionally and need not treat a missing key differently from a null one.
// Error keeps omitempty, so a success is {"result":<value>} with no error key.
type Response struct {
	Result any    `json:"result"`
	Error  string `json:"error,omitempty"`
}

// IsManagementRequest reports whether r should be served as NIP-86 rather than
// upgraded to a WebSocket or answered with NIP-11. Content-Type parameters such
// as a charset are tolerated.
func IsManagementRequest(r *http.Request) bool {
	if r.Method != http.MethodPost {
		return false
	}
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		// A bare, unparameterized header is still usable.
		return ct == ContentType
	}
	return mediaType == ContentType
}

// ParseRequest decodes a management request body.
func ParseRequest(body []byte) (Request, error) {
	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		return Request{}, fmt.Errorf("%w: %v", ErrMalformedRequest, err)
	}
	if req.Method == "" {
		return Request{}, ErrMissingMethod
	}
	return req, nil
}

// StringParam returns param i as a string.
func (r Request) StringParam(i int) (string, error) {
	if i >= len(r.Params) {
		return "", fmt.Errorf("%w: param %d", ErrMissingParam, i)
	}
	var s string
	if err := json.Unmarshal(r.Params[i], &s); err != nil {
		return "", fmt.Errorf("%w: param %d is not a string", ErrWrongParamType, i)
	}
	return s, nil
}

// IntParam returns param i as an int.
func (r Request) IntParam(i int) (int, error) {
	if i >= len(r.Params) {
		return 0, fmt.Errorf("%w: param %d", ErrMissingParam, i)
	}
	var n int
	if err := json.Unmarshal(r.Params[i], &n); err != nil {
		return 0, fmt.Errorf("%w: param %d is not a number", ErrWrongParamType, i)
	}
	return n, nil
}

// ObjectParam decodes param i into v.
func (r Request) ObjectParam(i int, v any) error {
	if i >= len(r.Params) {
		return fmt.Errorf("%w: param %d", ErrMissingParam, i)
	}
	if err := json.Unmarshal(r.Params[i], v); err != nil {
		return fmt.Errorf("%w: param %d: %v", ErrWrongParamType, i, err)
	}
	return nil
}
