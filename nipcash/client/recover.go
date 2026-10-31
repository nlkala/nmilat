package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ohstr/nmilat/nipCO"
	"github.com/ohstr/nmilat/nipcash"
)

// ActionKind is what Recover tells the caller to do with a pending call.
type ActionKind int

const (
	// ActionClose: the call is decided; Result holds its rebuilt response.
	ActionClose ActionKind = iota
	// ActionWait: ask again at Until (0: on the caller's next refresh).
	ActionWait
	// ActionShow: surface the call to its user; Reason says why.
	ActionShow
	// ActionReplace: re-place it automatically under the derived id.
	ActionReplace
	// ActionConfirmReplace: re-place it only with the user's confirmation.
	ActionConfirmReplace
	// ActionCloseFailed: close it; it never committed.
	ActionCloseFailed
	// ActionCloseRefused: close it as refused with Reason (its LastCode); the
	// user re-runs the operation as a new call.
	ActionCloseRefused
)

// Action is Recover's answer.
type Action struct {
	Kind   ActionKind
	Until  int64
	Reason string
	Result json.RawMessage
}

// LookupOutcome is a lookup's answer as Recover needs it. NotServed is an
// omitted or unanswered lookup; Malformed an answer that did not parse; Code an
// item-level refusal, never folded into NotServed.
type LookupOutcome struct {
	Status     nipcash.LookupStatus
	RetryAfter int64
	Result     json.RawMessage
	NotServed  bool
	Malformed  bool
	Code       string
}

// MaxReplaceAttempts bounds the re-placements of one call.
const MaxReplaceAttempts = 3

// Recover decides what to do with a pending call from its lookup (NIP-CASH
// §Looking Up a Call, Finality): done closes it (an order's done waits, its
// caller closes on the order's state); pending waits; unknown is the owner's to
// re-place (a mint or order only with confirmation; a call refused with a code
// is closed as refused instead), a foreign device's to show; expired is shown;
// an unanswered lookup waits, and is shown past deadline + H; a malformed one is
// shown.
//
// A cash_withdraw (NIP-CASH §Withdrawals): a done carries the payout's current
// state, whose receipt is checked against the call (else the answer reads as
// malformed and the call waits, never closed); it waits with the payout as
// Result, and the caller closes it on PayoutFinished. unknown re-places it
// automatically after only lost replies (Replace then checks the quote is
// still valid at ReplaceOptions.ChainTip, or the caller has no chain view,
// ReplaceOptions.NoChainView), with confirmation after
// REQUEST_REUSED, and closes it refused (reason code/detail) after a coded
// refusal; a call recovered from withdraw_request_id (no quote) is never
// re-placed.
func Recover(ctx context.Context, pc nipcash.PendingCall, lookup func(context.Context, nipcash.PendingCall) (LookupOutcome, error), localDevice string, now time.Time) (Action, error) {
	lo, err := lookup(ctx, pc)
	if err != nil {
		return Action{}, err
	}
	unanswered := func() Action {
		if now.Unix() > pc.Deadline+pc.MaxRequestAge {
			return Action{Kind: ActionShow, Reason: "lookup_unanswered"}
		}
		return Action{Kind: ActionWait}
	}
	isWithdraw := pc.Method == nipcash.MethodCashWithdraw
	if isWithdraw && !lo.Malformed && lo.Code == "" && !lo.NotServed && lo.Status == nipcash.LookupDone {
		if err := checkWithdrawDone(pc, lo.Result); err != nil {
			lo.Malformed = true
		}
	}
	switch {
	case lo.Malformed && isWithdraw:
		return Action{Kind: ActionWait, Reason: "lookup_malformed"}, nil
	case lo.Malformed:
		return Action{Kind: ActionShow, Reason: "lookup_malformed"}, nil
	case lo.NotServed, lo.Code == nipcash.ErrCodeRequestReused:
		return unanswered(), nil
	case lo.Code != "":
		return Action{Kind: ActionShow, Reason: "lookup_refused:" + lo.Code}, nil
	}
	isOrder := pc.Method == nipcash.MethodCashOrder
	if isOrder && pc.PayBy != 0 && (lo.Status == nipcash.LookupUnknown || lo.Status == nipcash.LookupExpired) {
		// The held reply proves the order was placed: never re-placed. Before
		// pay_by it is shown; after it the caller ends the order unpaid.
		if now.Unix() < pc.PayBy {
			return Action{Kind: ActionShow, Reason: "lookup_unknown"}, nil
		}
		return Action{Kind: ActionWait, Reason: string(lo.Status)}, nil
	}
	switch lo.Status {
	case nipcash.LookupDone:
		if isOrder || isWithdraw {
			return Action{Kind: ActionWait, Result: lo.Result}, nil
		}
		return Action{Kind: ActionClose, Result: lo.Result}, nil
	case nipcash.LookupPending:
		return Action{Kind: ActionWait, Until: lo.RetryAfter}, nil
	case nipcash.LookupExpired:
		return Action{Kind: ActionShow, Reason: "lookup_expired"}, nil
	case nipcash.LookupUnknown:
		switch {
		case pc.Device != localDevice:
			return Action{Kind: ActionShow, Reason: "foreign_device"}, nil
		case pc.LastCode != "" && isWithdraw && pc.LastDetail != "":
			return Action{Kind: ActionCloseRefused, Reason: pc.LastCode + "/" + string(pc.LastDetail)}, nil
		case pc.LastCode != "":
			return Action{Kind: ActionCloseRefused, Reason: pc.LastCode}, nil
		case isWithdraw && pc.Quote == "":
			return Action{Kind: ActionShow, Reason: "no_quote"}, nil
		case pc.Attempt >= MaxReplaceAttempts:
			return Action{Kind: ActionShow, Reason: "replace_cap"}, nil
		case pc.Method == nipcash.MethodMintCash || isOrder || (isWithdraw && pc.SawReused):
			return Action{Kind: ActionConfirmReplace}, nil
		default:
			return Action{Kind: ActionReplace}, nil
		}
	}
	return Action{Kind: ActionShow, Reason: "lookup_malformed"}, nil
}

