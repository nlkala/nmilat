package nip86

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestIsManagementRequest(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		contentType string
		want        bool
	}{
		{"exact", "POST", ContentType, true},
		{"with charset", "POST", ContentType + "; charset=utf-8", true},
		{"get is not", "GET", ContentType, false},
		{"nostr+json is nip11", "POST", "application/nostr+json", false},
		{"plain json is not", "POST", "application/json", false},
		{"absent", "POST", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "http://example.com/", strings.NewReader("{}"))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			if got := IsManagementRequest(req); got != tc.want {
				t.Fatalf("IsManagementRequest = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseRequest(t *testing.T) {
	req, err := ParseRequest([]byte(`{"method":"allowpubkey","params":["abc"]}`))
	if err != nil {
		t.Fatalf("ParseRequest: %v", err)
	}
	if req.Method != MethodAllowPubkey {
		t.Fatalf("method = %q", req.Method)
	}
	pubkey, err := req.StringParam(0)
	if err != nil || pubkey != "abc" {
		t.Fatalf("StringParam = %q, %v", pubkey, err)
	}
}

func TestParseRequest_Errors(t *testing.T) {
	if _, err := ParseRequest([]byte(`not json`)); !errors.Is(err, ErrMalformedRequest) {
		t.Fatalf("err = %v, want ErrMalformedRequest", err)
	}
	if _, err := ParseRequest([]byte(`{"params":[]}`)); !errors.Is(err, ErrMissingMethod) {
		t.Fatalf("err = %v, want ErrMissingMethod", err)
	}
}

func TestRequestParams(t *testing.T) {
	req := Request{Method: "x", Params: rawParams(t, `"s"`, `7`, `{"name":"vip"}`)}

	if _, err := req.StringParam(3); !errors.Is(err, ErrMissingParam) {
		t.Fatalf("err = %v, want ErrMissingParam", err)
	}
	if _, err := req.StringParam(1); !errors.Is(err, ErrWrongParamType) {
		t.Fatalf("err = %v, want ErrWrongParamType", err)
	}
	if n, err := req.IntParam(1); err != nil || n != 7 {
		t.Fatalf("IntParam = %d, %v", n, err)
	}
	var role struct{ Name string }
	if err := req.ObjectParam(2, &role); err != nil || role.Name != "vip" {
		t.Fatalf("ObjectParam = %+v, %v", role, err)
	}
}

func TestRouter_Dispatch(t *testing.T) {
	r := NewRouter()
	r.Handle(MethodAllowPubkey, func(_ context.Context, caller string, req Request) (any, error) {
		pubkey, err := req.StringParam(0)
		if err != nil {
			return nil, err
		}
		return caller + ":" + pubkey, nil
	})

	resp := r.Dispatch(context.Background(), "admin", Request{
		Method: MethodAllowPubkey,
		Params: rawParams(t, `"guest"`),
	})
	if resp.Error != "" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	if resp.Result != "admin:guest" {
		t.Fatalf("result = %v", resp.Result)
	}
}

func TestRouter_UnknownMethod(t *testing.T) {
	r := NewRouter()
	resp := r.Dispatch(context.Background(), "admin", Request{Method: "nosuchmethod"})
	if resp.Error == "" {
		t.Fatal("want an error for an unknown method")
	}
	if !strings.Contains(resp.Error, "nosuchmethod") {
		t.Fatalf("error should name the method, got %q", resp.Error)
	}
}

func TestRouter_MethodErrorBecomesResponseError(t *testing.T) {
	r := NewRouter()
	r.Handle(MethodDeleteClaim, func(context.Context, string, Request) (any, error) {
		return nil, errors.New("no such claim")
	})

	resp := r.Dispatch(context.Background(), "admin", Request{Method: MethodDeleteClaim})
	if resp.Error != "no such claim" {
		t.Fatalf("error = %q", resp.Error)
	}
	if resp.Result != nil {
		t.Fatalf("result should be empty, got %v", resp.Result)
	}
}

func TestRouter_SupportedMethodsIsSortedAndSelfDescribing(t *testing.T) {
	r := NewRouter()
	r.Handle(MethodUnallowPubkey, noopMethod)
	r.Handle(MethodAllowPubkey, noopMethod)

	resp := r.Dispatch(context.Background(), "admin", Request{Method: MethodSupportedMethods})
	want := []string{MethodAllowPubkey, MethodSupportedMethods, MethodUnallowPubkey}
	if !reflect.DeepEqual(resp.Result, want) {
		t.Fatalf("result = %v, want %v", resp.Result, want)
	}
}

// A method hidden from a caller must also be refused to that caller, or the
// list is advisory and the hiding cosmetic.
func TestRouter_VisibleHidesAndRefuses(t *testing.T) {
	r := NewRouter()
	r.Handle(MethodAllowPubkey, noopMethod)
	r.Visible = func(caller, method string) bool {
		return caller == "owner" || method == MethodSupportedMethods
	}

	listed := r.Dispatch(context.Background(), "guest", Request{Method: MethodSupportedMethods})
	if !reflect.DeepEqual(listed.Result, []string{MethodSupportedMethods}) {
		t.Fatalf("guest sees %v", listed.Result)
	}

	refused := r.Dispatch(context.Background(), "guest", Request{Method: MethodAllowPubkey})
	if refused.Error == "" {
		t.Fatal("guest must not reach a hidden method")
	}

	allowed := r.Dispatch(context.Background(), "owner", Request{Method: MethodAllowPubkey})
	if allowed.Error != "" {
		t.Fatalf("owner refused: %s", allowed.Error)
	}
}

// An error reply serializes result as an explicit null while a success still
// carries its result and no error key (change E).
func TestResponse_JSONShape(t *testing.T) {
	tests := []struct {
		name string
		resp Response
		want string
	}{
		{"error carries a null result", Response{Error: "boom"}, `{"result":null,"error":"boom"}`},
		{"success carries its result and no error", Response{Result: true}, `{"result":true}`},
		{"success with a null result is still just result", Response{Result: nil}, `{"result":null}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.resp)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(got) != tc.want {
				t.Fatalf("Response = %s, want %s", got, tc.want)
			}
		})
	}
}

func noopMethod(context.Context, string, Request) (any, error) { return true, nil }

func rawParams(t *testing.T, params ...string) []json.RawMessage {
	t.Helper()
	out := make([]json.RawMessage, 0, len(params))
	for _, p := range params {
		out = append(out, json.RawMessage(p))
	}
	return out
}
