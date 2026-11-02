package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/wire"

	"github.com/ohstr/nmilat/internal/cashtest"
	"github.com/ohstr/nmilat/nip01"
	"github.com/ohstr/nmilat/nipCO"
	"github.com/ohstr/nmilat/nipcash"
	"github.com/ohstr/nmilat/nipcash/transport"
)

const (
	wdAddr      = "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4"   // K-ADDR-1
	wdAddrTB    = "tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kxpjzsx"   // K-ADDR-6
	wdAddrRT    = "bcrt1qw508d6qejxtdg4y5r3zarvary0c5xw7kygt080" // K-ADDR-13
	wdAddrV2    = "bc1zw508d6qejxtdg4y5r3zarvaryvaxxpcs"         // K-ADDR-4
	wdAddrMix   = "BC1QW508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4"   // K-ADDR-11
	wdAmount    = uint64(21000000)
	wdPayout    = uint64(19258000)
	wdRedeemFee = uint64(105000)
)

// wdHub is a fake Hub that offers withdrawals and answers each method the way
// a test sets.
type wdHub struct {
	t         *testing.T
	hub       *fakeHub
	relay     *cashtest.Relay
	s         *BatchSession
	bill      Bill
	slicePriv string
	slicePub  string
	cashMode  bool
	secret    string

	mu        sync.Mutex
	offered   bool
	seen      []transport.Item
	quoteErr  *transport.ResultError
	callErr   *transport.ResultError
	badHash   bool
	lookup    func(item transport.Item) *transport.Result
	cancel    func(item transport.Item) *transport.Result
	sentBy    int64
	payout    *nipcash.Payout // the last committed payout
	lastQuote string
	reqID     string
}

func newWDHub(t *testing.T, cashMode bool, opts ...SessionOption) *wdHub {
	t.Helper()
	w := &wdHub{t: t, offered: true, cashMode: cashMode, sentBy: time.Now().Unix() + 3600}
	w.hub = newTestHub(t, w.decide)
	w.relay = cashtest.NewRelay(t)
	w.hub.hub(t, w.relay)
	ctx, cancel := e2eContext(t)
	defer cancel()
	s, err := OpenBatchSession(ctx, w.hub.hubXOnly, []string{w.relay.URL()}, opts...)
	if err != nil {
		t.Fatalf("OpenBatchSession: %v", err)
	}
	w.s = s
	_, target := sessionKeypair(t)
	w.bill = mustBill(t, target)
	w.slicePriv, w.slicePub = sessionKeypair(t)
	if cashMode {
		w.secret = strings.Repeat("5a", 32)
		w.slicePub, _ = nipcash.CashCommitment(w.secret)
	}
	return w
}

func (w *wdHub) cred() nipcash.Credential {
	if w.cashMode {
		return nipcash.BySecret(w.secret)
	}
	return nipcash.BySigning(w.slicePriv)
}

func (w *wdHub) sign(ev *nip01.Event) error { return ev.Sign(w.hub.hubPriv) }

func (w *wdHub) methods() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	var out []string
	for _, it := range w.seen {
		out = append(out, it.Method)
	}
	return out
}

func wdResult(item transport.Item, v any) *transport.Result {
	b, _ := json.Marshal(v)
	return &transport.Result{ID: item.ID, ResultType: item.Method, Result: b}
}