// ReplaceOptions tunes Replace.
type ReplaceOptions struct {
	// Confirmed is the user's explicit confirmation, required for a mint or order.
	Confirmed bool
	// DerivedEntryExists reports a written-ahead record of the derived id the
	// caller keeps elsewhere.
	DerivedEntryExists func(context.Context, string) (bool, error)
	// MaxAttempts defaults to MaxReplaceAttempts.
	MaxAttempts int
	// Quote is a new quote for a re-placed cash_withdraw, one the caller
	// verified with VerifyQuoteAnswer and its user confirmed; it needs
	// Confirmed. nil keeps the call's quote.
	Quote *nipcash.WithdrawQuote
	// ChainTip is the height of the chain tip the caller knows. A cash_withdraw
	// is re-placed automatically (without Confirmed) only while its quote is
	// still valid at ChainTip (NIP-CASH §Withdrawals); with ChainTip 0, or past
	// the quote's holds_until, Replace shows the call instead, so its user can
	// get a new quote, unless NoChainView is set.
	ChainTip int64
	// NoChainView says the caller has no view of the chain of its own. With
	// ChainTip 0, a cash_withdraw's quote then counts as valid until a call of
	// this withdrawal was refused quote_expired (NIP-CP §Lookups), which Recover
	// closes as ActionCloseRefused, so Replace re-places it without Confirmed.
	// A call whose LastDetail is quote_expired or quote_invalid is still shown
	// (quote_expired, quote_invalid), so its user can get a new quote: an error
	// answer to a cash_withdraw is never re-placed with the same quote
	// (NIP-CASH §Withdrawals).
	NoChainView bool
}

type replacementKey struct{}

// withReplacement marks ctx as carrying the re-placement of a call.
func withReplacement(ctx context.Context, next nipcash.PendingCall) context.Context {
	return context.WithValue(ctx, replacementKey{}, next)
}

// asReplacement completes a call about to be written ahead with what Replace
// knows of it, when ctx carries the re-placement of the same request_id.
func asReplacement(ctx context.Context, pc nipcash.PendingCall) nipcash.PendingCall {
	next, ok := ctx.Value(replacementKey{}).(nipcash.PendingCall)
	if !ok || next.RequestID != pc.RequestID {
		return pc
	}
	pc.DerivedFrom, pc.Attempt = next.DerivedFrom, next.Attempt
	if pc.Notify == "" {
		pc.Notify = next.Notify
	}
	return pc
}

// errReplacementMismatch is a re-placement whose send path packed no call
// under the derived id, or packed it with another method, target, amount,
// secrets or quote than the call it replaces (NIP-CASH §Looking Up a Call,
// Finality: always the derived id, never a fresh one).
var errReplacementMismatch = errors.New("nipcash/client: the re-placement must send the replaced call, unchanged, under its derived id")

