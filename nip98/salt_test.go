package nip98

import (
	"testing"

	"github.com/ohstr/nmilat/nip01"
)

// Verify never reads the anti-replay tag: the replay cache keys on the event
// id. So a header tagged ["nonce", <hex>] (what Sign wrote before
// v0.5.0-rc.47, and what some clients still write), one tagged
// ["salt", <hex>] (Sign now) and one with neither (a signer that adds no
// tag at all) verify alike, strictly, and once each.
func TestVerifyStrict_SaltOrNonceOrNeither(t *testing.T) {
	retag := func(name string) func(*nip01.Event) {
		return func(ev *nip01.Event) {
			kept := ev.Tags[:0]
			for _, tag := range ev.Tags {
				if tag[0] != "salt" {
					kept = append(kept, tag)
				}
			}
			ev.Tags = kept
			if name != "" {
				ev.Tags = append(ev.Tags, []string{name, "e7c3f33340c04d1a0ae8afd23be0eba0"})
			}
		}
	}
	for _, name := range []string{"salt", "nonce", ""} {
		seen := NewSeenIDs(0)
		h := signed(t, testPrivKey, retag(name), true)
		if _, err := VerifyAnyPubkey(backendRequest("POST", h), strictOpts(seen)); err != nil {
			t.Fatalf("tag %q: %v", name, err)
		}
		if _, err := VerifyAnyPubkey(backendRequest("POST", h), strictOpts(seen)); err != ErrReplayed {
			t.Fatalf("tag %q, again: %v, want ErrReplayed", name, err)
		}
	}
}