func (w *wdHub) decide(item transport.Item) *transport.Result {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.seen = append(w.seen, item)
	switch item.Method {
	case nipcash.MethodCashStatus:
		idType := "pubkey"
		if w.cashMode {
			idType = "cash"
		}
		res := map[string]any{"recipients": []map[string]any{{
			"identity_type": idType, "identity_value": w.slicePub, "amount_millis": wdAmount, "claimed": false,
			"redeem_fee_millis": wdRedeemFee, "net_redeemable_millis": wdAmount - wdRedeemFee, "min_transfer_millis": 1000, "expires_at": nil,
		}}}
		if w.offered {
			res["withdraw"] = map[string]any{"withdraw_fee_base": 500000, "withdraw_fee_ppm": 2000, "script_types": [][2]int{{0, 20}, {0, 32}, {1, 32}}, "final_depth": 6}
		}
		return wdResult(item, res)
	case nipcash.MethodCashWithdrawQuote:
		if w.quoteErr != nil {
			return &transport.Result{ID: item.ID, Error: w.quoteErr}
		}
		var req nipcash.CashWithdrawQuoteRequest
		_ = json.Unmarshal(item.Params, &req)
		ev, err := nipcash.NewWithdrawQuoteEvent(nipcash.QuoteFields{Wallet: item.Target, IdentityValue: w.slicePub, Address: req.Address,
			AmountMillis: wdAmount, NetworkFeeMillis: 1200000, HubFeeMillis: 542000, PayoutMillis: wdPayout, SentBy: w.sentBy, HoldsUntil: 870413},
			time.Now().Unix(), w.sign)
		if err != nil {
			w.t.Errorf("quote: %v", err)
			return nil
		}
		raw, _ := json.Marshal(ev)
		return wdResult(item, nipcash.CashWithdrawQuoteResult{Quote: string(raw), Address: req.Address, AmountMillis: wdAmount,
			NetworkFeeMillis: 1200000, HubFeeMillis: 542000, FeeMillis: wdAmount - wdPayout, PayoutMillis: wdPayout, SentBy: w.sentBy, HoldsUntil: 870413})
	case nipcash.MethodCashWithdraw:
		if _, isCancel, _ := nipcash.ParseWithdrawCancel(item.Params); isCancel {
			if w.cancel != nil {
				return w.cancel(item)
			}
			return nil
		}
		if w.callErr != nil {
			return &transport.Result{ID: item.ID, Error: w.callErr}
		}
		req, err := nipcash.ParseCashWithdrawRequest(item.Params, nipcash.WithdrawSlice{CashMode: w.cashMode, Commitment: w.slicePub})
		if err != nil {
			w.t.Errorf("withdraw request: %v", err)
			return nil
		}
		q, _ := nipcash.ParseWithdrawQuote(req.Quote)
		rid := req.RequestID
		if w.badHash {
			rid = strings.Repeat("ee", 16)
		}
		rc, _ := nipcash.NewWithdrawReceiptEvent(nipcash.ReceiptFields{Wallet: item.Target, QuoteID: q.ID, RequestID: rid, Address: q.Address,
			PayoutMillis: q.PayoutMillis, SentBy: q.SentBy}, time.Now().Unix(), w.sign)
		rcRaw, _ := json.Marshal(rc)
		p := &nipcash.Payout{State: nipcash.PayoutQueued, Address: q.Address, AmountMillis: q.AmountMillis, FeeMillis: q.FeeMillis(),
			PayoutMillis: q.PayoutMillis, SentBy: q.SentBy, FinalDepth: 6, Receipt: string(rcRaw)}
		w.payout, w.lastQuote, w.reqID = p, req.Quote, req.RequestID
		return wdResult(item, p)
	case nipcash.MethodCashLookup:
		if w.lookup != nil {
			return w.lookup(item)
		}
	}
	return nil
}

// locked runs f under the hub's lock. The hooks, and what decide records, are
// shared with the relays' goroutines, and a reply over a websocket is no
// happens-before edge the race detector can see.
func (w *wdHub) locked(f func()) {
	w.mu.Lock()
	defer w.mu.Unlock()
	f()
}

// lastPayout is a copy of the last committed payout.
func (w *wdHub) lastPayout() nipcash.Payout {
	w.mu.Lock()
	defer w.mu.Unlock()
	return *w.payout
}

// lookupDone answers a lookup done with p.
func lookupDone(item transport.Item, p *nipcash.Payout) *transport.Result {
	b, _ := json.Marshal(p)
	return wdResult(item, map[string]any{"status": "done", "result": json.RawMessage(b)})
}

func (w *wdHub) quote(t *testing.T, addr string) *nipcash.WithdrawQuote {
	t.Helper()
	ctx, cancel := e2eContext(t)
	defer cancel()
	q, err := w.s.quoteWithdraw(ctx, w.bill, "bitcoin", nipcash.CashWithdrawQuoteParams{Address: addr, Credential: w.cred()})
	if err != nil {
		t.Fatalf("quoteWithdraw: %v", err)
	}
	return q
}

// T21-20: no withdraw object → ErrWithdrawNotOffered, no quote sent.
func TestT21_20_NotOffered(t *testing.T) {
	w := newWDHub(t, false)
	w.locked(func() { w.offered = false })
	ctx, cancel := e2eContext(t)
	defer cancel()
	if _, err := w.s.quoteWithdraw(ctx, w.bill, "bitcoin", nipcash.CashWithdrawQuoteParams{Address: wdAddr, Credential: w.cred()}); !errors.Is(err, ErrWithdrawNotOffered) {
		t.Fatalf("err = %v", err)
	}
	for _, m := range w.methods() {
		if m != nipcash.MethodCashStatus {
			t.Errorf("sent %s", m)
		}
	}
	// A quote refused NOT_IMPLEMENTED is ErrWithdrawNotOffered too.
	w.locked(func() { w.offered = true })
	w.locked(func() { w.quoteErr = &transport.ResultError{Code: nipcash.ErrCodeNotImplemented, Message: "off"} })
	if _, err := w.s.quoteWithdraw(ctx, w.bill, "bitcoin", nipcash.CashWithdrawQuoteParams{Address: wdAddr, Credential: w.cred()}); !errors.Is(err, ErrWithdrawNotOffered) {
		t.Fatalf("NOT_IMPLEMENTED: %v", err)
	}
	if _, err := termsOf(&nipcash.CashStatusResult{}); !errors.Is(err, ErrWithdrawNotOffered) {
		t.Error("termsOf")
	}
}