// checkReplacement refuses, before anything is written ahead, calls that do not
// carry the re-placement ctx marks exactly once and unchanged. Without one in
// ctx it passes.
func checkReplacement(ctx context.Context, calls []nipcash.PendingCall) error {
	next, ok := ctx.Value(replacementKey{}).(nipcash.PendingCall)
	if !ok {
		return nil
	}
	n := 0
	for _, pc := range calls {
		if pc.RequestID != next.RequestID {
			continue
		}
		n++
		if pc.Method != next.Method || pc.Target != next.Target || pc.Amount != next.Amount ||
			pc.SourceSecret != next.SourceSecret || !sameSecrets(pc.Secrets, next.Secrets) ||
			pc.Quote != next.Quote || pc.Address != next.Address {
			return errReplacementMismatch
		}
	}
	if n != 1 {
		return errReplacementMismatch
	}
	return nil
}

func sameSecrets(a, b []nipcash.PendingSecret) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Replace re-places a call after an unknown lookup: under its derived id, never
// a fresh one, and never twice (a derived id already placed or written ahead
// closes the call instead). It sends the derived call through send, and
// records the derived id once the send path has written it ahead (or once send
// returns nil without the SDK writing it); send must go through an SDK entry point taking a
// Journal (CashTransfer, TransferMany, HubClient.MintCash, ...) with the ctx it
// is given: that path writes the call ahead exactly once, with its exact
// deadline, DerivedFrom and Attempt, and builds fresh proofs. It refuses,
// before writing anything, a send that packs no call under the derived id or
// changes its method, target, amount, secrets or quote.
func Replace(ctx context.Context, pc nipcash.PendingCall, j Journal, o ReplaceOptions, send func(context.Context, nipcash.PendingCall) error) (Action, error) {
	if j == nil {
		return Action{}, ErrNoJournal
	}
	if (pc.Method == nipcash.MethodMintCash || pc.Method == nipcash.MethodCashOrder) && !o.Confirmed {
		return Action{}, ErrReplaceNeedsConfirmation
	}
	isWithdraw := pc.Method == nipcash.MethodCashWithdraw
	if isWithdraw && (o.Quote != nil || pc.SawReused) && !o.Confirmed {
		return Action{}, ErrReplaceNeedsConfirmation
	}
	if isWithdraw && pc.Quote == "" {
		return Action{}, errNoWithdrawQuote
	}
	if o.Quote != nil && !isWithdraw {
		return Action{}, fmt.Errorf("nipcash/client: only a cash_withdraw takes a new quote")
	}
	max := o.MaxAttempts
	if max == 0 {
		max = MaxReplaceAttempts
	}
	if pc.Attempt >= max {
		return Action{Kind: ActionShow, Reason: "replace_cap"}, nil
	}
	id, err := nipcash.DerivedRequestID(pc.RequestID)
	if err != nil {
		return Action{}, err
	}
	placed, err := j.DerivedPlaced(ctx, id)
	if err != nil {
		return Action{}, err
	}
	if !placed && o.DerivedEntryExists != nil {
		if placed, err = o.DerivedEntryExists(ctx, id); err != nil {
			return Action{}, err
		}
	}
	if placed {
		return Action{Kind: ActionClose, Reason: "derived_already_placed"}, nil
	}
	// The quote is checked only once the derived id is known not to have gone
	// out: a re-placement already placed closes the call, whatever its quote.
	if isWithdraw {
		holdsUntil := pc.HoldsUntil
		if o.Quote != nil {
			holdsUntil = o.Quote.HoldsUntil
		}
		switch {
		case o.ChainTip > 0 && !nipcash.QuoteValidAt(holdsUntil, o.ChainTip):
			return Action{Kind: ActionShow, Reason: "quote_expired"}, nil
		case o.ChainTip <= 0 && o.NoChainView && o.Quote == nil && pc.LastDetail == nipcash.DetailQuoteExpired:
			return Action{Kind: ActionShow, Reason: "quote_expired"}, nil
		case o.ChainTip <= 0 && o.NoChainView && o.Quote == nil && pc.LastDetail == nipcash.DetailQuoteInvalid:
			return Action{Kind: ActionShow, Reason: "quote_invalid"}, nil
		case o.ChainTip <= 0 && !o.Confirmed && !o.NoChainView:
			return Action{Kind: ActionShow, Reason: "quote_unchecked"}, nil
		}
	}
	next := pc
	next.RequestID, next.DerivedFrom, next.Attempt, next.LastCode, next.SawReused = id, pc.RequestID, pc.Attempt+1, "", false
	if pc.Method == nipcash.MethodCashOrder {
		notify, err := nipCO.NewNotify()
		if err != nil {
			return Action{}, err
		}
		next.Notify = notify
	}
	if o.Quote != nil {
		raw, err := o.Quote.Raw()
		if err != nil {
			return Action{}, err
		}
		if o.Quote.Wallet != pc.Target {
			return Action{}, fmt.Errorf("nipcash/client: the new quote is for another bill")
		}
		// notify and the re-mint secret are kept (NIP-CASH §Withdrawals).
		next.Quote, next.Address, next.HoldsUntil, next.Amount, next.Receipt = raw, o.Quote.Address, o.Quote.HoldsUntil, o.Quote.AmountMillis, ""
	}
	h := max64(pc.MaxRequestAge, nipcash.MinMaxRequestAge)
	next.Deadline = time.Now().Unix() + nipcash.DeadlineWindow(h)
	keepUntil := next.Deadline + h
	rec := &derivedRecord{id: id, record: func(ctx context.Context) error {
		var err error
		if r, ok := j.(DerivedParentRecorder); ok {
			err = r.RecordDerivedFrom(ctx, id, pc.RequestID, keepUntil)
		} else {
			err = j.RecordDerived(ctx, id, keepUntil)
		}
		if err != nil {
			return fmt.Errorf("nipcash/client: record derived id: %w", err)
		}
		return nil
	}}
	// The derived call is written ahead once, by the send path, with the exact
	// deadline it is packed under; the context carries what the send path
	// cannot rebuild from the params (DerivedFrom, Attempt, a fresh notify).
	// The derived id is recorded once that write-ahead succeeds, so a send that
	// failed before it (an unusable Hub, a refused re-placement) records
	// nothing, and a later Replace can still place it.
	err = send(context.WithValue(withReplacement(ctx, next), derivedRecordKey{}, rec), next)
	if !rec.fired {
		if err != nil {
			return Action{}, err
		}
		// A send that wrote nothing ahead through the SDK still placed the id.
		if err := rec.fire(ctx); err != nil {
			return Action{}, err
		}
	}
	if err != nil {
		return Action{}, err
	}
	return Action{Kind: ActionWait}, nil
}