// E-WD-13 and T21-29: local refusals before anything is sent.
func TestWithdraw_LocalRefusals(t *testing.T) {
	w := newWDHub(t, false)
	ctx, cancel := e2eContext(t)
	defer cancel()
	status := &nipcash.CashStatusResult{Withdraw: &nipcash.WithdrawTerms{FeePpm: 1, ScriptTypes: [][2]int{{0, 20}, {0, 32}, {1, 32}}, FinalDepth: 6},
		Recipients: []nipcash.RecipientStatus{{IdentityType: "pubkey", IdentityValue: w.slicePub, AmountMillis: wdAmount}}}
	for addr, want := range map[string]error{wdAddrMix: nipcash.ErrAddressMalformed, wdAddrV2: nipcash.ErrScriptTypeNotPaid, "zz1qw508d6qejxtdg4y5r3zarvary0c5xw7k3c49eg": nipcash.ErrAddressNetwork} {
		if _, err := w.s.quoteWithdraw(ctx, w.bill, "bitcoin", nipcash.CashWithdrawQuoteParams{Address: addr, Credential: w.cred(), Status: status}); !errors.Is(err, want) {
			t.Errorf("%s: %v, want %v", addr, err, want)
		}
	}
	if n := len(w.methods()); n != 0 {
		t.Errorf("%d items sent", n)
	}
	// The coin's other networks pass locally; the Hub's network decides.
	if _, err := w.s.quoteWithdraw(ctx, w.bill, "bitcoin", nipcash.CashWithdrawQuoteParams{Address: wdAddrTB, Credential: w.cred(), Status: status}); err != nil {
		t.Errorf("tb on bitcoin: %v", err)
	}
	// WithWithdrawHRP narrows the check to exactly one HRP.
	narrow := newWDHub(t, false, WithWithdrawHRP("bc"))
	for _, addr := range []string{wdAddrTB, wdAddrRT} {
		if _, err := narrow.s.quoteWithdraw(ctx, narrow.bill, "bitcoin", nipcash.CashWithdrawQuoteParams{Address: addr, Credential: narrow.cred(), Status: status}); !errors.Is(err, nipcash.ErrAddressNetwork) {
			t.Errorf("narrowed %s: %v", addr, err)
		}
	}
	// An unknown chain and no option: no network check.
	if st, err := w.s.classifyWithdrawAddress("zz1qw508d6qejxtdg4y5r3zarvary0c5xw7k3c49eg", "othercoin"); err != nil || st != (nipcash.ScriptType{Version: 0, ProgramLen: 20}) {
		t.Errorf("unknown chain: %v %v", st, err)
	}
}

// E-WD-01, T11-48, T21-32: quote, then withdraw, written ahead, receipt verified.
func TestWithdraw_QuoteThenWithdraw(t *testing.T) {
	for _, cashMode := range []bool{false, true} {
		w := newWDHub(t, cashMode)
		q := w.quote(t, strings.ToUpper(wdAddr))
		if q.Address != wdAddr || q.PayoutMillis != wdPayout || q.FeeMillis() != wdAmount-wdPayout {
			t.Fatalf("quote %+v", q.QuoteFields)
		}
		if cashMode {
			w.mu.Lock()
			quoteItem := w.seen[len(w.seen)-1]
			w.mu.Unlock()
			if quoteItem.HasProof() || strings.Contains(string(quoteItem.Params), w.secret) {
				t.Error("a cash-mode quote carries a proof or the secret")
			}
		}
		j := newTestJournal()
		ctx, cancel := e2eContext(t)
		var publishedBeforeWrite bool
		ctx = WithPublishFilter(ctx, func(*nip01.Event) bool {
			publishedBeforeWrite = len(j.all()) == 0
			return true
		})
		p, err := w.s.withdraw(ctx, w.bill, nipcash.CashWithdrawParams{Quote: q, Credential: w.cred()}, j, "dev")
		cancel()
		if err != nil {
			t.Fatalf("withdraw: %v", err)
		}
		if publishedBeforeWrite {
			t.Error("published before the write-ahead")
		}
		if p.State != nipcash.PayoutQueued || p.PayoutMillis != wdPayout {
			t.Errorf("payout %+v", p)
		}
		calls := j.all()
		if len(calls) != 1 {
			t.Fatalf("%d calls", len(calls))
		}
		pc := calls[0]
		if pc.Method != nipcash.MethodCashWithdraw || pc.Target != w.bill.Target() || pc.Amount != wdAmount || pc.Address != wdAddr ||
			pc.Quote == "" || pc.HoldsUntil != 870413 || !nipCO.ValidNotify(pc.Notify) || pc.Deadline == 0 || pc.HubIdentity != w.hub.hubXOnly {
			t.Errorf("written call %+v", pc)
		}
		if cashMode {
			if len(pc.Secrets) != 1 || pc.Secrets[0].Role != nipcash.SecretRoleRemint || pc.SourceSecret != w.secret {
				t.Errorf("cash-mode secrets %+v", pc.Secrets)
			}
		} else if len(pc.Secrets) != 0 || pc.SourceSecret != "" {
			t.Errorf("pubkey secrets %+v", pc.Secrets)
		}
		if relays := j.accepted[pc.RequestID]; len(relays) == 0 {
			t.Error("AcceptedRecorder not told")
		}
		if got := w.s.payoutWatchRelays(pc.Notify, nil); len(got) == 0 || got[0] != w.relay.URL() {
			t.Errorf("watch relays %v", got)
		}
		// Recover on the done (T11-51): waits with the payout, never closes.
		act, err := Recover(context.Background(), pc, answered(LookupOutcome{Status: nipcash.LookupDone, Result: mustJSON(t, w.lastPayout())}), "dev", time.Now())
		if err != nil || act.Kind != ActionWait || act.Result == nil {
			t.Errorf("Recover done: %+v %v", act, err)
		}
		bad := w.lastPayout()
		bad.PayoutMillis, bad.FeeMillis = wdPayout-1000, wdAmount-wdPayout+1000
		act, _ = Recover(context.Background(), pc, answered(LookupOutcome{Status: nipcash.LookupDone, Result: mustJSON(t, &bad)}), "dev", time.Now())
		if act.Kind != ActionWait || act.Reason != "lookup_malformed" {
			t.Errorf("Recover bad receipt: %+v", act)
		}
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// T11-52, E-WD-12 and the WP-11 code rules: a receipt that does not verify, BAD_REQUEST,
// and an id-bearing refusal with its detail.
func TestWithdraw_Answers(t *testing.T) {
	ctx, cancel := e2eContext(t)
	defer cancel()

	w := newWDHub(t, false)
	q := w.quote(t, wdAddr)
	w.locked(func() { w.badHash = true })
	_, err := w.s.withdraw(ctx, w.bill, nipcash.CashWithdrawParams{Quote: q, Credential: w.cred()}, newTestJournal(), "dev")
	var ind *IndeterminateError
	if !errors.As(err, &ind) || !errors.Is(err, nipcash.ErrReceiptInvalid) {
		t.Errorf("bad receipt: %v", err)
	}

	w = newWDHub(t, false)
	q = w.quote(t, wdAddr)
	w.locked(func() { w.callErr = &transport.ResultError{Code: nipcash.ErrCodeBadRequest, Message: "x"} })
	var fin *FinalError
	if _, err := w.s.withdraw(ctx, w.bill, nipcash.CashWithdrawParams{Quote: q, Credential: w.cred()}, newTestJournal(), "dev"); !errors.As(err, &fin) {
		t.Errorf("BAD_REQUEST: %v", err)
	}
	for _, code := range []string{nipcash.ErrCodeRestricted, nipcash.ErrCodeNotImplemented} {
		w.locked(func() { w.callErr = &transport.ResultError{Code: code, Message: "x", Detail: "quote_expired"} })
		_, err := w.s.withdraw(ctx, w.bill, nipcash.CashWithdrawParams{Quote: q, Credential: w.cred()}, newTestJournal(), "dev")
		if !errors.As(err, &ind) || ind.Code != code || ind.Detail != "quote_expired" || ind.Call.LastDetail != nipcash.DetailQuoteExpired || ind.Call.LastCode != code {
			t.Errorf("%s: %v", code, err)
		}
	}
	w.locked(func() {
		w.callErr = &transport.ResultError{Code: nipcash.ErrCodeRestricted, Message: "x", Detail: "fee_over_cap"}
	})
	if _, err := w.s.withdraw(ctx, w.bill, nipcash.CashWithdrawParams{Quote: q, Credential: w.cred()}, newTestJournal(), "dev"); !errors.As(err, &ind) || ind.Detail != "" {
		t.Errorf("unknown detail kept: %v", err)
	}
	// A quote refusal is definitive, with its detail.
	w.locked(func() {
		w.quoteErr = &transport.ResultError{Code: nipcash.ErrCodeRestricted, Message: "x", Detail: "network"}
	})
	if _, err := w.s.quoteWithdraw(ctx, w.bill, "bitcoin", nipcash.CashWithdrawQuoteParams{Address: wdAddr, Credential: w.cred()}); !errors.As(err, &fin) || fin.Detail != "network" {
		t.Errorf("quote refusal: %v", err)
	}
	// A tampered quote id is refused before anything is written.
	tampered := *q
	tampered.ID = strings.Repeat("0", 64)
	if _, err := w.s.withdraw(ctx, w.bill, nipcash.CashWithdrawParams{Quote: &tampered, Credential: w.cred()}, newTestJournal(), "dev"); err == nil {
		t.Error("a tampered quote was sent")
	}
}

// T21-26: lookups.
func TestT21_26_LookupWithdraw(t *testing.T) {
	ctx, cancel := e2eContext(t)
	defer cancel()
	w := newWDHub(t, true)
	q := w.quote(t, wdAddr)
	j := newTestJournal()
	if _, err := w.s.withdraw(ctx, w.bill, nipcash.CashWithdrawParams{Quote: q, Credential: w.cred()}, j, "dev"); err != nil {
		t.Fatal(err)
	}
	pc := j.all()[0]

	// A bearer lookup goes first without the secret, then with it after UNAUTHORIZED.
	var lookups []nipcash.CashLookupRequest
	broadcast := w.lastPayout()
	vout, depth := uint32(1), uint32(0)
	txHex, txid := payoutTx(t, wdPayout/1000)
	broadcast.State, broadcast.Signed, broadcast.Tx, broadcast.Txid, broadcast.Vout, broadcast.Depth = nipcash.PayoutBroadcast, true, txHex, txid, &vout, &depth
	w.locked(func() {
		w.lookup = func(item transport.Item) *transport.Result {
			var r nipcash.CashLookupRequest
			_ = json.Unmarshal(item.Params, &r)
			lookups = append(lookups, r)
			if r.CashSecret == "" {
				return &transport.Result{ID: item.ID, Error: &transport.ResultError{Code: nipcash.ErrCodeUnauthorized, Message: "x"}}
			}
			return lookupDone(item, &broadcast)
		}
	})
	lo, p, err := w.s.lookupWithdraw(ctx, w.bill, pc, w.cred())
	if err != nil || lo.Status != nipcash.LookupDone || p == nil || p.State != nipcash.PayoutBroadcast || p.TxErr != nil {
		t.Fatalf("lookup: %+v %+v %v", lo, p, err)
	}
	w.locked(func() {
		if len(lookups) != 2 || lookups[0].CashSecret != "" || lookups[1].CashSecret != w.secret || lookups[0].RequestID != pc.RequestID {
			t.Errorf("lookups %+v", lookups)
		}
	})
	// A tx that does not pay the payout: TxErr, the payout and its state kept.
	short, shortID := payoutTx(t, wdPayout/1000-1)
	w.locked(func() { broadcast.Tx, broadcast.Txid = short, shortID })
	_, p, _ = w.s.lookupWithdraw(ctx, w.bill, pc, w.cred())
	if p == nil || p.TxErr == nil || p.State != nipcash.PayoutBroadcast {
		t.Errorf("tx check: %+v", p)
	}
	// A cancelled cash-mode payout: the re-minted bill in the clear, verified.
	tok := remintToken(t, w.hub.hubPriv, w.hub.hubXOnly)
	cancelled := w.lastPayout()
	cancelled.State, cancelled.NewWalletPubkey = nipcash.PayoutCancelled, strings.Repeat("22", 32)
	cancelled.NewWalletToken = nipcash.TokenOrTombstone{Token: tok}
	w.locked(func() { w.lookup = func(item transport.Item) *transport.Result { return lookupDone(item, &cancelled) } })
	pcNoSecret := pc
	pcNoSecret.SourceSecret = ""
	_, p, err = w.s.lookupWithdraw(ctx, w.bill, pcNoSecret, w.cred())
	if err != nil || p == nil || p.NewWalletToken.Token != tok {
		t.Errorf("cancelled: %+v %v", p, err)
	}
	// Another Hub's re-minted bill reads as malformed.
	otherPriv, otherX := sessionKeypair(t)
	foreign := remintToken(t, otherPriv, otherX)
	w.locked(func() { cancelled.NewWalletToken = nipcash.TokenOrTombstone{Token: foreign} })
	if lo, p, _ := w.s.lookupWithdraw(ctx, w.bill, pcNoSecret, w.cred()); !lo.Malformed || p != nil {
		t.Errorf("foreign re-mint: %+v", lo)
	}
	// A recovered call (no quote, no deadline): the receipt is checked against the payout.
	recovered := nipcash.PendingCall{RequestID: pc.RequestID, Method: nipcash.MethodCashWithdraw, Target: pc.Target, HubIdentity: pc.HubIdentity}
	w.locked(func() {
		w.lookup = func(item transport.Item) *transport.Result {
			var r nipcash.CashLookupRequest
			_ = json.Unmarshal(item.Params, &r)
			if r.Deadline == 0 {
				t.Error("a recovered lookup without a deadline")
			}
			return lookupDone(item, w.payout)
		}
	})
	if lo, p, err := w.s.lookupWithdraw(ctx, w.bill, recovered, w.cred()); err != nil || p == nil || lo.Malformed {
		t.Errorf("recovered: %+v %v", lo, err)
	} else if id, _ := nipcash.ReceiptQuoteID(p.Receipt); id != q.ID {
		t.Errorf("ReceiptQuoteID %s", id)
	}
}

func payoutTx(t *testing.T, value uint64) (string, string) {
	t.Helper()
	v, prog, err := nipcash.DecodeWitnessAddress(wdAddr, "")
	if err != nil {
		t.Fatal(err)
	}
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(&chainhash.Hash{7}, 0), nil, nil))
	tx.AddTxOut(wire.NewTxOut(1000, nipcash.WitnessScript(0, bytes.Repeat([]byte{3}, 20))))
	tx.AddTxOut(wire.NewTxOut(int64(value), nipcash.WitnessScript(v, prog)))
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(buf.Bytes()), tx.TxHash().String()
}