type derivedRecordKey struct{}

// derivedRecord records a re-placement's derived id once.
type derivedRecord struct {
	id     string
	once   sync.Once
	fired  bool
	err    error
	record func(context.Context) error
}

func (r *derivedRecord) fire(ctx context.Context) error {
	r.once.Do(func() { r.fired, r.err = true, r.record(ctx) })
	return r.err
}

// derivedWritten records the derived id of the re-placement ctx carries once
// calls holding it were written ahead.
func derivedWritten(ctx context.Context, calls []nipcash.PendingCall) error {
	r, ok := ctx.Value(derivedRecordKey{}).(*derivedRecord)
	if !ok {
		return nil
	}
	for _, pc := range calls {
		if pc.RequestID == r.id {
			return r.fire(ctx)
		}
	}
	return nil
}

// errNoWithdrawQuote is a withdrawal re-placed without its quote: a call
// recovered from withdraw_request_id is never re-placed.
var errNoWithdrawQuote = errors.New("nipcash/client: a withdrawal without its quote is never re-placed")

// checkWithdrawDone checks a cash_withdraw lookup's done: a payout whose receipt
// verifies against the call (its Hub, bill, quote id, request_id and the
// quote's address, payout and sent_by; for a call without a quote, those of the
// payout).
func checkWithdrawDone(pc nipcash.PendingCall, result json.RawMessage) error {
	p, err := nipcash.ParsePayout(result)
	if err != nil {
		return err
	}
	return verifyCallReceipt(pc, p)
}

// verifyCallReceipt checks p's receipt against the call pc, and that p pays
// what the call's quote says.
func verifyCallReceipt(pc nipcash.PendingCall, p *nipcash.Payout) error {
	e := nipcash.ReceiptExpectation{HubXOnly: pc.HubIdentity, Target: pc.Target, RequestID: pc.RequestID,
		Address: p.Address, PayoutMillis: p.PayoutMillis, SentBy: p.SentBy}
	if pc.Quote != "" {
		q, err := nipcash.ParseWithdrawQuote(pc.Quote)
		if err != nil {
			return fmt.Errorf("%w: the call's quote", nipcash.ErrReceiptInvalid)
		}
		e.QuoteID, e.Address, e.PayoutMillis, e.SentBy = q.ID, q.Address, q.PayoutMillis, q.SentBy
		if p.Address != q.Address || p.PayoutMillis != q.PayoutMillis || p.SentBy != q.SentBy || p.AmountMillis != q.AmountMillis {
			return fmt.Errorf("%w: the payout differs from the quote", nipcash.ErrReceiptInvalid)
		}
	}
	return nipcash.VerifyWithdrawReceipt(p.Receipt, e)
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