func remintToken(t *testing.T, minterPriv, minterXOnly string) string {
	t.Helper()
	amount := wdAmount
	sign, err := nipcash.LocalMintSigner(minterPriv, "satscash")
	if err != nil {
		t.Fatal(err)
	}
	tok, err := nipcash.SignToken(nipcash.Token{HRP: "satscash", WalletPubkey: strings.Repeat("22", 32), Secret: strings.Repeat("33", 32),
		RelayURLs: []string{"wss://r.example"}, AttestedAmountMillis: &amount}, nipcash.SchemeMintV1, nil, minterXOnly, sign)
	if err != nil {
		t.Fatal(err)
	}
	s, err := nipcash.Encode(tok)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// E-WD-07 and the cancel answers.
func TestWithdraw_Cancel(t *testing.T) {
	ctx, cancel := e2eContext(t)
	defer cancel()
	w := newWDHub(t, false)
	w.locked(func() { w.sentBy = time.Now().Unix() - nipcash.CancelRequestAfter - 10 })
	q := w.quote(t, wdAddr)
	j := newTestJournal()
	if _, err := w.s.withdraw(ctx, w.bill, nipcash.CashWithdrawParams{Quote: q, Credential: w.cred()}, j, "dev"); err != nil {
		t.Fatal(err)
	}
	pc := j.all()[0]
	before := len(w.methods())
	var refused *CancelRefusedError
	if _, err := w.s.cancelWithdraw(ctx, w.bill, pc, w.cred(), q.SentBy+nipcash.CancelRequestAfter-1); !errors.As(err, &refused) || !refused.Local {
		t.Errorf("local guard: %v", err)
	}
	if len(w.methods()) != before {
		t.Error("the local guard sent something")
	}
	var sawSecret bool
	w.locked(func() {
		w.cancel = func(item transport.Item) *transport.Result {
			sawSecret = strings.Contains(string(item.Params), "cash_secret")
			return &transport.Result{ID: item.ID, Error: &transport.ResultError{Code: nipcash.ErrCodeRestricted, Message: "too early"}}
		}
	})
	now := time.Now().Unix()
	if _, err := w.s.cancelWithdraw(ctx, w.bill, pc, w.cred(), now); !errors.As(err, &refused) || refused.Code != nipcash.ErrCodeRestricted || refused.Local {
		t.Errorf("RESTRICTED: %v", err)
	}
	w.locked(func() {
		if sawSecret {
			t.Error("a pubkey cancel carries a cash_secret")
		}
	})
	w.locked(func() {
		w.cancel = func(item transport.Item) *transport.Result {
			return &transport.Result{ID: item.ID, Error: &transport.ResultError{Code: nipcash.ErrCodeRequestReused, Message: "x"}}
		}
	})
	if _, err := w.s.cancelWithdraw(ctx, w.bill, pc, w.cred(), now); !errors.Is(err, ErrCancelAnswer) {
		t.Errorf("REQUEST_REUSED: %v", err)
	}
	broadcast := w.lastPayout()
	vout, depth := uint32(1), uint32(0)
	broadcast.State, broadcast.Txid, broadcast.Vout, broadcast.Depth = nipcash.PayoutBroadcast, strings.Repeat("ab", 32), &vout, &depth
	w.locked(func() { w.cancel = func(item transport.Item) *transport.Result { return wdResult(item, &broadcast) } })
	if p, err := w.s.cancelWithdraw(ctx, w.bill, pc, w.cred(), now); err != nil || p.State != nipcash.PayoutBroadcast {
		t.Errorf("answer: %+v %v", p, err)
	}
	if len(j.all()) != 1 {
		t.Error("a cancel request was written ahead")
	}
}

// T21-24.
func TestT21_24_PayoutFinished(t *testing.T) {
	for st, want := range map[nipcash.PayoutState]bool{nipcash.PayoutFinal: true, nipcash.PayoutCancelled: true,
		nipcash.PayoutQueued: false, nipcash.PayoutBroadcast: false, nipcash.PayoutConfirmed: false} {
		if PayoutFinished(&nipcash.Payout{State: st}) != want {
			t.Errorf("%s", st)
		}
	}
	if PayoutFinished(nil) {
		t.Error("nil")
	}
}

// T21-32: WatchPayout subscribes on the accepted relays and signals a notice.
func TestT21_32_WatchPayout(t *testing.T) {
	w := newWDHub(t, false)
	notify := strings.Repeat("0123456789abcdef", 4)
	other := cashtest.NewRelay(t)
	if got := w.s.payoutWatchRelays(notify, []string{other.URL()}); len(got) < 2 || got[0] != other.URL() {
		t.Errorf("accepted first: %v", got)
	}
	ctx, cancel := e2eContext(t)
	defer cancel()
	ch, err := w.s.WatchPayout(ctx, notify, []string{other.URL()})
	if err != nil {
		t.Fatal(err)
	}
	ev, err := nipCO.NewOrderNotice(notify)
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.After(5 * time.Second)
	for {
		other.Publish(ev)
		select {
		case <-ch:
			return
		case <-deadline:
			t.Fatal("no signal")
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// T11-49, T11-50: Recover and Replace for cash_withdraw.
func TestT11_49_RecoverReplaceWithdraw(t *testing.T) {
	now := time.Now()
	base := nipcash.PendingCall{RequestID: testRequestID, Method: nipcash.MethodCashWithdraw, Device: "a", Deadline: now.Unix() - 10,
		MaxRequestAge: 300, Quote: "{}", HoldsUntil: 100, Notify: strings.Repeat("ab", 32), Target: strings.Repeat("cd", 32),
		Secrets: []nipcash.PendingSecret{{Secret: strings.Repeat("11", 32), Role: nipcash.SecretRoleRemint}}}
	with := func(f func(*nipcash.PendingCall)) nipcash.PendingCall { pc := base; f(&pc); return pc }
	unknown := LookupOutcome{Status: nipcash.LookupUnknown}
	for name, c := range map[string]struct {
		pc     nipcash.PendingCall
		device string
		want   Action
	}{
		"lost replies":   {base, "a", Action{Kind: ActionReplace}},
		"REQUEST_REUSED": {with(func(p *nipcash.PendingCall) { p.SawReused = true }), "a", Action{Kind: ActionConfirmReplace}},
		"quote_expired": {with(func(p *nipcash.PendingCall) { p.LastCode, p.LastDetail = "RESTRICTED", nipcash.DetailQuoteExpired }), "a",
			Action{Kind: ActionCloseRefused, Reason: "RESTRICTED/quote_expired"}},
		"non-owner":      {base, "b", Action{Kind: ActionShow, Reason: "foreign_device"}},
		"recovered call": {with(func(p *nipcash.PendingCall) { p.Quote = "" }), "a", Action{Kind: ActionShow, Reason: "no_quote"}},
	} {
		got, err := Recover(context.Background(), c.pc, answered(unknown), c.device, now)
		if err != nil || got.Kind != c.want.Kind || got.Reason != c.want.Reason {
			t.Errorf("%s: %+v %v", name, got, err)
		}
	}
	if got, _ := Recover(context.Background(), base, answered(LookupOutcome{Status: nipcash.LookupExpired}), "a", now); got.Kind != ActionShow {
		t.Errorf("expired: %+v", got)
	}

	// Replace keeps notify, quote and the re-mint secret; a new quote needs confirmation.
	ctx := context.Background()
	derived, _ := nipcash.DerivedRequestID(testRequestID)
	var sent nipcash.PendingCall
	send := func(_ context.Context, next nipcash.PendingCall) error { sent = next; return nil }
	// M8: re-placed automatically only while its quote holds at the known tip.
	for tip, reason := range map[int64]string{0: "quote_unchecked", 101: "quote_expired"} {
		sent = nipcash.PendingCall{}
		act, err := Replace(ctx, base, newTestJournal(), ReplaceOptions{ChainTip: tip}, send)
		if err != nil || act.Kind != ActionShow || act.Reason != reason || sent.RequestID != "" {
			t.Errorf("tip %d: %+v %v (sent %q), want shown %s", tip, act, err, sent.RequestID, reason)
		}
	}
	// A caller with no chain view of its own treats the quote as valid until
	// refused quote_expired (NIP-CP §Lookups); a tip it does know still wins.
	sent = nipcash.PendingCall{}
	if act, err := Replace(ctx, base, newTestJournal(), ReplaceOptions{NoChainView: true}, send); err != nil || act.Kind != ActionWait || sent.RequestID != derived {
		t.Errorf("no chain view: %+v %v (sent %q)", act, err, sent.RequestID)
	}
	expired := with(func(p *nipcash.PendingCall) { p.LastDetail = nipcash.DetailQuoteExpired })
	sent = nipcash.PendingCall{}
	if act, err := Replace(ctx, expired, newTestJournal(), ReplaceOptions{NoChainView: true}, send); err != nil || act.Reason != "quote_expired" || sent.RequestID != "" {
		t.Errorf("no chain view after quote_expired: %+v %v", act, err)
	}
	invalid := with(func(p *nipcash.PendingCall) { p.LastDetail = nipcash.DetailQuoteInvalid })
	sent = nipcash.PendingCall{}
	if act, err := Replace(ctx, invalid, newTestJournal(), ReplaceOptions{NoChainView: true}, send); err != nil || act.Kind != ActionShow || act.Reason != "quote_invalid" || sent.RequestID != "" {
		t.Errorf("no chain view after quote_invalid: %+v %v (sent %q)", act, err, sent.RequestID)
	}
	if act, err := Replace(ctx, base, newTestJournal(), ReplaceOptions{NoChainView: true, ChainTip: 101}, send); err != nil || act.Reason != "quote_expired" {
		t.Errorf("no chain view with a known tip past the quote: %+v %v", act, err)
	}
	if _, err := Replace(ctx, base, newTestJournal(), ReplaceOptions{ChainTip: 100}, send); err != nil {
		t.Fatal(err)
	}
	if sent.RequestID != derived || sent.Notify != base.Notify || sent.Quote != base.Quote || len(sent.Secrets) != 1 || sent.Secrets[0].Role != nipcash.SecretRoleRemint {
		t.Errorf("re-placed %+v", sent)
	}
	w := newWDHub(t, false)
	q := w.quote(t, wdAddr)
	qpc := base
	qpc.Target = w.bill.Target()
	if _, err := Replace(ctx, qpc, newTestJournal(), ReplaceOptions{Quote: q}, send); !errors.Is(err, ErrReplaceNeedsConfirmation) {
		t.Errorf("quote without confirmation: %v", err)
	}
	if _, err := Replace(ctx, with(func(p *nipcash.PendingCall) { p.SawReused = true }), newTestJournal(), ReplaceOptions{}, send); !errors.Is(err, ErrReplaceNeedsConfirmation) {
		t.Errorf("after REQUEST_REUSED without confirmation: %v", err)
	}
	if _, err := Replace(ctx, qpc, newTestJournal(), ReplaceOptions{Quote: q, Confirmed: true}, send); err != nil {
		t.Fatal(err)
	}
	raw, _ := q.Raw()
	if sent.Quote != raw || sent.Address != wdAddr || sent.Notify != base.Notify || sent.HoldsUntil != q.HoldsUntil {
		t.Errorf("new quote %+v", sent)
	}
	if _, err := Replace(ctx, with(func(p *nipcash.PendingCall) { p.Quote = "" }), newTestJournal(), ReplaceOptions{}, send); err == nil {
		t.Error("a recovered call was re-placed")
	}
}

// T11-40 (withdraw), T21-28: no error text carries a secret, notify, request_id, the address or
// the quote.
func TestT21_28_ErrorTextPrivate(t *testing.T) {
	ctx, cancel := e2eContext(t)
	defer cancel()
	w := newWDHub(t, true)
	q := w.quote(t, wdAddr)
	var errs []error
	j := newTestJournal()
	w.locked(func() { w.badHash = true })
	_, err := w.s.withdraw(ctx, w.bill, nipcash.CashWithdrawParams{Quote: q, Credential: w.cred()}, j, "dev")
	errs = append(errs, err)
	w.locked(func() { w.badHash = false })
	for _, code := range []string{nipcash.ErrCodeBadRequest, nipcash.ErrCodeRestricted} {
		w.locked(func() { w.callErr = &transport.ResultError{Code: code, Message: "refused", Detail: "own_address"} })
		_, err := w.s.withdraw(ctx, w.bill, nipcash.CashWithdrawParams{Quote: q, Credential: w.cred()}, j, "dev")
		errs = append(errs, err)
	}
	w.locked(func() {
		w.quoteErr = &transport.ResultError{Code: nipcash.ErrCodeRestricted, Message: "refused", Detail: "network"}
	})
	_, err = w.s.quoteWithdraw(ctx, w.bill, "bitcoin", nipcash.CashWithdrawQuoteParams{Address: wdAddr, Credential: w.cred()})
	errs = append(errs, err)
	_, err = w.s.quoteWithdraw(ctx, w.bill, "bitcoin", nipcash.CashWithdrawQuoteParams{Address: wdAddrMix, Credential: w.cred()})
	errs = append(errs, err)
	pc := j.all()[0]
	_, err = w.s.cancelWithdraw(ctx, w.bill, pc, w.cred(), 0)
	errs = append(errs, err)
	raw, _ := q.Raw()
	secrets := []string{w.secret, pc.Notify, pc.RequestID, wdAddr, strings.ToLower(wdAddrMix), q.ID, raw}
	for _, s := range pc.Secrets {
		secrets = append(secrets, s.Secret)
	}
	for i, err := range errs {
		if err == nil {
			t.Errorf("case %d: no error", i)
			continue
		}
		for _, s := range secrets {
			if s != "" && strings.Contains(strings.ToLower(err.Error()), s) {
				t.Errorf("case %d: %q carries %q", i, err.Error(), s)
			}
		}
	}
}

// A cancelled payout's re-minted bill, answered to a cancel request, is
// checked as a lookup's is: this Hub's bill for the named wallet, or the
// answer is refused.
func TestWithdraw_CancelChecksTheRemint(t *testing.T) {
	ctx, cancel := e2eContext(t)
	defer cancel()
	w := newWDHub(t, true)
	w.locked(func() { w.sentBy = time.Now().Unix() - nipcash.CancelRequestAfter - 10 })
	q := w.quote(t, wdAddr)
	j := newTestJournal()
	if _, err := w.s.withdraw(ctx, w.bill, nipcash.CashWithdrawParams{Quote: q, Credential: w.cred()}, j, "dev"); err != nil {
		t.Fatal(err)
	}
	pc := j.all()[0]
	now := time.Now().Unix()
	cancelled := w.lastPayout()
	cancelled.State, cancelled.NewWalletPubkey = nipcash.PayoutCancelled, strings.Repeat("22", 32)
	answer := func(tok string) {
		c := cancelled
		c.NewWalletToken = nipcash.TokenOrTombstone{Token: tok}
		w.locked(func() { w.cancel = func(item transport.Item) *transport.Result { return wdResult(item, &c) } })
	}

	own := remintToken(t, w.hub.hubPriv, w.hub.hubXOnly)
	answer(own)
	if p, err := w.s.cancelWithdraw(ctx, w.bill, pc, w.cred(), now); err != nil || p.NewWalletToken.Token != own {
		t.Fatalf("own re-mint: %+v %v", p, err)
	}
	otherPriv, otherX := sessionKeypair(t)
	answer(remintToken(t, otherPriv, otherX))
	if p, err := w.s.cancelWithdraw(ctx, w.bill, pc, w.cred(), now); !errors.Is(err, ErrCancelAnswer) || p != nil {
		t.Errorf("foreign re-mint: %+v %v, want ErrCancelAnswer", p, err)
	}
	cancelled.NewWalletPubkey = strings.Repeat("44", 32)
	answer(own)
	if p, err := w.s.cancelWithdraw(ctx, w.bill, pc, w.cred(), now); !errors.Is(err, ErrCancelAnswer) || p != nil {
		t.Errorf("re-mint for another wallet: %+v %v, want ErrCancelAnswer", p, err)
	}
}
