# Changelog

## [0.5.0]

NIP-CASH's private transport, NIP-29 groups hosted by the relay, huddle audio,
NIP-86 relay management and HTTP bridges, and a long run of relay access,
privacy and performance fixes.

### Added

- Adds `nipOA.Sign`, `Signer`, `KeySigner`, `RemoteSigner`, `Caller`,
  `MethodSignAttestation` (`nipoa_attest`), `Until` and
  `Conditions.ExpiredAt`: an owner attestation signed by a key here or
  through a remote signer, verified before it is returned; and
  `nipLS.AttestationMethod`, which serves `nipoa_attest` from a signer's
  custom methods, refusing every request unless its `AttestFunc` allows it.
- Adds `nip98.Sign` and `nip98.EventSigner`: a NIP-98 header signed through
  any event signer (`nip46.LocalKey`, `nipLS.Client`, a NIP-46 client), each
  event with a random `salt` tag so two requests in one second get two ids;
  `NewAuthHeader` is `Sign` with a hex key.
- Adds `nip98.Options.Window`, `Seen`, `ExpectURL`, `ExpectMethod` and `Now`,
  and `nip98.SeenIDs` (`NewSeenIDs`, `ErrReplayed`, `ErrReplayCacheFull`): a
  server that gates on NIP-98 accepts each event once within its window, at
  the exact URL and method it expects.
- Adds `nippf`, NIP-PF Poll Forms: forms (kind 1069), responses and
  withdrawals (1019), results (1070) and NIP-88 poll extras (`rec`,
  `noteRequired`, `why`, notes on votes). Strict parsers, every form,
  response and result validity rule, the lifecycle (open, ended, decided,
  default, cancelled, expired), selection and tally, builders that read back
  what they write, and private delivery: one rumor gift-wrapped to each `p`
  and to the sender.
- Adds `nip59.WrapOption`s to `WrapSeal`: `WithWrapExpiration` puts a NIP-40
  expiration on a wrap sealed through a signer, as it does for `WrapWithKey`.
- Adds `relayclient.NWCOption`, `WithAnswerWait` and `DefaultAnswerWait`
  (`NewNWCClient` takes options), and `client.WithRelayAnswerWait`.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Adds `client.WithExpectedMinter` and `client.ErrBillUnverified`: `Connect`,
  `ConnectCashSlice` and `ConnectZap` refuse a bill whose provenance does not
  verify, or not against the expected Hub, before any relay is dialled.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Adds `client.ErrPowRemineTooHard`, a relay asking for more proof of work
  than the session's budget after the first publish attempt.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Adds `nipcash.ErrRowAmbiguous`, a scope-all roster with two unclaimed rows
  for the caller.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Adds `transport.ErrEnvelopeKeys` and `transport.Item.UnmarshalJSON`, which
  reads an item by its exact keys.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Adds to `testdata/vectors/connection-key.json` a leading-`@` email, an empty
  and a decomposed local part, a domain and an email domain at and one past
  253 characters, and the platform `nostr` refused.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Adds `resultToState` to `testdata/vectors/pocket.json` (K-POCKET-7): NIP-CP
  §Balance's result to state table, so every client tests the same mapping.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Adds `nipcash.QuoteForSlice`, binding a withdraw quote to the bill, the
  slice's identity and its current amount (the fake Hub uses it; conformance
  adds CT-WD-27, same identity at another amount), `ErrClaimAmountUnbound`,
  `transport.ErrProofWrongExpiration` (also matching `ErrProofWrongNonce`),
  `transport.ServeVerdict.Served`, `client.ReplaceOptions.ChainTip`,
  `TransferOutcome.CallError`/`ConsolidateOutcome.CallError`,
  `nipcw.Credential.Pubkey` and `nipcw/client.ErrJoinSigner`,
  `nipCP.Synced.Answered` and `ErrTooFewRelays`, and
  `listing.VersionBook.CurrentByD` with `TapCache.Passed`.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Adds `client.ReplaceOptions.NoChainView`: a caller with no view of the
  chain re-places a withdrawal while its quote has not been refused
  `quote_expired` (NIP-CP §Lookups), instead of showing it
  `quote_unchecked`; `listing.TapResult.Passed`; and `textsan.MaxClusterBytes`
  with `textsan.Options.MaxBytes`.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Adds `client.NeedsStatusError`/`ErrNeedsStatus` for a redeem whose outcome
  is unknown (never resent) and `Client.RedeemFollowUp`, `BillNoSlice`,
  `LocalReason`/`LocalRetryable` with `ReasonStatusWidened`,
  `TransferOutcome.Remainder`, and
  `nipcash.CashTransferParams.NeedsRemainder`/`WithRemainder` with
  `SecretRoleRemainder`. ([#215](https://github.com/ohstr/nmilat/pull/215))
- Adds mint lookups, faults on `mint_cash` and held redeem payments to the
  fake Hub, and the E-MINT-02..05/07/09/11/13 tests.
  ([#215](https://github.com/ohstr/nmilat/pull/215))
- Adds tests covering the client scenarios of WP-11 end to end against the
  fake ledger (mint, transfer, status, redeem, lookup and envelope cases)
  and `RedeemProgressOf`, and tags existing tests with the T11 and E-* ids
  they cover. ([#214](https://github.com/ohstr/nmilat/pull/214))
- Adds `nipcash.Source.Ordered` and enforces `ErrMixedOrdered` in
  `ConsolidateMany`, `CashConsolidate` and the composites, which carry an
  `Ordered` mark onto the bills they build.
  ([#213](https://github.com/ohstr/nmilat/pull/213))
- Adds `Client.Unreconciled`, `Client.CashRedeemWith` with
  `RedeemOptions{AutoAmount, Payee}` and `ErrInvoiceAmount`,
  `ErrTooManySources`, `OutcomeNotSent`, and
  `nipcw/client.WalletGuard.Check`, `ObserveRefusal` and `Do` with
  `ErrWalletFrozen` and `ErrStaleKey`.
  ([#213](https://github.com/ohstr/nmilat/pull/213))
- Adds zap transfers, `cash_consolidate`, live-wallet and spin-off caps and
  archive drops to the fake ledger, and a ledger circle (joins, join
  lookups, member requests, extra and revoked connections, issued wallets);
  the fake's conformance adapter implements every `Admin` hook.
  ([#212](https://github.com/ohstr/nmilat/pull/212))
- Adds conformance cases CT-JOIN-01..18, CT-JLK-01..06, CT-POW-04..06,
  CT-XFER-04/-12/-20, CT-CONS-01..03/-05/-08/-09 and CT-LK-15, the optional
  `CircleFunder` interface, and the E-JOIN end-to-end tests.
  ([#212](https://github.com/ohstr/nmilat/pull/212))
- Adds fake-Hub orders on a fake chain (mempool, confirmation tiers, reorgs,
  evictions), minting once per payment with splitting at `max_millis`,
  notices on every change, reservations released by the clock, retirable and
  extra connections, a broken-signer fault and a restart that keeps order
  records; conformance cases CT-ORD-04..16, CT-ORDC-01..10, CT-LK-14 and
  CT-OSL-01..05, with the optional `OrderConnectionAdder` and `Restarter`
  interfaces. ([#211](https://github.com/ohstr/nmilat/pull/211))
- Adds the end-to-end order tests E-ORD-01..09 and E-ORDC-01..02, and the
  fuzz targets F-CO-1 and F-ORD-1.
  ([#211](https://github.com/ohstr/nmilat/pull/211))
- Adds `nip57.Invoice.Timestamp`, `Expiry` (3600 when the invoice has no `x`
  field, `DefaultBolt11Expiry`) and `ExpiresAt`.
  ([#210](https://github.com/ohstr/nmilat/pull/210))
- Adds `nipCO.CheckOrderInvoice`, `nip57test.SignedBolt11` (a signed BOLT11
  invoice builder), `nipcash.ErrRedeemAmountOnAmountInvoice`,
  `ErrPubkeyNotLowerHex` and `ErrAttestationKey`.
  ([#208](https://github.com/ohstr/nmilat/pull/208))
- Adds `nipcash.ErrBech32m` and `nipIC.ErrNConnectionBech32m`.
  ([#206](https://github.com/ohstr/nmilat/pull/206))
- Adds to the fake Hub a `cash_redeem` fake payee and `cash_transfer` splits
  on its bill ledger, `FaultStallCommit`, and the optional conformance
  interfaces `Payee` and `SliceRemover`; CT-WD-09 now runs every leg,
  CT-WD-22 and E-WD-14 are added.
  ([#206](https://github.com/ohstr/nmilat/pull/206))
- Adds `vectors` (`Files`, `Build`, `WriteAll`, `Version`), which builds
  every cross-implementation vector file (37) from nmilat with deterministic
  inputs, and `cmd/export-vectors`; the files are committed under
  `testdata/vectors` with a drift test and a known-answer read-back test.
  ([#205](https://github.com/ohstr/nmilat/pull/205))
- Adds tests freezing the remaining known-answer negatives (relay hash,
  reply key, params hash, the connection key never produced) and member
  connection keys. ([#205](https://github.com/ohstr/nmilat/pull/205))
- Adds `relay.Reject(prefix, err)`, so an event validator chooses its NIP-01
  prefix (other errors stay `invalid:`); `nipCP/relayreg` now refuses a bare
  kind 9525 as `blocked:`.
  ([#204](https://github.com/ohstr/nmilat/pull/204))
- Adds `relay.WithRecipientOnlyKinds` (`SessionConfig.RecipientOnlyKinds`):
  an opt-in rule serving a kind only to a connection AUTHed as one of its
  `p`, on REQ, COUNT, NEG-OPEN and the live tail, and
  `SessionHandler.QueryHandler` for POST /query.
  ([#204](https://github.com/ohstr/nmilat/pull/204))
- Adds `nipcw/relayreg`, which refuses a follow set dated more than 15
  minutes ahead, and `nipcw.FollowSetMaxFuture`.
  ([#204](https://github.com/ohstr/nmilat/pull/204))
- Adds `nip46/signeruri.Dial`, which opens a signer from a `bunker://`,
  `bunker+unix://` or `unix://` URI or a bare absolute socket path and
  refuses any other scheme, `nipLS.DenyAll`, and `nipLS.ErrMessageTooLarge`
  (the client no longer sends a request line over 1 MiB).
  ([#203](https://github.com/ohstr/nmilat/pull/203))
- Adds conformance cases CT-WD-01..21 and 23..26, CT-STAT-13 and CT-ENV-16,
  with the optional `WithdrawChain` and `WithdrawToggle` interfaces; the
  suite mints its bills with `mint_cash` over the Cash Hub connection.
  ([#202](https://github.com/ohstr/nmilat/pull/202))
- Adds a bill ledger to the fake Hub (`ServeWithdrawals`: `mint_cash`,
  `cash_status` with the withdraw object, quotes, withdrawals, cancels,
  lookups, re-mints, notices), a fake chain (blocks, mempool, reorgs,
  evictions, withheld transactions) and the payout faults, with end-to-end
  client tests E-WD-02..06 and 08..11.
  ([#202](https://github.com/ohstr/nmilat/pull/202))
- Adds `MintCashResult.MarshalJSON`/`UnmarshalJSON`, so a Hub encodes its
  `mint_cash` reply from the result type, and `Expiry.IsZero`.
  ([#201](https://github.com/ohstr/nmilat/pull/201))
- Adds `nipcw/client.(*Client).Info`, the `get_info` read at the online
  check. ([#201](https://github.com/ohstr/nmilat/pull/201))
- Adds `nipCP` withdraw entries: `Entry.Address`, `Quote` and `Receipt`, the
  `remint` secret role, pending `cash_withdraw` entries, `ApplyWithdraw`
  (receipt checked first, source spent, re-minted bill on `cancelled`,
  `done` on `final` or `cancelled`) and `WithdrawDoneRewrite`; and
  `DialRelaysWith(DialConfig{AuthWait})`.
  ([#201](https://github.com/ohstr/nmilat/pull/201))
- Adds `relay/client.WithPublishFilter`, `WithAfterPublish`,
  `PublishFilterFrom` and `AfterPublishFrom`; the `nipcash/client` hooks of
  the same names delegate to them.
  ([#201](https://github.com/ohstr/nmilat/pull/201))
- Adds `niplsp`, the NIP-LSP Lightning Service Provider Announcement (kind
  10735): `FromLSPSInfo` builds the fields from an LSP's LSPS0/1/2/5 answers
  in either unit dialect, `NewAnnouncement`, `SigningPayload` and `Sign` (an
  injected node signer that only signs kind 10735), `ParseAnnouncement` in
  the NIP's client order (`*AnnouncementError{Step}`), `Seen`/`Newest`
  (newest version per node, no fallback), `CheckLive` against the node's
  live LSPS answers, and `NodeIDs`.
  ([#199](https://github.com/ohstr/nmilat/pull/199))
- Adds `nipcash/listing` (NIP-CASH §Listing Offers): kind-31190 listings
  built from signed `cashorder1`/`circlehub1` records (`NewListing`,
  `ClosedListing`, `NewListingDeletion`); the wallet's checks before showing
  (`ParseListing`, `ListingError`), `VersionBook` and the at-tap outcomes
  (`CheckTap`, `TapCache`); operator offers (kind 11191), recommendations
  (kind 31191), reports and curator sets; signals and ranking from trusted
  authors only; `SanitizeText`; and the trusted and global fetch filters.
  ([#198](https://github.com/ohstr/nmilat/pull/198))
- Adds listings to the fake Hub (`cashtest.Lister`) and conformance cases
  CT-LIST-01..06. ([#198](https://github.com/ohstr/nmilat/pull/198))
- Adds the dependency `github.com/rivo/uniseg`, used to count user-perceived
  characters. ([#198](https://github.com/ohstr/nmilat/pull/198))
- Adds on-chain withdrawals to `nipcash` (`cash_withdraw_quote`,
  `cash_withdraw`): witness address checks (`DecodeWitnessAddress`,
  `NormalizeAddress`, `ClassifyAddress`, `WithdrawHRPs`), `WithdrawFee`, the
  kind-23187 quote, call requests, the kind-23188 receipt, payouts, cancels,
  `Detail`/`ParseDetail`, and the Hub's decisions (`CheckQuoteRequest`,
  `CheckWithdrawPayout`, `PayoutStateOf`, `CancelDue`, `RemintExpiry`,
  `WithdrawRecordRetained`, `WithdrawBlocksHubDelete`).
  ([#197](https://github.com/ohstr/nmilat/pull/197))
- Adds the withdrawal flows to `nipcash/client`: `WithdrawTerms`,
  `QuoteWithdraw`, `Withdraw` (written ahead, receipt verified),
  `LookupWithdraw`, `CancelWithdraw`, `WatchPayout`, `PayoutFinished`, the
  batch forms, `WithWithdrawHRP`, `ReplaceOptions.Quote`, and the
  `cash_withdraw` rules of `Recover` and `Replace`. `PendingCall` gains
  `Quote`, `Address`, `HoldsUntil`, `Receipt` and `LastDetail`.
  ([#197](https://github.com/ohstr/nmilat/pull/197))
- Adds conformance cases CT-WD-01 (wire part) and CT-SCOPE-05, and
  `cashtest.WithWithdrawals`.
  ([#197](https://github.com/ohstr/nmilat/pull/197))
- Adds `nip59.WrapWithKey`, `UnwrapWithKey` and `WithWrapExpiration`: the NIP-59
  receive path, and wraps under a key the caller keeps.
  ([#196](https://github.com/ohstr/nmilat/pull/196))
- Adds `nipCP` (Cash Pocket): kind-9525 entries with a strict rumor codec,
  sealed and opened under a recorded wrap key; resolution per `w`, balance
  per chain and `MapStatus`; pending entries written ahead for `mint_cash`,
  `cash_transfer`, `cash_consolidate` and `cash_order`; `ApplyResult`,
  `ApplyLookup` and `ApplyOrder`; `PocketJournal`, a `client.Journal` that
  reads entries back from two or more relays; sync, republish, NIP-40 wrap
  expiration and prune helpers; and `nipCP/relayreg`, which refuses bare
  kind-9525 events. ([#196](https://github.com/ohstr/nmilat/pull/196))
- Adds `nipcash.ChainForHRP` (the chain a bill's token prefix names) and
  `client.DerivedParentRecorder`, an optional Journal extension that
  `Replace` tells the replaced call's request_id.
  ([#196](https://github.com/ohstr/nmilat/pull/196))
- Adds `nipCO` (Cash Orders): the `cashorder1` codec
  (`EncodeOrderConnection`, `DecodeOrderConnection`, `SignOrderConnection`,
  `OrderConnection.Verify`), selling terms (`ParseOrderInfo`,
  `Terms.Validate`, `SellsNothing`), `OrderFee`, `ConfirmationsFor`,
  `MintSplit`, the `cash_order` request and order shapes, the buyer key
  (`BuyerKeyFromSecret`, `NewCashOrderTarget`, `NewBuyerKey`, `NewNotify`),
  the kind-23189 order notice, and the Hub's decisions (`CheckPlacement`,
  `OrderStateAt`, `OrderRetained`, `OrderBlocksHubDelete`,
  `OrderHubDeleteAt`, `UnpaidOnchainDeleteAt`, `UsedMarkDroppable`).
  ([#195](https://github.com/ohstr/nmilat/pull/195))
- Adds `nipCO/client`: `Connect` (verifies the string, runs the online
  check, verifies the announcement), `OrderClient.Terms`, `Place` (written
  ahead, private transport only, `TermsChanged`, on-chain address check),
  `Lookup`, `Watch`, `Finished` and `EndedUnpaid`.
  ([#195](https://github.com/ohstr/nmilat/pull/195))
- Adds `nipcash.MethodCashOrder`, `client.BatchSession.CashOrderMany`,
  `LookupOrderMany` and `BillCallsUsable`, and `client.ItemOutcome.Accepted`
  and `Pending`; `Recover` and `Replace` use the new constant and a fresh
  notify on re-placement. ([#195](https://github.com/ohstr/nmilat/pull/195))
- Adds an Order connection and order book to the fake Hub, and conformance
  cases CT-ORD-01, -02, -03 and -14 for Hubs implementing `OrderPayer`.
  ([#195](https://github.com/ohstr/nmilat/pull/195))
- Adds `nipCZ` (Cash Zaps): `ParseCashZap` and `ParseCashZapInfo` for kinds
  9522 and 10522, `VerifyClient` (rules 1-8) and `VerifyWallet` (rules 1-6,
  8 and 9, with a caller-supplied rule 9 check and the
  `CheckStatus`/`CheckInPlaceTransfer` helpers), and a `Verdict` with
  `Countable`, `ExpiresAtTrusted`, `CreditSender`, `MinterUnlisted`,
  `MayContactMinter` and `Received`.
  ([#194](https://github.com/ohstr/nmilat/pull/194))
- Adds `nipCZ.ZapTransfer`, which builds a zap's `cash_transfer` and refuses
  a ConnectionKey-bound bill on a target (`ErrZapNeedsPubkeyBill`),
  `NewCashZap`, which copies `issued_at`/`expires_at` exactly and never
  writes `expiration`, `PlanSend`/`PlanSendCandidates`, `Count`,
  `SelectCopy`/`SelectTrustedCopy`, `ZapTargetFor`, `ZapContextFromEvent`,
  `Chains`/`ChainForHRP`, `SameMinter`, `IncomingFilter`, `SentFilters`,
  `PublishRelays` and `IncomingRelays`.
  ([#194](https://github.com/ohstr/nmilat/pull/194))
- Adds `nipCZ/relayreg`, an opt-in relay validator for kind 9522 that
  applies rules 1-6. ([#194](https://github.com/ohstr/nmilat/pull/194))
- Adds `nipAZ.ServerPolicy` and `ReceiptPolicy`, so a server applies every
  NIP-AZ request rule (advertised chain, callback amount, own endpoint,
  trusted `ia`) and every receipt rule including rule 10, `SupportsAltZap`,
  `AltZapRequest.EconomicSender`, `Endpoint`, `AltZapReceipt.Amount` and
  `NormalizeEndpoint`. ([#193](https://github.com/ohstr/nmilat/pull/193))
- Adds `ia` to alt-zap requests and receipts, `nipAZ.ResolveIdentity` and
  `SweepAllowed` (with `HeldValue`), which resolve a ConnectionKey or
  release held value only through the request's IA's current attestation,
  and `AltZapReceiptParams.LinkAuthorized`, without which `r`/`R` are
  refused (`ErrLinkNotAuthorized`).
  ([#193](https://github.com/ohstr/nmilat/pull/193))
- Adds `nip57test.Spec.IA`, `TestIA` and `AltZapLnurl`.
  ([#193](https://github.com/ohstr/nmilat/pull/193))
- Adds the NIP-CW join building blocks: `nipcw.Terms`, `TermsTag`,
  `ParseTerms`, `RenewalAtLeast`, `ValidRenewal`,
  `CreateCircleWalletParams.Prepare` with a single-use `*PreparedJoin`
  (`Item`, `ParseResult`, `ErrPreparedJoinUsed`), `CircleWalletDetails` and
  `EncryptDetails`, `JoinLookupParams` and
  `nipcash/client.(*BatchSession).LookupCircleJoin`.
  ([#192](https://github.com/ohstr/nmilat/pull/192))
- Adds NIP-CW Hub helpers: `VerifyJoinProof`,
  `ParseCreateCircleWalletRequest`, `JoinProofError`,
  `JoinProofRetainUntil`, `NewMemberRequest`, `VerifyMemberRequest`,
  `MemberRequestRetainUntil`, `BudgetPeriodStart`, `BudgetPeriod`,
  `FeeSkim`, `RoutingFeeReserve`, `MaxCost`, `ResolveFollowSet`,
  `EncryptHiddenMember`, `DecryptHiddenMember`, `SealHiddenMembers`, and
  every NIP-CW join and wallet error code.
  ([#192](https://github.com/ohstr/nmilat/pull/192))
- Adds `nipcw.SignCircleHubConnection`, `CircleHubConnection.Verify` and
  `Fields`, `nipcw.WalletConnection`, and in `nipcw/client` `LookupJoin`,
  `DeleteWallet`, `ParseRotationNotification`, `WalletGuard`,
  `JoinIndeterminateError`, `PendingJoin`, `JoinJournal`,
  `JoinResult.Wallet`, `Client.Verified`, `ErrNoAnnouncement`,
  `ErrNotAccepting`, `ErrRotationWarningRequired`, `ErrHubNonConforming` and
  `ErrOverHubLimit`. ([#192](https://github.com/ohstr/nmilat/pull/192))
- Adds an optional Circle Wallet Hub connection to the fake Hub
  (`cashtest.WithCircleHub`).
  ([#192](https://github.com/ohstr/nmilat/pull/192))
- Adds the client's write-ahead journal for calls carrying a `request_id`:
  `nipcash.PendingCall`, `client.Journal` (and the optional
  `AcceptedRecorder`), envelopes packed one at a time with `not_after` =
  now + `DeadlineWindow(H)` and each call written ahead with that exact
  deadline before publishing, `FinalError` for `BAD_REQUEST`,
  `IndeterminateError` for any other answer short of a result,
  `ItemOutcome.RequestID` and `NeedsLookup`, `Recover` and `Replace` (derived
  ids, at most 3 re-placements, a mint or order only with confirmation), and
  `ErrNoJournal`, `ErrNotSent`, `ErrDeadlineTooClose`, `ErrDuplicateRequestID`
  and `ErrReplaceNeedsConfirmation`.
  ([#191](https://github.com/ohstr/nmilat/pull/191))
- Adds minting over a `cashhub1` string (`ConnectCashHub`, `HubInfo`,
  `HubClient.Info` with the online check, `HubClient.MintCash` with a
  journal, `LookupMint`), lookups of transfers and consolidates
  (`Client.LookupTransfer`, `LookupConsolidate`; a bearer call is looked up
  without its secret first), `ConnectCashSlice`, `CashSecretCredential`,
  `ConnectZap`, `SelectVerifiedCopy` and `Bill.Ordered`.
  ([#191](https://github.com/ohstr/nmilat/pull/191))
- Adds proof of work to the client: a session reads each relay's NIP-11
  minimum, mines a request to the highest one with its `expiration` at
  `not_after`, re-mines the same request on a `pow:` refusal, and reports a
  minimum over `WithPowBudget` or mining not done in time as `ErrPowTooHard`
  (an `ErrNotSent`). Sessions take `SessionOption`s.
  ([#191](https://github.com/ohstr/nmilat/pull/191))
- Adds a proof-of-work relay double to `internal/cashtest`
  (`Relay.SetMinPow`, NIP-11, the `expiration` window, id dedup) and the
  CT-POW-01..03 conformance cases.
  ([#191](https://github.com/ohstr/nmilat/pull/191))
- Adds `request_id` to `mint_cash`, `cash_transfer` and `cash_consolidate`
  (`RequestID`, filled with a fresh id by the SDK until the journal lands),
  `zap_target` and `remainder_commitment` to `cash_transfer`, a zap bill's
  `issued_at` and `expires_at`, `cash_lookup` (`CashLookupParams`,
  `ParseCashLookupResult`), tombstones (`Tombstone`, `ParseTombstone`,
  `*BillGone`, `TombstoneReason`), `RedeemFee`, `RequiredInvoiceAmount`,
  cash secrets and commitments, `OrderedResult`, the `withdraw` terms of
  `cash_status`, and `OwnRow` and `Claimable`.
  ([#190](https://github.com/ohstr/nmilat/pull/190))
- Adds the Hub-side serving helpers of the private transport:
  `transport.ClassifyEnvelope` (E1 to E6 in order), `ItemShapeOK`,
  `BearerShapeOK`, `VerifyBillProofAny`, `WrapResponseChunk`,
  `PaddedReplySize`, and `nipcash.ServingMethod`, `ReuseCheckApplies`,
  `RequiresRequestID` and `AnswersWithTombstone`.
  ([#189](https://github.com/ohstr/nmilat/pull/189))
- Adds proof of work for private requests: `transport.WrapRequestPow` mines
  a request with an `expiration` equal to its `not_after`, `Remine` re-mines
  the same envelope, `ParsePowRejection` reads a `pow:` refusal and
  `RelayPowMinimum` reads a relay's NIP-11 minimum.
  ([#189](https://github.com/ohstr/nmilat/pull/189))
- Adds `network` and `chain` to `get_info` results (`GetInfoResult.Chain`,
  `ValidNetwork`), `accepting`, `fee_reserve`, `dust_threshold_millis`,
  `rotation_delay_secs` and `per_wallet_max_receive_millis` to a Circle
  Wallet Hub's terms (`CircleWalletInfo.MissingAlways` names absent keys),
  receive budgets, the NIP-CASH and NIP-CW error codes and rotation
  notifications, `delete_circle_wallet`, and `nip47.WithExpiration`,
  `RequestExpiration` and `NWCClient.CallWithOptions`.
  ([#188](https://github.com/ohstr/nmilat/pull/188))
- Adds `request_id` and lookup helpers for NIP-CASH: `ValidRequestID`,
  `NewRequestID`, `DerivedRequestID`, `CheckDeadline`, `CommitAllowed`,
  `DeadlineWindow`, `DecideLookup`, `RecordRetained`, `TombstoneAnswerable`,
  `SliceLive`, the `ErrCode*` constants, `CodeRetryable` and `FinalCode`.
  ([#187](https://github.com/ohstr/nmilat/pull/187))
- Adds `max_request_age` to the Hub announcement (`Limits.MaxRequestAgeSeconds`,
  `AnnouncedLimits.MaxRequestAge`) and `Announcement.ValidateForBillCalls`,
  which reports a Hub without one of at least 60 seconds as unusable.
  ([#187](https://github.com/ohstr/nmilat/pull/187))
- Adds shared claim-proof helpers for NIP-CASH Hubs: `VerifyClaimProof`
  (the exact tag schema per method and slice, bound values, freshness,
  consumed ids, the connection_key attestation checks, `RESTRICTED` for an
  untrusted IA), `VerifySliceAttestation`, `ItemSignerMatchesClaim`,
  `FreshAt`, `ClaimRetainUntil`, `SourcesHash` and `NewIdentityHashOf`.
  ([#186](https://github.com/ohstr/nmilat/pull/186))
- Adds canonical NIP-IC ConnectionKeys for x, discord, email and dns (IDNA
  A-labels, refusing any input the mapping would change), keys taken byte for
  byte from an attestation for other platforms (`ConnectionKeyFromAttested`,
  `nipcash.ConnectionKeyFromAttestation`, `ConnectionKeyFromNConnection`),
  and the NIP-IC verification model (`nipIC.Evaluate`).
  ([#185](https://github.com/ohstr/nmilat/pull/185))
- Adds `nipIC.NewIdentityConnection`, which links a user's `35521` to one or
  more attestations by reference, `auth_type` support for attestations
  verified by other means, and `Attestation.Reverify` for cross-IA
  re-verification of `public_post` evidence.
  ([#185](https://github.com/ohstr/nmilat/pull/185))
- Adds the `cash-mint:v1` and `cash-zap:v1` mint signatures with
  `relay_hash` (`MintPayloadV1`, `ZapPayloadV1`, `ProvenancePayload`,
  `VerifyZapProvenance`, `MinterXOnly`), `nipcash.SignToken`, which refuses
  to issue a token whose signature does not recover to the node key,
  `LocalMintSigner`, `IsReservedSignText`, and the online check
  (`OnlineCheck`, `HubIdentityEqual`).
  ([#184](https://github.com/ohstr/nmilat/pull/184))
- Adds signed connection strings: `cashhub1` carries the Hub identity, a
  node signature and the chain (TLV 4, 5, 6), with `SignCashHubConnection`,
  `CashHubConnection.Verify`, `ConnSignedText`, `SignConnection`,
  `VerifyConnection`, `LocalChainSigner`, `RelayHash` and `QRForm`.
  ([#183](https://github.com/ohstr/nmilat/pull/183))
- Adds TLV 8, the token's signature scheme (`Token.Scheme`), and
  `Token.UsableRelays`.
  ([#183](https://github.com/ohstr/nmilat/pull/183))
- Adds the `nipcash/conformance` suite, which runs the spec's Hub scenarios
  against any Hub through `HubUnderTest` and optional `Clocked`, `Faulty`,
  `Admin`, `HubDeleter`, `OrderPayer`, `PowRelay` and `Lister` adapters; this
  first set covers envelope serving and connection scope (CT-ENV-01, 02, 03,
  10, 11, 14, 15 and CT-SCOPE-03).
  ([#182](https://github.com/ohstr/nmilat/pull/182))
- Adds `transport.DecodeResponseChunk`, which refuses a reply chunk only when
  it is unusable and reports ids never sent or answered twice, and the error
  `detail` field (`ResultError.Detail`). A reply that omits every item now
  carries `"results": []`.
  ([#181](https://github.com/ohstr/nmilat/pull/181))
- Adds `nip44.MessageKeys` and `nip44.OpenWithKeys`: one payload's message
  keys, MAC checked, which open that payload and no other. The package now
  tests against the official NIP-44 v2 vectors.
  ([#207](https://github.com/ohstr/nmilat/pull/207))
- Adds the `nip44_message_keys` method to `nipLS` and `nip46/bunker` (server
  and client), and `nipLS.ResultPolicy`, which sees an encrypt or decrypt
  result before it is returned.
  ([#207](https://github.com/ohstr/nmilat/pull/207))
- Adds `nip59.Seal`, `WrapSeal` and `Unwrap`, which take the key as a signer
  and never open a seal from a sender not asked for. `Prove` and
  `VerifyProof` show one seal's content to a third party without any key.
  ([#207](https://github.com/ohstr/nmilat/pull/207))
- Adds NIP-17 recipient, subject and reply tags, `ParseChatMessage`, and
  kind 10050 DM relay lists.
  ([#207](https://github.com/ohstr/nmilat/pull/207))
- Adds builders for NIP-25 reactions, NIP-10 replies, NIP-18 reposts and
  NIP-09 deletion requests, and NIP-02/NIP-51 list edits that keep every
  other item.
  ([#207](https://github.com/ohstr/nmilat/pull/207))
- Adds `WithSearchBatch` and `WithSearchInFlight`: `SearchIndexer` keeps
  several batches enqueued with a `search.Enqueuer` engine and acknowledges
  them in order, only past what was applied. A large relay's search can then
  keep up, since the engine merges queued batches.
  ([#179](https://github.com/ohstr/nmilat/pull/179))
- Adds `EventStore.EachProfile`: every author's newest profile, read through
  the kind index instead of every stored event, so a search rebuild reads
  only profiles.
  ([#179](https://github.com/ohstr/nmilat/pull/179))
- Adds `EventStore.IndexStatus`: each derived index's state (`ready`,
  `behind`, `needs_rebuild`, ...), the action that brings it current and why,
  its exact entries, and a paused rebuild's position.
  ([#173](https://github.com/ohstr/nmilat/pull/173))
- Adds `RebuildIndex` (resumes where a restart stopped it), `CatchUpIndex`,
  `RescoreIndex`, `ClearIndex`, `ReapIndex` and `RunIndexAction`;
  `WithEventStoreIndexMaintenance(MaintenanceManual)` leaves running them to
  the embedder.
  ([#173](https://github.com/ohstr/nmilat/pull/173))
- Adds durable feeds (`WithEventStoreFeeds`, `FeedChanges`, `AckFeed`,
  `FeedStatus`) and `SearchIndexer`, which keeps a search engine in step with
  stored profiles, their deletions and score changes.
  ([#173](https://github.com/ohstr/nmilat/pull/173))
- Adds `StoreObserver.IndexBatchCommitted`, one call per index operation
  transaction.
  ([#173](https://github.com/ohstr/nmilat/pull/173))
- Adds the `zap-ranking` cache action: receivers ranked by counted zaps in
  one coin over one window, as a relay-signed kind 25521 tagged `window`,
  `coin` and `limit`. `SessionConfig.Zaps` lists the windows (`all` is all
  time) and coins a request may name; the first of each is the default and
  `limit` is capped at `MaxLimit`. Lists are computed together in one pass on
  first use, reused for `Refresh`, served stale while one refresh runs,
  skipped when nothing was written and dropped when idle.
  ([#167](https://github.com/ohstr/nmilat/pull/167))
- Adds each zap's coin (`Coin`: sats from a `bc`/`tb`/`tbs`/`bcrt` invoice,
  loki from `fc`, else a 9521's `chain` tag) to the zap and social indexes.
  A store from before rebuilds both once in the background on open.
  ([#167](https://github.com/ohstr/nmilat/pull/167))
- Adds `zap_amt_recd_<coin>`/`zap_amt_sent_<coin>` NIP-85 tags for each
  configured coin besides sats with an amount, and `ProfileStats.AmtRecd`/
  `AmtSent` per coin.
  ([#167](https://github.com/ohstr/nmilat/pull/167))
- Adds exact counters of stored events per kind and distinct authors, kept
  in the same transaction as the indexes (`EventStore.Counts`); an older
  store is counted in the background in short write transactions.
  ([#167](https://github.com/ohstr/nmilat/pull/167))
- Adds `RelayDiscovery` and `EventStore.PublishSelfSigned`, to publish the
  relay's own NIP-66 kind 30166 (counts, zap totals, supported NIPs,
  requirements), stored only when something changed.
  ([#167](https://github.com/ohstr/nmilat/pull/167))
- Adds `EventStore.ZapTotals`, `TopZapped` (one window and coin, uncached)
  and `Commits`, `utils.ParseDuration` (`s`, `m`, `h`, `d`, `w`, `mo`) and
  the `nip57/nip57test` package, which builds valid 9735 and 9521 receipts
  in either coin for tests.
  ([#167](https://github.com/ohstr/nmilat/pull/167))
- Adds NIP-85 user assertions: with `SessionConfig.EnableProfileStats`, a REQ
  for kind 30382 with `#d` pubkeys gets relay-signed stats (rank, followers,
  following, zaps received/sent with first and last time, verified
  NIP-05/lightning address), and NIP-11 lists 85.
  ([#165](https://github.com/ohstr/nmilat/pull/165))
- Adds followers and received zaps to a profile's search score: each sender
  adds a capped amount, scaled by their own profile score, so many throwaway
  keys or a self-run LNURL server add little. Weights live in `search.social`.
  ([#165](https://github.com/ohstr/nmilat/pull/165))
- Adds `EventStore.GetProfileStats`: a profile's zaps received and sent (count,
  amount, distinct zappers, first and last time) and follower/following counts,
  from an index kept current on every write and built once on first start.
  ([#165](https://github.com/ohstr/nmilat/pull/165))
- Adds `config.Configure`, so an embedder can set its own scoring weights.
  ([#164](https://github.com/ohstr/nmilat/pull/164))
- Adds `ProfileMetrics.CheckScore`, which rescores a profile from its stored
  check results with the current weights, without fetching anything.
  ([#164](https://github.com/ohstr/nmilat/pull/164))
- Adds `nip47.CircleWalletInfo.PerWalletMaxMillis` and `MinBudgetRenewal`, a
  Hub's per-wallet ceiling and budget-renewal floor from NIP-CW `get_info`, so a
  client can check a join before sending it.
  ([#159](https://github.com/ohstr/nmilat/pull/159))
- Adds `nipcw/client.ErrJoinNotSent`, which marks a `Join` that failed before
  anything was sent, so a caller can tell it from one whose outcome is unknown.
  ([#157](https://github.com/ohstr/nmilat/pull/157))
- Adds `nip47.CircleWalletInfo.Identity`, the node key a Hub reports in
  `get_info` when it offers NIP-CW Private Join, and the `circle_policy`
  constants `CirclePolicyAllowlist` and `CirclePolicyFollowSet`.
  ([#146](https://github.com/ohstr/nmilat/pull/146))
- Adds `nipcw/client.Client.Join`, which joins a circle over NIP-CASH's private
  transport when the Hub offers it, so the join is not visible on the shared
  Hub connection; a fallback to the standard transport is reported, never silent.
  ([#146](https://github.com/ohstr/nmilat/pull/146))
- Adds `utils.ParsePrivateKeyHex`, which decodes a hex secret key and refuses
  one outside [1, n-1].
  ([#141](https://github.com/ohstr/nmilat/pull/141))
- Adds `nip11.Metadata.Banner`, NIP-11's "banner" image URL, settable from
  config as `banner`.
  ([#132](https://github.com/ohstr/nmilat/pull/132))
- Adds `EventStore.RebuildZapIndex`, which rebuilds the zap index beside the
  live one and swaps it in atomically, so top-zapped results never go empty
  during a rebuild.
  ([#131](https://github.com/ohstr/nmilat/pull/131))
- Adds queue-depth and drop counters for profile verification and search
  indexing, `MembershipService.Count`, `GroupsService.Count`, and
  `EventStore.TaskQueueDepth`/`InflightScans`, so an embedder can graph them.
  ([#131](https://github.com/ohstr/nmilat/pull/131))
- Adds `relay.Observer` and `StoreObserver` (`WithObserver`,
  `WithEventStoreObserver`) for relay metrics: sessions, messages, EVENT
  outcomes by reason and kind class, REQ/COUNT latency, subscriptions, AUTH,
  sends and store batches, without allocating on the hot path.
  ([#131](https://github.com/ohstr/nmilat/pull/131))
- Adds `MaxAmountMillis` to `nipcw.CreateCircleWalletResponse`, the spend cap
  as the Hub resolved it, and the refusal codes `nipcw.ErrNotReady`,
  `ErrRestricted`, `ErrWalletExists` and `ErrRateLimited`, as NIP-CW now specifies.
  ([#123](https://github.com/ohstr/nmilat/pull/123))
- Adds `nipAZ.ValidateAltZapRequestForRelay`/`ValidateAltZapReceiptForRelay`,
  which `nipAZ/relayreg` now uses: relays keep receipts for real payments and
  reject only forged AltZap events, as NIP-AZ specifies.
  ([#118](https://github.com/ohstr/nmilat/pull/118))
- Adds `NWCClient.Notifications`, which delivers a wallet's NIP-47 payment
  notifications as they arrive, and `NWCClient.Done`, which tells a
  long-lived holder when to reconnect. ([#116](https://github.com/ohstr/nmilat/pull/116))
- Adds `utils.RequestPayInvoice` for any LNURL-pay callback (LUD-12 comments
  included) and `Tag`, `CommentAllowed`, `Metadata` on `utils.PayResponse`.
  ([#116](https://github.com/ohstr/nmilat/pull/116))
- Adds `nip98.NewAuthHeader`, `utils.ResolveNip05`, and hex-key
  `nip44.EncryptHex`/`DecryptHex`. ([#116](https://github.com/ohstr/nmilat/pull/116))
- Adds the invoice memo and currency prefix to `nip57.Invoice`
  (`Description`, `Currency`). ([#116](https://github.com/ohstr/nmilat/pull/116))
- Adds `WithRequestHandler`, REQ handlers run ahead of the built-in ones, and
  `Session.SendEvents`/`SendEOSE` to answer a REQ in the handler's own order, so
  an embedder can serve NIP-50 search from its own index, best match first.
  ([#111](https://github.com/ohstr/nmilat/pull/111))
- Adds `nip11.Metadata.Icon`, NIP-11's `icon` URL, settable as `nip11.icon`.
  ([#106](https://github.com/ohstr/nmilat/pull/106))
- Adds `SessionConfig.EventPolicy` and `WithEventsPolicy`, a hook that can refuse
  any event after validation, over the WebSocket and `POST /events`, so an
  embedder can enforce its own bans and kind rules.
  ([#106](https://github.com/ohstr/nmilat/pull/106))
- Adds optional pending NIP-29 join requests: with `GroupsService.SetPendingJoins(true)`
  a closed group keeps a join request without an invite code for its admins,
  who accept with kind 9000 or decline with 9001.
  ([#105](https://github.com/ohstr/nmilat/pull/105))
- Adds an optional NIP-29 moderation log: with `GroupsService.SetModerationLog(true)`
  accepted moderation events are stored and readable by the group's members
  (invites by its admins), so clients can show who did what.
  ([#104](https://github.com/ohstr/nmilat/pull/104))
- Adds kind 39003 group roles: the relay publishes one per NIP-29 group from its
  moderation policy, and republishes them when the policy changes, so clients
  know which roles exist. `GroupsService.SetRoleDescriptions` describes them.
  ([#103](https://github.com/ohstr/nmilat/pull/103))
- Adds `RelayStatus.LastEOSE`, when a relay last finished a complete sync, and
  `RelayStatus.Closed`, the relay's reason when it closed the pool's subscription.
  ([#91](https://github.com/ohstr/nmilat/pull/91))
- Adds `nip19.ParsePublicKey`, which reads an npub, nprofile or hex public key
  and rejects an nsec, a wrong length or a point off the curve.
  ([#91](https://github.com/ohstr/nmilat/pull/91))
- Adds `nip46/bunker` for NIP-46 signing over relays: `Dial`
  (`bunker://`), `NewPairing` (`nostrconnect://`) and `Resume` for clients,
  and `NewServer`, with a policy per request, for signers. ([#85](https://github.com/ohstr/nmilat/pull/85))
- Adds `nipLS` for signing through a local signer over a unix socket
  (`bunker+unix:///path.sock`): `Dial` for clients, `NewServer` to serve a
  key behind your own policy. ([#83](https://github.com/ohstr/nmilat/pull/83))
- Adds `nip46.Signer` and `nip46.LocalKey`, one signing interface that a
  local key, a socket client and a bunker client all implement.
  ([#85](https://github.com/ohstr/nmilat/pull/85))
- Adds NIP-42 authentication through a signer (`ConnectionConfig.Signer`,
  `ReadEventsFromRelayWithSigner`), so a key behind a bunker or socket
  signer can authenticate. ([#87](https://github.com/ohstr/nmilat/pull/87))
- Adds `relay/client.Pool`, a subscription per relay that survives
  reconnects, with optional resync and de-duplication. ([#87](https://github.com/ohstr/nmilat/pull/87))
- Adds `ReadAllEventsFromRelay`, which pages a query to its end without
  skipping events that share a second. ([#87](https://github.com/ohstr/nmilat/pull/87))
- Adds `nip01.LatestVersion`, which picks the winning version of a
  replaceable or addressable event the way the relay store does, ignoring
  future-dated ones. ([#87](https://github.com/ohstr/nmilat/pull/87))
- Adds `WithEventStoreScanSlotWait` (default 10s): a REQ that can't get a scan
  slot is answered `CLOSED` `error: relay busy, try again later` instead of
  never. ([#81](https://github.com/ohstr/nmilat/pull/81))
- Adds `WithEventStoreMaxConcurrentScans` (default 2 × NumCPU), so read load
  can't starve writes of disk I/O. ([#78](https://github.com/ohstr/nmilat/pull/78))
- Adds `nipcash/transport`, NIP-CASH's private transport: many requests in one
  padded, encrypted envelope, with per-item proofs and a node-signed Hub
  announcement. ([#39](https://github.com/ohstr/nmilat/pull/39))
- Adds the batch API on `nipcash/client` (`RedeemMany`, `StatusMany`,
  `TransferMany`, `ConsolidateMany`), which acts on many bills in one relay
  event so their number and timing stay private. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Adds `nipcash/client.BatchSession`, which fetches, verifies and caches a
  Hub's announcement. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Adds `ItemOutcome` and `SafeToResend`, which tell a refused batch item from
  one the Hub said nothing about. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Accepts a Hub reply split across several events (`seq`, `total`). ([#43](https://github.com/ohstr/nmilat/pull/43))
- Adds `cash_status`'s `scope` (`all` or `mine`, defaulting to `mine` on the
  private transport) and an `AttestationEvent`, so a connection-key recipient
  can read their own row. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Adds `mint_cash`'s optional `idempotency_key`, so a retried mint doesn't fund
  a second wallet. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Adds `GetInfoResult.PrivateMethods`, naming the bill methods a Hub serves only
  over the private transport. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Adds `BillState`, `CashStatus` and `CashStatusResult`; the `list_recipients`
  names stay as deprecated aliases for one release. ([#39](https://github.com/ohstr/nmilat/pull/39))
- Exports `nip01.Event.Serialize`, so a remote signer can produce the exact
  signed preimage. ([#39](https://github.com/ohstr/nmilat/pull/39))
- Adds `nip29` for NIP-29 relay-based groups, with `ModerationPolicy` for a
  relay's role rules. ([#38](https://github.com/ohstr/nmilat/pull/38))
- Hosts NIP-29 groups on the relay: create and delete, members and roles, join
  and leave, metadata, invites, pins and moderator delete. New groups are
  private and closed. ([#56](https://github.com/ohstr/nmilat/pull/56))
- Adds NIP-29 subgroups: a group can name a parent, cycles and cross-group
  admin are checked, and deleting a parent makes its children roots. ([#70](https://github.com/ohstr/nmilat/pull/70))
- Warns at startup when NIP-29 groups are hosted without `nip11.url`, which
  makes private groups invisible even to their creator. ([#72](https://github.com/ohstr/nmilat/pull/72))
- Adds `nip53` for live streams, meeting spaces and rooms, presence and live
  chat. ([#38](https://github.com/ohstr/nmilat/pull/38))
- Adds `nipA0` for voice messages and `nip71` for video events. ([#38](https://github.com/ohstr/nmilat/pull/38))
- Adds the huddle audio transport (`huddle/wire`, `huddle/room`,
  `huddle/wsaudio`), which relays Opus between peers without decoding it.
  ([#38](https://github.com/ohstr/nmilat/pull/38))
- Adds `Room.EvictPubkey` and `Manager.EvictPubkey`, which remove a pubkey from
  a live call. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Adds `nip86`, the Relay Management API, with `nip86.NewHandler` and CORS for
  browsers. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Adds `nip86.Config.Authorize`, an optional hook that admits a NIP-98-verified
  key `AllowedPubkeys` does not list, so a relay can route NIP-86 for its
  members without listing each as an admin; an error reply now carries `result`
  as an explicit null. ([#229](https://github.com/ohstr/nmilat/pull/229))
- Adds `relay.NewQueryHandler`, which serves `POST /query`: NIP-98-signed HTTP
  queries with REQ's access rules. ([#50](https://github.com/ohstr/nmilat/pull/50))
- Adds `relay.NewEventsHandler`, which serves `POST /events`: one signed event
  over HTTP. ([#51](https://github.com/ohstr/nmilat/pull/51))
- Adds `examples/` with runnable embedding patterns. ([#50](https://github.com/ohstr/nmilat/pull/50))
- Adds `nip98.Verify` options for an allowlist, body binding and a trailing
  slash, and `nip98.VerifyAnyPubkey`. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Adds `MethodSwitchRelays` and `MethodLogout` to `nip46`. ([#42](https://github.com/ohstr/nmilat/pull/42))
- Adds `ValidateZapRequestForRelay` and `ValidateZapReceiptForRelay`, and
  `ErrMissingDescriptionHash`. ([#35](https://github.com/ohstr/nmilat/pull/35))
- Answers NIP-42 AUTH from `relay/client.Connection` when given
  `ConnectionConfig.SigningKeyHex`. ([#59](https://github.com/ohstr/nmilat/pull/59))
- Adds `Connection.AuthState`, `AuthMessage` and `AuthSettled`, which report
  the NIP-42 outcome. ([#61](https://github.com/ohstr/nmilat/pull/61))
- Adds `ReadEventsFromRelayWithAuth`, which retries once after authenticating
  when the relay refuses the first read. ([#61](https://github.com/ohstr/nmilat/pull/61))

### Changed

- Changes `Replace` to re-place a withdrawal automatically only while its
  quote holds at `ReplaceOptions.ChainTip`, showing it (`quote_unchecked`,
  `quote_expired`) otherwise, so callers must now pass the tip; to refuse,
  before anything is written, a send that packs no call under the derived id
  or changes its target, amount, secrets or quote; and to record the derived
  id only once the call is written ahead.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Changes a batch refused whole with QUOTA_EXCEEDED or BAD_REQUEST to keep
  that refusal for every call carrying a `request_id`, repacking only items
  without one, and drops the unreachable stale-policy resend; a batch stopped
  partway by ErrNotSent gives each unsent item `OutcomeNotSent`.
  Since every spend carries a `request_id`, only reads are repacked in
  practice: a spend in such an envelope is closed refused by `Recover`, and
  its caller issues it again under a fresh `request_id`.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Changes `Replace` to check whether a withdrawal's derived id was already
  placed before checking its quote, so a live re-placement closes as
  `derived_already_placed` rather than showing `quote_expired` or
  `quote_unchecked`.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Changes `TapCache` to count a signer as passed at tap once steps 2 and 3
  succeed, an unavailable Hub included, so `CurrentByD` keeps showing that
  listing, as unavailable, beside a copy under another pubkey.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Changes `textsan`, and so `listing.SanitizeText` and NIP-LSP text, to strip
  every format character (Unicode Cf: U+206A-U+206F, U+FFF9-U+FFFB and the
  tag characters included), keeping a zero-width joiner only between two
  emoji, and to cap the result at 32 bytes per allowed character.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Changes `ItemOutcome.SafeToResend` to answer true only for BAD_REQUEST or
  a definitive code on a call without a `request_id`, never for INTERNAL or
  REQUEST_REUSED; `RedeemOptions.AutoAmount` with no `Payee` to read the payee
  from the invoice's signature; and `CashRedeem` to return
  `*nipcash.BillGone` for a destroyed bill.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Changes `VersionBook` to show, when several pubkeys list one `d`, only the
  listing whose signer passed at tap, or none; `listing.SanitizeText` to strip
  control characters; and `textsan` to strip U+200E, U+200F, U+061C, U+00AD,
  U+180E and U+2061-U+2064.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Changes `PocketJournal.DerivedEntryExists` to return `ErrTooFewRelays`
  rather than report an absence when fewer than `MinStored` relays answered.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Changes a batch refused whole with QUOTA_EXCEEDED or BAD_REQUEST to
  re-read the announcement and, when the policy changed, repack and resend
  once; `TransferMany` to generate the remainder target of a partial cash
  split; `State` to read NOT_FOUND as `BillNoSlice` and a widened scope-mine
  answer as indeterminate; and `cashtest.Relay.Duplicate` to reach Hub
  listeners too. ([#215](https://github.com/ohstr/nmilat/pull/215))
- Changes `ConsolidateMany` to refuse an item over `max_consolidate_sources`
  or one that fits no envelope before anything is written, and
  `CreateCircleWalletMany` to give every join an outcome, building spilled
  joins in a later envelope.
  ([#213](https://github.com/ohstr/nmilat/pull/213))
- Changes `nipCO.CheckOrderInvoice` to refuse an order invoice that does not
  expire at `pay_by`. ([#210](https://github.com/ohstr/nmilat/pull/210))
- Changes `nipCO/client.Place` to refuse a Lightning reply whose invoice is
  not a BOLT11 invoice for exactly `pay_millis`; the fake Hub answers orders
  with a signed invoice for exactly `pay_millis`, expiring at `pay_by`.
  ([#208](https://github.com/ohstr/nmilat/pull/208))
- Changes `CashRedeemParams.Request` to refuse an `amount` with an invoice
  that carries one, and a pubkey recipient, target or `ia_pubkey` that is
  not 64 lowercase hex. ([#208](https://github.com/ohstr/nmilat/pull/208))
- Changes `nipIC.ConnectionKeyFromAttested` to refuse a non-canonical user
  id on a platform with a canonical form, and
  `nipcash.ConnectionKeyFromAttestation` to refuse an attestation whose `d`
  is not the ConnectionKey of its platform and user id.
  ([#208](https://github.com/ohstr/nmilat/pull/208))
- Changes `nip44.json` to follow NIP-44 v2's extended length prefix: the
  pinned upstream file's `invalid.encrypt_msg_lengths` rows of 65536 bytes and
  more are listed as `superseded`, the NIP's extended length prefix vectors
  (65535, 65536, 65537) are added as `extendedLengthPrefix`, and
  `nmilatDivergences` is gone. ([#208](https://github.com/ohstr/nmilat/pull/208))
- Changes `NormalizeAddress` to report a base58check address as
  `ErrAddressNoWitness` rather than a mixed-case error, and
  `WithdrawTerms.Validate` to refuse an empty `script_types`.
  ([#206](https://github.com/ohstr/nmilat/pull/206))
- Changes the fake Hub to answer a commit stalled past its deadline
  `BAD_REQUEST`, as NIP-CASH requires, instead of dropping the reply.
  ([#206](https://github.com/ohstr/nmilat/pull/206))
- Changes `relay.ClassifyKind` to count kinds 9522 and 9523 as zaps.
  ([#204](https://github.com/ohstr/nmilat/pull/204))
- Changes the Hub-side encoding of `expires_at` in `mint_cash`,
  `cash_consolidate` and `cash_status` results: omitted for a bill that
  never expires (clients still read absent, null and 0 as never).
  ([#201](https://github.com/ohstr/nmilat/pull/201))
- Changes `HubClient.MintCash` to apply
  `WithPublishFilter`/`WithAfterPublish` to the mint's NIP-47 request.
  ([#201](https://github.com/ohstr/nmilat/pull/201))
- Changes `niplsp` and `nipcash/listing` to share one text sanitizer;
  `niplsp` now also drops an emoji variation selector left after a stripped
  check-mark look-alike. ([#201](https://github.com/ohstr/nmilat/pull/201))
- Changes `nipCO/client` to check order addresses with nipcash's witness
  address helpers, and `FinalError.Detail`/`IndeterminateError.Detail` to
  carry `ParseDetail` of the answer.
  ([#197](https://github.com/ohstr/nmilat/pull/197))
- Changes `client.Connect` and `ConnectCashSlice` to accept only a cash
  token: no NIP-47 URI form is parsed anywhere in the cash NIPs, which an
  import-graph test pins. ([#196](https://github.com/ohstr/nmilat/pull/196))
- Changes `client.Replace` to leave the write-ahead of the re-placed call to
  the send path, which writes it once with its exact deadline, `DerivedFrom`
  and `Attempt`. ([#196](https://github.com/ohstr/nmilat/pull/196))
- Changes the `nip59` seal to carry empty tags (`[]`) instead of `null`.
  ([#196](https://github.com/ohstr/nmilat/pull/196))
- Changes `nipAZ.ValidateAltZapRequest` to take a `ServerPolicy` and
  `ValidateAltZapReceipt` a `ReceiptPolicy`. `p`/`P` tags must be 64
  lowercase hex with a NIP-IC platform name and are never lowercased;
  `amount` must be a plain positive integer; an `lnurl` is accepted in
  either case, not mixed. The relay validators apply the SHOULD rules too,
  and the builders refuse to sign an event that would fail validation.
  ([#193](https://github.com/ohstr/nmilat/pull/193))
- Changes `nipcw.CircleHubConnection` to require TLV 4, 5 and 6 (`Pubkey`,
  `Relays`, `HubIdentity`, `Signature`, `Chain`); encoding refuses an
  unsigned connection. `nipcw/client.Connect` takes a `circlehub1` string
  and `Options` and runs the online check; `Join` joins only over the
  private transport, with a `JoinJournal`, and warns before joining a Hub
  without rotation. ([#192](https://github.com/ohstr/nmilat/pull/192))
- Changes `nipcw.CreateCircleWalletRequest` to carry `max_receive`, an
  unsigned `expiry` and an always-sent `budget_renewal`,
  `CreateCircleWalletResponse` to report `ConnectionPubkey`,
  `MaxReceiveMillis`, `DustThresholdMillis`, `RotationDelaySecs`,
  `MissingTerms`, `LeakedFields` and `CapsRaised`, `ConnectionKeyFromSecret`
  to refuse a secret that is not 64 lowercase hex, and
  `nipcash/client.BatchCircleJoin` to carry a `*nipcw.PreparedJoin` (`Join`)
  instead of `Params`. ([#192](https://github.com/ohstr/nmilat/pull/192))
- Changes `TransferMany`, `ConsolidateMany`, `CashTransfer` and
  `CashConsolidate` to take a `Journal` and a device id (the composites take
  them in their params); a call without a journal is refused. A session uses
  the newest validly signed announcement and treats the Hub as unusable when
  that one is rejected, never falling back to an older one; it publishes to
  the announced relays, then the bill's own. `Connect` refuses a `cashhub1`
  string and dials a token on its usable relays.
  ([#191](https://github.com/ohstr/nmilat/pull/191))
- Changes `cash_status` to the final wire: no `cash_secret` for a cash-mode
  bill, scope `mine` by default, a destroyed bill answered with a
  `{"status":"spent","reason":…}` tombstone for every bill method, and
  `RecipientStatus.ExpiresAt` an `Expiry` (absent, null or 0 is never).
  ([#190](https://github.com/ohstr/nmilat/pull/190))
- Changes `client.CheckClaim` to read the caller's own row from its
  credential, matching a `connection_key` row by its value and its IA, and to
  report a spent or withdrawn slice as not claimable rather than as not found.
  ([#190](https://github.com/ohstr/nmilat/pull/190))
- Changes `NewCashTarget` and `RecipientToken` to return an error, the
  `cash_consolidate` item to target `Sources[0]` and be signed by its
  credential, and a cash-mode mint recipient to a caller-generated
  `*CashTarget` commitment.
  ([#190](https://github.com/ohstr/nmilat/pull/190))
- Changes the private transport to the final serving rules: no `v` on
  envelopes or replies; `cash_lookup`, `cash_order` and the withdrawal methods
  servable; `Decode` keeps proofless, secretless and repeated items for the
  Hub's own steps and reports an over-cap consolidate as `QUOTA_EXCEEDED`;
  `CheckFreshness(e, now, H)` replaces `MaxNotAfterWindow`; padding stops at
  `max_bytes`, and replies pad to the bucket under 65535; the verification
  budget counts every signed event an item carries; an announced `max_bytes`
  over 65535 is clamped (`Announcement.EffectiveLimits`).
  ([#189](https://github.com/ohstr/nmilat/pull/189))
- Changes the kind-11190 announcement to the tag form: `inbox`, one `limit`
  per name, `r` and optional `addr` tags (`Announcement.Addrs`), an empty
  `content` and no version. `ParseAnnouncement` reads only tags and accepts an
  inbox equal to the author; `NewAnnouncement` takes optional addrs and
  parses its own event before returning it.
  ([#187](https://github.com/ohstr/nmilat/pull/187))
- Changes claim proofs to carry the `m` tag, and `s` on `cash_consolidate`,
  through `ClaimBinding`; item and bill proofs now must carry `expiration`
  equal to the envelope's `not_after`, an empty `content` and lowercase hex
  `d`, `h` and `ph`, with freshness counted in whole seconds.
  ([#186](https://github.com/ohstr/nmilat/pull/186))
- Changes `nipIC.NewConnectionKey`, `nipAZ.Connection` and
  `nipcash.ConnectionKey` to return an error and refuse non-canonical input
  and platforms without a canonical form, and `BySigningConnectionKey` to
  take the attestation only.
  ([#185](https://github.com/ohstr/nmilat/pull/185))
- Changes `nipIC.NewAttestation` to write a 90-day `expiration` unless the IA
  asks for none, and to refuse an invalid platform name or a non-canonical
  user id; `ParseAttestation`, `ParseIdentityConnection` and the
  `nconnection` codec refuse duplicate, malformed or missing fields.
  ([#185](https://github.com/ohstr/nmilat/pull/185))
- Changes `nipcash.Encode` to refuse unsigned tokens and empty, blank or
  non-ASCII relay hints, and the bech32 credential decoders to refuse a
  string that mixes upper and lower case (`ErrMixedCase`); an all-uppercase
  string decodes. A token whose TLV 5/6 come without a valid TLV 8 now has no
  provenance.
  ([#183](https://github.com/ohstr/nmilat/pull/183))
- Changes `CashHubConnection` to `{Pubkey, Relays, Secret, Label, HasLabel,
  HubIdentity, Signature, Chain}`; `EncodeCashHubConnection` refuses an
  unsigned connection and `DecodeCashHubConnection` one missing TLV 4, 5 or 6.
  ([#183](https://github.com/ohstr/nmilat/pull/183))
- Changes `ItemOutcome` so a reply split into chunks reports unanswered and
  contradicted items separately (`OutcomeUnanswered`, `OutcomeConflict`), a
  result whose `result_type` is not the item's method is never read as a
  result, and an envelope that got no reply is unanswered rather than not
  served. A `switch` over `OutcomeState` needs a default branch.
  ([#181](https://github.com/ohstr/nmilat/pull/181))
- Changes a derived index that missed writes while off to catch up from
  where it stopped instead of rebuilding. Too many deletes meanwhile, a
  layout change or writes by another program still mean a rebuild.
  ([#173](https://github.com/ohstr/nmilat/pull/173))
- Changes index operations to hold the writer at most 25 ms or 256 events
  per transaction; a swap or clear only moves a pointer and the old data is
  freed in chunks.
  ([#173](https://github.com/ohstr/nmilat/pull/173))
- Changes search indexing and profile checks to follow stored profiles
  through feeds instead of the EVENT path, so none is lost to a full queue, a
  restart or the feature being off. `search.Searcher` and `search.Writer`
  replace `search.Service`, `ServiceImpl` and `ReadOnlyService`.
  ([#173](https://github.com/ohstr/nmilat/pull/173))
- Changes the social score queue and a rescore's position to be kept in the
  store; a change to `search.social` rescores every profile.
  ([#173](https://github.com/ohstr/nmilat/pull/173))
- Removes `RebuildZapIndex`, `ReindexZaps`, `ClearZapIndex`,
  `RebuildSocialIndex`, `ClearSocialIndex`, `ZapIndexState`,
  `SocialIndexState`, `TakeSocialRescore`, `ProfileVerificationWorker.QueueJob`
  and `TotalDropped`, in favour of the calls above.
  ([#173](https://github.com/ohstr/nmilat/pull/173))
- Changes the zap and social indexes to opt-in with `WithEventStoreIndexes`:
  a store opened without one neither writes nor rebuilds it, and its readers
  return `ErrIndexDisabled`. A disabled index keeps its data, marked stale
  by the first write it misses and rebuilt in the background once enabled
  again (`ZapIndexState`/`SocialIndexState` report it).
  ([#171](https://github.com/ohstr/nmilat/pull/171))
- Changes NIP-45 COUNT: a filter of only `kinds` (or nothing) is answered
  exactly from the counters, past `max_limit`, without a scan; any other
  filter that stops at its limit answers `"approximate": true`.
  ([#167](https://github.com/ohstr/nmilat/pull/167))
- Changes the social score to count received sats only; zaps in other coins
  still count toward `zap_cnt_*` and zappers.
  ([#167](https://github.com/ohstr/nmilat/pull/167))
- Removes the `top-zapped`/`get_top_zapped` cache action,
  `GetTopZapped`/`ZapStats`, `SessionConfig.EnableTopZapped`,
  `DefaultCacheWindow`/`DefaultCacheLimit`, `WithSessionTopZapped` and
  `WithSessionCacheConfig`; use `zap-ranking` and `SessionConfig.Zaps`.
  ([#167](https://github.com/ohstr/nmilat/pull/167))
- Changes profile verification to check a profile's `lud06` LNURL when it has
  no `lud16`, so older profiles get their lightning address verified too.
  ([#165](https://github.com/ohstr/nmilat/pull/165))
- Changes profile verification to fetch only for a new or changed identifier,
  an expired result, a retry or `VerifyNow`, rate-limited per host;
  `SessionConfig.Verification` sets the TTL, backoff and limits, and the unused
  `verification_ttl_hours` and `default_predefined_bonus` settings are removed.
  ([#164](https://github.com/ohstr/nmilat/pull/164))
- Logs session open/close, AUTH, rejected events and refused WebSocket
  upgrades at debug instead of info/warn/error, and rate-limits slow/failed
  write warnings. A busy relay no longer floods its log, and clients can no
  longer write error lines at will.
  ([#148](https://github.com/ohstr/nmilat/pull/148))
- Logs a write to a client that already hung up (broken pipe, connection
  reset) at debug instead of warn/error, and only once per close. Ordinary
  disconnects no longer look like relay faults.
  ([#153](https://github.com/ohstr/nmilat/pull/153))
- Changes `create_circle_wallet` to the member-generated connection key NIP-CW
  now specifies: the client sends only `connection_pubkey` (new
  `nipcw.ConnectionKey`, bound by a `connection` tag in the identity proof) and
  builds `PairingURI` itself from the Hub's `wallet_pubkey` and `relays`. The
  Hub-generated secret is gone, so Hubs must implement the new join.
  ([#127](https://github.com/ohstr/nmilat/pull/127))
- Changes the NIP-CASH params hash (`transport.CanonicalParamsHash`) to write
  strings as RFC 8785 does, so `<`, `>`, `&`, U+2028 and U+2029 are no longer
  escaped. Params containing them hash differently from earlier releases, so
  clients and Hubs must upgrade together.
  ([#123](https://github.com/ohstr/nmilat/pull/123))
- Moves `nipAZ` to NIP-AZ's new kinds: 9520 requests, 9521 receipts, 9523
  on-behalf requests. Events of the old kinds 5520, 5521 and 5523 are no longer
  parsed, validated or indexed. ([#114](https://github.com/ohstr/nmilat/pull/114))
- Removes `nipAZ`'s kind 5522 direct payment (`NewAltZapDirectPaymentRequest`,
  `KindAltZapDirectPayment`), which NIP-AZ does not define.
  ([#114](https://github.com/ohstr/nmilat/pull/114))
- Enforces NIP-AZ's sender and settlement rules: a 9520 request must not carry
  `P`, a 9523 carries exactly one, and a receipt's chain and preimage must match
  its request and invoice. ([#116](https://github.com/ohstr/nmilat/pull/116))
- Pages each filter of a `Pool` subscription to its end on connect and on
  resync, so a relay's result cap no longer cuts it short (`PoolConfig.PageSize`).
  ([#91](https://github.com/ohstr/nmilat/pull/91))
- Moves `ReadEventsFromStore` from `relay/client` to `relay`, so `relay/client`
  and `nip46/bunker` no longer pull in bbolt and yaml. ([#91](https://github.com/ohstr/nmilat/pull/91))
- Requires Go 1.26.10 for its standard-library security fixes.
  ([#89](https://github.com/ohstr/nmilat/pull/89), [#161](https://github.com/ohstr/nmilat/pull/161))
- Sends the NIP-42 AUTH challenge on every connection, not only when
  `auth_required` is on. ([#65](https://github.com/ohstr/nmilat/pull/65))
- Returns whether the relay refused the read from
  `ReadEventsFromRelayWithAuth`. This changes its signature. ([#65](https://github.com/ohstr/nmilat/pull/65))
- Wakes live subscriptions when a write commits, instead of polling every
  50ms. ([#81](https://github.com/ohstr/nmilat/pull/81))
- Rebuilds the store's indexes on first open. This upgrade is one way, so copy
  the database first. ([#37](https://github.com/ohstr/nmilat/pull/37))
- Enforces NIP-70: an event tagged `["-"]` is accepted only from its
  authenticated author, so NIP-43 clients must AUTH first. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Serves bill methods only through the `Client` over the private transport.
  This breaks callers of the standard transport. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Requires mint provenance and removes `MintSignature` from the cash params.
  This breaks callers that set it. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Requires every credential to name a usable relay. Credentials without one
  no longer decode. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Renames the `circle_wallet` field `available_mloki` to `available_millis`.
  This breaks readers of the old name. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Changes `RekeyCashSlice` to take its new target from the caller, so the
  secret can be saved before the call. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Returns the conversation key from `transport.WrapRequest`. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Refuses an envelope that repeats an identical request, and caps an item id
  at 256 bytes. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Delivers REQ events from the bytes the scan already read, making delivery
  about a third faster. ([#50](https://github.com/ohstr/nmilat/pull/50))
- Normalizes `RegisterLetteredNIP`'s id and panics on one that can't name a
  NIP. ([#38](https://github.com/ohstr/nmilat/pull/38))
- Takes the `*GroupsService` in `NewQueryHandler`, so `POST /query` applies
  REQ's NIP-29 visibility. ([#74](https://github.com/ohstr/nmilat/pull/74))

### Removed

- Removes `transport.EnvelopeVersion`, `transport.ErrEnvelopeVersion` and the
  `Version` fields of `transport.Envelope` and `transport.ResponseEnvelope`,
  kept since the envelope lost its version on the wire; drop them from
  struct literals.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Removes `nipcw.CreateCircleWalletParams.Request`, `ParseResult` and `Item`
  (use `Prepare`), `CreateCircleWalletResponse.PairingURI`,
  `CircleHubConnection.WalletPubkey`/`RelayURLs`, the standard-transport
  `nipcw/client.(*Client).CreateCircleWallet`, `JoinOptions.RequirePrivate`,
  `JoinResult.Private`/`FallbackReason` and `ErrPrivateJoinUnavailable`.
  ([#192](https://github.com/ohstr/nmilat/pull/192))
- Removes `Client.MintCash` (use `HubClient.MintCash`).
  ([#191](https://github.com/ohstr/nmilat/pull/191))
- Removes `MintCashParams.IdempotencyKey`, `nipcash.Anyone()` and
  `ErrWrongCredentialForTarget`,
  `RecipientResult.CashSecret`, `CashStatusResult.Error`/`IsSpent`,
  `CashStatusParams.AttestationEvent`, and `nipcash.MatchClaim`,
  `MatchClaimAuto`, `NoLocalIdentity` and `ErrSpent` (use `OwnRow`,
  `Claimable` and `*BillGone`).
  ([#190](https://github.com/ohstr/nmilat/pull/190))
- Removes `nipcash.MintPayload` and the `lokicash-mint:v1` payload: a token
  signed over it carries no valid provenance.
  ([#184](https://github.com/ohstr/nmilat/pull/184))

### Fixed

- Fixes a no-chain-view `Replace` re-placing a withdrawal refused `quote_invalid`
  with the same quote; like `quote_expired` it is now shown, so its user can get a
  new quote.
  ([#227](https://github.com/ohstr/nmilat/pull/227))
- Fixes `nip98.Verify` accepting a pubkey in upper or mixed case, a second
  spelling of one key, and reading an Authorization header of any size: it
  refuses both, the header past `nip98.MaxHeader` (16 KiB), and checks every
  `payload` tag against the body, not only the first.
- Strips the full ballot-X emoji block (U+1F5F4–U+1F5F7) from untrusted text, not just the
  bold-script-X pair, so two script-X badge look-alikes can no longer survive sanitization.
  ([#221](https://github.com/ohstr/nmilat/pull/221))
- Sends `MultiPayInvoice` and `MultiPayKeysend` on to the next relay hint as
  a single call does, with the same signed request, so the wallet pays each
  item once; the first answer per item wins.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Makes `nipcash.OwnRow` with scope `all` take the caller's unclaimed row
  wherever the roster lists it, then its most recently claimed one, instead of
  the first match, now that a Hub lists claimed history rows too.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Serves a NIP-09 deletion request (kind 5) back by id and by kind once the
  relay applied it, and never applies one against another deletion request.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Indexes, once when the relay opens a store, the deletion requests earlier
  versions stored without indexes, and drops their extra copies, so they are
  served and a republish is a duplicate.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Omits `budget_renewal` from a `nipcw.CreateCircleWalletRequest` with no
  renewal, which wrote `""` and was refused as empty; `private-join.json`'s
  join params change with it.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Holds re-mining for a relay that raised its proof-of-work minimum to the
  session's `WithPowBudget`: a minimum above its difficulty is refused, and
  mining stops at its duration or the deadline less `MinPublishMargin`, not at
  `not_after`.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Dials nothing in `client.Connect`, which opened the bill's first relay
  before its provenance was checked; the session dials the relays when a bill
  method runs, after the check.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Sends an NWC call on to the next relay hint when the relays it went to stay
  silent past their share of the call's deadline, or close its subscription,
  instead of waiting on the first relay that connected (NIP-CASH §Multiple
  Relay Hints); the first answer from any of them is the call's.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Sends a private-transport envelope on to the next relay when the one that
  took it stays silent past its share of the time to `not_after`, or closes,
  instead of waiting out `not_after` there.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Refuses every call, reads and bearer `cash_redeem` included, while a Hub's
  newest announcement is rejected, and drops the older one instead of keeping
  its inbox (`Inbox()` is then empty) until a valid newer one is seen.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Matches envelope and item keys exactly: `ClassifyEnvelope` drops, and
  `Decode` refuses, an envelope repeating a key or spelling one two ways that
  differ only in case (`transport.ErrEnvelopeKeys`), and such an item is
  omitted, so E2 to E4 check the `nonce` and `not_after` the run uses.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Sorts canonical params keys by UTF-16 code units, as RFC 8785 and NIP-CASH
  §Item Proofs say, not by UTF-8 bytes. Hashes change only for keys above
  U+FFFF; `canonical-params.json` gains a case for that order.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Refuses a bolt11 invoice carrying more than one 32-byte payment hash field
  (`nip57.ErrBolt11PaymentHashRepeated`), where it kept the last: flnd keeps
  the first, so a redeem could bind a hash the Hub never pays.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Renames three vector cases whose names said the opposite of their expected
  error (`claim.json` "an unparseable expiration is refused", "a repeated
  platform tag is refused"; `proof.json` "wrong not_after").
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Gives `lookup.json`'s `done` cases the shape NIP-CASH §Looking Up a Call
  defines, a destroyed bill's tombstone in place of its token, and drops the
  withdrawal tombstone case; renames `request-id.json`'s list to
  `privateTransportRequiresRequestId`.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Renames `budget-period.json`'s weekly case from `K-BP-8`, which the daily
  case also carried, to `K-BP-9`.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Refuses an `email` user id whose local part starts with `@` in
  `nipIC.CanonicalUserID` and `NewConnectionKey`: NIP-IC says no value carries
  a leading `@`, and such a value is refused, never repaired.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Refuses the platform name `nostr` in `nipIC.ValidPlatformName` (new
  `nipIC.PlatformNostr`), so no ConnectionKey, 35522 or nconnection is built on
  it, and the `nipAZ` request and receipt builders refuse an identity whose
  platform is not a NIP-IC platform name: a key sent with `nostr` was read as a
  native pubkey.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Maps chain `bitcoin` to mainnet invoices (`bc`) only in
  `nipAZ.InvoiceCurrencies`, so a testnet, signet or regtest invoice fails
  receipt rule 4 and the relay no longer counts one as sats.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Pins email and dns canonicalization to Unicode 15.0.0 (`nipIC.UnicodeVersion`):
  a build whose linked IDNA or NFC tables are another version (Go 1.27 links
  17.0.0) refuses those values with `nipIC.ErrUnicodeVersion` instead of
  computing canonical forms other clients do not share.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Runs `nipLS` encrypt and decrypt requests under the same lock as
  `sign_event`, from `Policy.Authorize` through `AuthorizeResult`, so a policy's
  quota or one-time approval on them is spent once under concurrent requests.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Refuses `nip44_message_keys` under a `nipLS` policy that allows decryption
  but not encryption, since its keys seal any plaintext under the payload's
  nonce; new `nipLS.Request.Encrypts` and `Decrypts` report it as both.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Refuses at `niplsp` step 4 an announcement whose `created_at` is more than
  `niplsp.MaxFutureSecs` (15 minutes) ahead of the client's clock, and `Seen`
  neither counts one as seen nor restores such a mark, so a future-dated event
  can no longer hide the node's later announcements, a `closed` one included.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Refuses in `nipCZ.IdentityFromAttestation` an attestation whose `d` is not
  the ConnectionKey of its platform and evidence user id
  (`nipcash.ErrAttestationKey`), as `nipcash.ConnectionKeyFromAttestation` does.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Runs a `nip46/bunker` server's `Policy.Authorize` and the operation it
  allows one request at a time per client pubkey, so a client's quota or
  one-time approval is spent once, while other clients' requests still run.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Refuses `nip44_message_keys` under a decrypt-only `nip46/bunker` policy, as
  `nipLS` does; new `Call.Encrypts` and `Decrypts` report it as both.
  ([#220](https://github.com/ohstr/nmilat/pull/220))
- Fixes GO-2026-5970 and GO-2026-5026 by updating `golang.org/x/text` to
  v0.39.0 and `golang.org/x/net` to v0.55.0.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes the proof vector case "no expiration tag is fine", which expects a
  refusal: it is now "missing expiration tag is refused".
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes the conformance suite skipping every withdrawal case against a Hub
  that lists `cash_withdraw_quote` and `cash_withdraw` only on its bills'
  connections: they are `cash_redeem`-scope methods of a Cash Wallet
  connection (NIP-CASH §Methods, §Scope Surface), so when the Cash Hub
  connection does not list them the suite reads the offer from a probe bill's
  own `get_info`; CT-WD-01 and CT-SCOPE-05 check the bill's lists too. The
  probe bill is redeemed right after (with a Payee), so it does not outlast
  CT-WD-24's 400 days.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes CT-WD-02 refusing an expired bill's tombstone: a Hub that has already
  swept the bill answers a quote for it with its tombstone (spent, expired).
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes `RekeyCashSlice` reporting success with no token when the Hub spins
  the slice off, stranding the funds: it returns the new wallet's pubkey and
  token, and consolidates an interim spin-off over its own connection.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes `nipCP.Sync` trusting the id a relay claims: each wrap's id and
  signature are checked before it is keyed, credited to a relay or marked
  seen, and it is marked seen only once it opens, so one relay can no longer
  hide a pocket entry.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes `HubClient.MintCash` closing a mint whose token's provenance is not
  the connection's Hub identity or whose wallet differs (now
  `*IndeterminateError`); a `cash_consolidate` reply with an empty
  `new_wallet_token` or `new_wallet_pubkey` being accepted (now
  `ErrEmptyToken`, the call stays pending); a cancelled withdrawal's re-minted
  bill going unchecked; and an order read back with `cash_lookup` skipping the
  invoice and address checks.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes `VerifyClaimProof` accepting any amount when a transfer or
  consolidate expectation carries none; `Item.VerificationCost` undercounting
  signed events when another params member is mistyped (an E6 budget bypass);
  item and bill proofs accepting a binding tag that is not exactly one string;
  a slice attestation repeating d, p, platform, evidence or expiration; an
  envelope refused at E5 or E6 not counting as served; `Envelope.Validate`
  skipping bill proofs and refusing a proofless cash_status or quote; and
  `CashConsolidateParams.Pending` wrapping on overflow.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes connections signed or encoded with `Label` but not `HasLabel`
  (unverifiable), and labels that are not valid UTF-8.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes nipCZ refusing a 9522 carrying `e` and `a` whose coordinate author is
  not `p`, counting a zap on a target event that does not verify, and taking
  minters from a 10522 another author signed.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes listings accepting a mixed-case `connection`, hiding a listing whose
  `get_info` lacks the `cash_order` key (now unavailable), refusing repeated
  hints of an unknown payment kind, and reading a curator set by any `d`.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes niplsp never using a remembered node's announcement again after a
  restart, reading an LSPS1 answer with no known option as not asked, calling
  a live check usable without list_protocols, and accepting a non-ASCII
  website or image URL.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes `TransferFromSources` dropping the session options on reconnect,
  `LookupJoin` accepting a credential that is not the join's signer,
  `ParseCreateCircleWalletRequest` reading an empty `budget_renewal` as never,
  `nipCP` `Balance` wrapping on overflow, a Spoofed nipIC verdict naming the
  pubkey of a claim whose signature fails, and `signeruri` passing the
  untrimmed URI on.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes `transport.DecodeResponse` refusing a reply larger than the request's
  `max_bytes`: a Hub pads each reply chunk against its own `pad_bucket_bytes`
  and 65535, as `EncodeResponse` does, so a reply is now refused for size
  only past 65535 bytes.
  ([#218](https://github.com/ohstr/nmilat/pull/218))
- Fixes the credential decoders and `nipIC.DecodeNConnection` accepting a
  bech32m checksum: these strings are NIP-19 bech32 (on-chain addresses keep
  BIP-350). ([#206](https://github.com/ohstr/nmilat/pull/206))
- Fixes `ParseWithdrawCancel` giving a different reason from run to run when
  several fields are bad. ([#206](https://github.com/ohstr/nmilat/pull/206))
- Fixes `nip01.SubscriptionFilter` losing its tag filters when marshalled by
  value. ([#206](https://github.com/ohstr/nmilat/pull/206))
- Fixes a proof missing several tags reporting a different missing tag from
  run to run. ([#205](https://github.com/ohstr/nmilat/pull/205))
- Fixes `nipLS` line limits: a line of exactly 1 MiB plus its newline is
  accepted on both sides; a longer one is answered `invalid: line too long`,
  discarded without buffering, and the connection kept; a result whose
  response would exceed 1 MiB is answered `result too large`.
  ([#203](https://github.com/ohstr/nmilat/pull/203))
- Fixes `nipCP` relay queries losing an event that arrived just before EOSE.
  ([#201](https://github.com/ohstr/nmilat/pull/201))
- Fixes `IndeterminateError`'s text including the call's request_id.
  ([#197](https://github.com/ohstr/nmilat/pull/197))
- Fixes `CashTransferResult.RecipientToken` to report a spin-off answered
  with its tombstone as `*BillGone`.
  ([#196](https://github.com/ohstr/nmilat/pull/196))
- Fixes `client.AcceptedRecorder` being told item ids instead of
  request_ids. ([#195](https://github.com/ohstr/nmilat/pull/195))
- Fixes chunked private replies so a chunk that contradicts another no longer
  discards outcomes already decided, including a new bill's only token. An item
  answered twice, or answered beside an envelope-level error, is reported on its
  own as `OutcomeConflict`, and the reply as `ErrReplyConflict`.
  ([#181](https://github.com/ohstr/nmilat/pull/181))
- Fixes the relay hanging for good when a query with several tag names (or a
  `before_id`) ran inside a read while a write grew the store, as a NIP-29
  group create does. The query is now planned in the caller's transaction.
  ([#217](https://github.com/ohstr/nmilat/pull/217))
- Fixes a filter with several tag names walking every tag's index range, so
  `#d` lookups plus a common tag such as `#l` took seconds. The scan now
  follows the smallest range and checks the other tags per event.
  ([#177](https://github.com/ohstr/nmilat/pull/177))
- Fixes a REQ failing with `event not found` when a matching event was
  deleted or replaced while its scan ran.
  ([#177](https://github.com/ohstr/nmilat/pull/177))
- Fixes an index rebuild or catch-up panicking (index out of range [-1]), and
  freeing old index data spinning without progress, when waiting for a busy
  store's writer took longer than the 25 ms batch budget.
  ([#175](https://github.com/ohstr/nmilat/pull/175))
- Fixes a rebuild putting back an event deleted while it ran.
  ([#173](https://github.com/ohstr/nmilat/pull/173))
- Fixes a contact list losing the follows it shares with an older one
  removed after it.
  ([#173](https://github.com/ohstr/nmilat/pull/173))
- Fixes a profile the store rejected as older still reaching search.
  ([#173](https://github.com/ohstr/nmilat/pull/173))
- Fixes follower counts doubling when a contact list and its replacement
  were written in the same batch.
  ([#171](https://github.com/ohstr/nmilat/pull/171))
- Fixes `ClearZapIndex`/`ClearSocialIndex` leaving the index marked current
  and racing a running rebuild. They now refuse during one, and work on a
  disabled index.
  ([#171](https://github.com/ohstr/nmilat/pull/171))
- Fixes social scores staying stale after a social index rebuild; every
  profile is rescored in the background.
  ([#171](https://github.com/ohstr/nmilat/pull/171))
- Fixes a crashed rebuild's leftover generation receiving every later write;
  it is dropped on open.
  ([#171](https://github.com/ohstr/nmilat/pull/171))
- Fixes a NIP-77 sync with a large diff sending one message of several MB,
  which strfry and most relays drop mid-sync. Messages now stop at
  `Negentropy.FrameSizeLimit` (30 000 bytes by default) and the rest follows
  in later rounds, split and encoded as the reference negentropy does.
  ([#169](https://github.com/ohstr/nmilat/pull/169))
- Fixes zap amounts in different coins being added together in the zap
  ranking, NIP-85 `zap_amt_*` and the social score.
  ([#167](https://github.com/ohstr/nmilat/pull/167))
- Fixes an unknown cache action falling through to a REQ with an empty
  filter, which streamed the whole store; it's `CLOSED "invalid: unknown
  cache action ..."`. A ranking window like `7d` was also read with Go's
  duration parser and silently became the default.
  ([#167](https://github.com/ohstr/nmilat/pull/167))
- Fixes a panic in a profile check killing the relay: each check and each
  verification job now recovers, logs the panic and counts the check as a
  transient failure. ([#166](https://github.com/ohstr/nmilat/pull/166))
- Fixes a REQ with a malformed `ids` or `authors` value being logged at error
  and answered `CLOSED "error: ..."`; it's now `CLOSED "invalid: ..."`, logged
  at debug. A client going over `max_message_length` or resetting the
  connection is no longer logged as an unexpected session error. ([#166](https://github.com/ohstr/nmilat/pull/166))
- Fixes top-zapped counting invalid receipts, self-zaps, duplicate receipts of
  one payment and deleted receipts; once a recipient's LNURL server key is
  verified, only receipts that server signed count.
  ([#165](https://github.com/ohstr/nmilat/pull/165))
- Fixes the gibberish-name search penalty hitting names written in non-Latin
  scripts. ([#164](https://github.com/ohstr/nmilat/pull/164))
- Fixes search scores dropping a profile's base score once its checks ran, and
  counting a picture twice; a broken picture now takes its bonus back.
  ([#164](https://github.com/ohstr/nmilat/pull/164))
- Fixes profiles shown as having no NIP-05 or lightning address when they do:
  identifiers with TLDs longer than four letters are checked now, and a
  timeout, block or 5xx keeps the previous result and is retried instead of
  counting as "not verified" for two weeks.
  ([#164](https://github.com/ohstr/nmilat/pull/164))
- Fixes a data race when two connections open at once on a relay whose
  `SessionConfig` leaves `OutgoingBufferSize` or `MaxConcurrentStoreTasks` at
  zero: each session now fills those defaults on its own copy instead of
  writing them into the config every session shares.
  ([#162](https://github.com/ohstr/nmilat/pull/162))
- Fixes NIP-50 search results coming back newest profile first instead of
  best match first, and `limit` keeping the most recently updated profiles;
  results now follow the search ranking, with `limit` applied after it.
  ([#155](https://github.com/ohstr/nmilat/pull/155))
- Fixes `nipcw/client.Client.Join` hiding why a join was not private when the
  standard join fails after a fallback; `errors.Is` now finds the fallback
  reason, such as `cashclient.ErrNoAnnouncement`, as well as the refusal.
  ([#151](https://github.com/ohstr/nmilat/pull/151))
- Fixes `NWCClient` accepting a response signed by any key, so anyone on the
  relay could answer a request with a forged result; responses must now come
  from the wallet's pubkey and carry a valid signature.
  ([#138](https://github.com/ohstr/nmilat/pull/138))
- Fixes `nip44.Encrypt` sealing an empty plaintext no peer can open and
  wrapping the length of one above 65535 bytes; it now follows NIP-44's current
  rules, 1 byte to 4 GiB with a 6-byte length prefix from 65536 bytes on.
  ([#139](https://github.com/ohstr/nmilat/pull/139))
- Fixes the BOLT11 decoder wrapping an amount too large for int64 millis into a
  different, sometimes negative, amount; such an invoice is now refused.
  ([#140](https://github.com/ohstr/nmilat/pull/140))
- Fixes a secret key of 0 or at or above the curve order being reduced into a
  meaningless key; `GetPublicKey`, `Event.Sign` and every encrypt, sign and
  proof path now refuse it with `utils.ErrInvalidPrivateKey`.
  ([#141](https://github.com/ohstr/nmilat/pull/141))
- Fixes event ids for tags containing U+2028 or U+2029 differing from every
  other client, so a relay or Hub on nmilat no longer rejects such an event
  signed elsewhere; tags are now escaped exactly like the content.
  ([#142](https://github.com/ohstr/nmilat/pull/142))
- Fixes `nip04.Decrypt` panicking on a ciphertext that is not a whole number
  of AES blocks, so one malformed payload could crash the caller; it now
  returns an error. ([#144](https://github.com/ohstr/nmilat/pull/144))
- Fixes `ReindexZaps` skipping NIP-57 zap receipts (kind 9735), so a rebuilt
  zap index no longer drops them from top-zapped results.
  ([#131](https://github.com/ohstr/nmilat/pull/131))
- Fixes the NIP-CASH private transport listening for a reply on a different
  relay than the one the request was sent to, and treating a relay's refusal
  of the request as sent; a reply chunk delivered by two relays now counts
  once instead of voiding the reply.
  ([#128](https://github.com/ohstr/nmilat/pull/128))
- Fixes the NIP-CASH private transport sometimes losing a reply when the
  request reached the relay before the reply subscription; it now waits for the
  relay's EOSE before sending. ([#129](https://github.com/ohstr/nmilat/pull/129))
- Fixes NIP-CASH request builders accepting what the spec forbids: an empty
  `mint_cash`, a `connection_key` without `ia_pubkey`, a negative or sub-second
  expiry, duplicate consolidate sources, and a nil transfer target (which
  panicked) now return errors. ([#125](https://github.com/ohstr/nmilat/pull/125))
- Fixes `ItemOutcome.SafeToResend` saying yes for a Hub success the client
  could not decode; such an item ran, so it is now marked `Undecodable` and not
  safe to resend. ([#125](https://github.com/ohstr/nmilat/pull/125))
- Fixes `BatchSession`'s `*Many` methods sending a bill minted by another node,
  which the Hub can only answer with silence; such a bill is now refused locally.
  ([#125](https://github.com/ohstr/nmilat/pull/125))
- Fixes the NIP-CASH private transport reporting only the last relay's failure
  when none could be reached; the error now names every relay it tried.
  ([#125](https://github.com/ohstr/nmilat/pull/125))
- Fixes `nipcash.VerifyProvenance` checking every token under flnd's
  signing prefix: a `satscash1` token now verifies under lnd's, and an HRP
  with no known prefix reports no provenance instead of a wrong minter.
  ([#121](https://github.com/ohstr/nmilat/pull/121))
- Fixes `CheckClaim` reporting a spent bill as `ErrClaimNotFound`: a Hub's
  tombstone now returns the new `nipcash.ErrSpent`, so a caller can tell
  "spent" from "this token does not name you".
  ([#121](https://github.com/ohstr/nmilat/pull/121))
- Fixes `CashConsolidate` returning the still-encrypted token as
  `NewWalletToken` when the delivery could not be decrypted; it now returns an
  error, as `CashTransfer` does. ([#121](https://github.com/ohstr/nmilat/pull/121))
- Fixes `transport.Decode` rejecting a whole envelope over one malformed item:
  it now drops that item into the new `Envelope.Rejected` and keeps its
  co-items, since NIP-CASH authorizes each item on its own.
  ([#121](https://github.com/ohstr/nmilat/pull/121))
- Fixes chunked NIP-CASH replies read as incomplete or thrown away: the client
  bounds a Hub's `total` by the items it asked for, keeps earlier chunks when a
  later one is malformed, and returns a Hub's refusal ahead of "incomplete".
  ([#121](https://github.com/ohstr/nmilat/pull/121))
- Fixes the cash token, `cashhub1` and `circlehub1` encoders writing a blank
  relay entry that sat beside a real one; blank entries are now skipped.
  ([#121](https://github.com/ohstr/nmilat/pull/121))
- Fixes `nipAZ.NewAltZapOnBehalfRequest`, whose events never verified because
  they were signed before their kind was set to 9523.
  ([#116](https://github.com/ohstr/nmilat/pull/116))
- Fixes deleting a NIP-29 subgroup: it now leaves its parent's `child` list,
  and an edit to the parent drops ids that no longer exist, so a stale child no
  longer breaks reads of the parent's subgroups.
  ([#109](https://github.com/ohstr/nmilat/pull/109))
- Fixes NIP-29 edits that flip `private` on a linked group: one whose parent
  or a subgroup would end up with the other setting is now refused, as linking
  such groups already was.
  ([#102](https://github.com/ohstr/nmilat/pull/102))
- Fixes NIP-29 moderation from the relay's own key: NIP-11 `self` now acts
  as admin of every hosted group, so an operator can recover a group whose
  admins are gone. `GroupsService.SetRelayAdmins` adds other operator keys.
  ([#101](https://github.com/ohstr/nmilat/pull/101))
- Fixes NIP-29 `restricted` and `hidden` group flags having no effect: a
  restricted group now refuses posts from non-members, and a hidden group's
  metadata (kinds 39000-39005) is no longer served to non-members.
  ([#100](https://github.com/ohstr/nmilat/pull/100))
- Fixes `Pool` with a Signer waiting 5s before subscribing on relays that never
  send AUTH; it now waits only when challenged, and syncs again after a late AUTH.
  ([#93](https://github.com/ohstr/nmilat/pull/93))
- Fixes `Pool` handing an event on twice in one sync when it arrived both live
  and in a page. ([#93](https://github.com/ohstr/nmilat/pull/93))
- Fixes `Pool` subscribing before NIP-42 AUTH settled and never retrying after
  a `restricted:` or `auth-required:` CLOSED, which left its subscription
  silently dead on relays that require AUTH. ([#91](https://github.com/ohstr/nmilat/pull/91))
- Fixes reads starving under many open subscriptions; live subscriptions now
  read only new events. ([#81](https://github.com/ohstr/nmilat/pull/81))
- Returns NIP-77 and NIP-05 fetch errors (NIP-05: 503) instead of answering
  from partial results. ([#81](https://github.com/ohstr/nmilat/pull/81))
- Answers every EVENT with exactly one OK, including `error: relay busy` after
  10s queued. ([#78](https://github.com/ohstr/nmilat/pull/78))
- Retries a failed store batch event by event, so only the failing event is
  rejected. ([#78](https://github.com/ohstr/nmilat/pull/78))
- Fixes a data race when connections opened at the same time. ([#75](https://github.com/ohstr/nmilat/pull/75))
- Fixes a filter's `limit` returning the oldest events instead of the newest.
  ([#37](https://github.com/ohstr/nmilat/pull/37))
- Fixes `#h=1` matching nothing once a longer value such as `h=10` exists.
  ([#37](https://github.com/ohstr/nmilat/pull/37))
- Fixes NIP-77 reconciliation of events that share a timestamp. ([#37](https://github.com/ohstr/nmilat/pull/37))
- Fixes deleting an ephemeral event leaving its expiration entry. ([#37](https://github.com/ohstr/nmilat/pull/37))
- Fixes the relay rejecting about 73% of real zap receipts; ingest enforces
  only NIP-57's MUST rules. ([#35](https://github.com/ohstr/nmilat/pull/35))
- Fixes the swapped `have=` and `want=` values in a description-hash mismatch.
  ([#35](https://github.com/ohstr/nmilat/pull/35))
- Checks the PoW floor before the signature, the cheaper check. ([#39](https://github.com/ohstr/nmilat/pull/39))
- Stops NIP-59 setting the seal's `PubKey` to the private key before signing,
  and randomizes `created_at` on seals and wraps. ([#39](https://github.com/ohstr/nmilat/pull/39))
- Fixes `nipcw` missing the join result's fields. ([#39](https://github.com/ohstr/nmilat/pull/39))
- Fixes a tagless event's id, which was wrong on every relay. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Fixes batch replies being lost at EOSE, cash-mode bills not batching, and a
  nil proof breaking every proofless item. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Fixes four ways a hostile Hub or relay could mislead a batching client, and
  surfaces an envelope-level refusal per item. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Bounds collecting a reply by the envelope's own deadline. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Tries every relay hint in `NewNWCClient`, not only the first. ([#43](https://github.com/ohstr/nmilat/pull/43))
- Fixes `ParseNostrconnect` rejecting spec-compliant URIs, using only the first
  relay, and dropping `perms`. ([#42](https://github.com/ohstr/nmilat/pull/42))
- Fixes NIP-98 treating an anti-replay `nonce` tag as proof-of-work. ([#51](https://github.com/ohstr/nmilat/pull/51))
- Fixes `POST /query` ignoring its `limit`, and paging past same-second events
  with `BeforeID`. ([#52](https://github.com/ohstr/nmilat/pull/52))
- Redials and re-authenticates once in `ReadEventsFromRelayWithAuth` if the
  relay drops the connection mid-wait. ([#63](https://github.com/ohstr/nmilat/pull/63))
- Fixes a redundant CLOSE ending the whole session. ([#65](https://github.com/ohstr/nmilat/pull/65))
- Fixes a query without a group tag returning private groups' metadata to
  anyone. ([#66](https://github.com/ohstr/nmilat/pull/66))
- Fixes an authenticated read losing the race with its own AUTH. ([#66](https://github.com/ohstr/nmilat/pull/66))
- Fixes `COUNT` revealing how many private groups exist, and an unknown group
  being distinguishable from a private one. ([#70](https://github.com/ohstr/nmilat/pull/70))
- Withholds private group content from non-members by any filter, live tails
  and `COUNT` included. ([#74](https://github.com/ohstr/nmilat/pull/74))
- Lets only members post into a private or closed group. ([#74](https://github.com/ohstr/nmilat/pull/74))
- Purges a deleted group's events and never serves a deleted group's mirrors.
  ([#74](https://github.com/ohstr/nmilat/pull/74))
- Applies NIP-43 joins and removals to members' open connections, and
  publishes the kind:13534 member list. ([#74](https://github.com/ohstr/nmilat/pull/74))
- Applies REQ's access checks to `COUNT`, and refuses unauthenticated clients
  with `auth-required:`. ([#74](https://github.com/ohstr/nmilat/pull/74))
- Fixes every `#d` lookup naming no group being refused. ([#74](https://github.com/ohstr/nmilat/pull/74))
- Rejects another key's create of an existing group id. ([#74](https://github.com/ohstr/nmilat/pull/74))
- Fixes NIP-26 signing the wrong delegation string; re-issue older tokens.
  ([#74](https://github.com/ohstr/nmilat/pull/74))
- Reports a refused anonymous read in `ReadEventsFromRelayWithAuth`, and
  retries a read answered before AUTH landed. ([#74](https://github.com/ohstr/nmilat/pull/74))
- Bounds `NewConnection`'s handshake by the caller's context. ([#74](https://github.com/ohstr/nmilat/pull/74))

## [0.4.0]

NIP-CASH's bearer mode becomes cash mode; this release needs lokihub
0.5.0-rc.6 or later.

### Breaking

- Renames NIP-CASH's bearer mode to cash mode on the wire (`identity_type:
  "cash"`, `cash_secret`). Existing tokens are unaffected. ([#32](https://github.com/ohstr/nmilat/pull/32))
- Renames the Go API to match (`CashTarget`, `CashSecret`, `IsCash`,
  `SplitCashSliceString`, `RekeyCashSlice`, …), with no deprecated aliases.
  ([#32](https://github.com/ohstr/nmilat/pull/32))

## [0.3.2]

A fix for a relay that could freeze.

### Fixed

- Fixes the relay freezing until restarted when a write had to grow the
  database while a REQ was sending. ([#29](https://github.com/ohstr/nmilat/pull/29))

### Changed

- Removes debug logging from busy relay paths. ([#29](https://github.com/ohstr/nmilat/pull/29))

## [0.3.1]

A zap invoice helper and a consolidate fix.

### Added

- Adds `nip57.RequestZapInvoice`, which resolves a LUD-16 address and fetches
  a zap invoice in one call. ([#28](https://github.com/ohstr/nmilat/pull/28))
- Exports `utils.ParsePayMetadataChains` and `utils.FetchLud16PayResponse`.
  ([#28](https://github.com/ohstr/nmilat/pull/28))

### Changed

- Adds an optional `Content` to `nip57.ZapRequestParams`. ([#28](https://github.com/ohstr/nmilat/pull/28))

### Fixed

- Fixes `CashConsolidateParams.ParseResult` failing after a successful merge
  when the new wallet's token arrives unencrypted. ([#28](https://github.com/ohstr/nmilat/pull/28))

## [0.3.0]

Git collaboration over Nostr, and cash-slice lifecycle helpers.

### Added

- Adds `nip34` for git collaboration (repositories, patches, pull requests,
  issues, statuses) and `nip22` for comments. ([#23](https://github.com/ohstr/nmilat/pull/23))
- Adds `utils.FormatATag`. ([#23](https://github.com/ohstr/nmilat/pull/23))
- Adds `ResolvedConnectionKey`, `SplitBearerSliceString`, `CheckClaim` and
  `IsPubkeyTarget` to `nipcash`. ([#22](https://github.com/ohstr/nmilat/pull/22))
- Adds `RekeyBearerSlice`, `TransferFromSources` and `PartialProgressError` to
  `nipcash/client`. ([#22](https://github.com/ohstr/nmilat/pull/22))
- Adds `GetInfoResult.CircleWallet` and `FeeSkimMloki`, which were dropped on
  unmarshal. ([#21](https://github.com/ohstr/nmilat/pull/21))

### Changed

- Accepts a bearer or connection-key target in `cash_consolidate`, and renames
  `ErrConsolidateTargetNotPubkey` to `ErrConsolidateTargetInvalid`. ([#22](https://github.com/ohstr/nmilat/pull/22))

### Fixed

- Fixes `TransferFromSources` reusing a stale client, which failed every
  multi-source transfer. ([#22](https://github.com/ohstr/nmilat/pull/22))

## [0.2.9]

Fixes for stalled delivery and corrupted events.

### Fixed

- Defaults `DataWriteTimeout` to 30s, so one slow reader can't stall every
  subscription on its connection, and logs slow writes. ([#19](https://github.com/ohstr/nmilat/pull/19))
- Fixes `FindEventBytes` serving corrupted event JSON under load. ([#19](https://github.com/ohstr/nmilat/pull/19))

## [0.2.8]

Hub connection strings and NWC fixes.

### Added

- Adds `cashhub1…` and `circlehub1…` encodings for a Hub's connection, with an
  optional label. ([#16](https://github.com/ohstr/nmilat/pull/16))
- Adds `nip47.ParseResponseEventWithFallback` and
  `relay/client.SubscriptionClosedError`. ([#14](https://github.com/ohstr/nmilat/pull/14))

### Fixed

- Fixes `NWCClient` misreading untagged NIP-44 responses as NIP-04. ([#14](https://github.com/ohstr/nmilat/pull/14))
- Fixes a burst on one kind starving the other kinds in a live subscription.
  ([#15](https://github.com/ohstr/nmilat/pull/15))

## [0.2.7]

Identity Connection, NIP-CASH and NIP-CW, and a Go 1.26.8 minimum.

### Changed

- Requires Go 1.26.8, which clears several standard-library CVEs. ([#7](https://github.com/ohstr/nmilat/pull/7))
- Changes `nipAZ` to take an `Identity` instead of raw strings, sign
  internally, and require the on-behalf sender. This breaks existing callers.
  ([#4](https://github.com/ohstr/nmilat/pull/4))

### Added

- Adds `nipIC` for NIP-IC: attestations, identity connections, the npv1
  challenge and `nconnection` encoding. ([#4](https://github.com/ohstr/nmilat/pull/4))
- Adds `nipcash` and `nipcash/client` for NIP-CASH cash tokens. ([#9](https://github.com/ohstr/nmilat/pull/9))
- Adds `nipcw` and `nipcw/client` for NIP-CW circle wallets. ([#9](https://github.com/ohstr/nmilat/pull/9))

### Fixed

- Stops `NewAltZapReceipt` re-deriving `p`/`P` tags from the embedded
  request. ([#4](https://github.com/ohstr/nmilat/pull/4))
- Sizes `nipIC.NewChallenge`'s entropy to 16 bytes. ([#4](https://github.com/ohstr/nmilat/pull/4))
- Returns an error when the delete path fails to update the expiration index.
  ([#7](https://github.com/ohstr/nmilat/pull/7))
- Accepts an event whose PoW tag overclaims its difficulty unless `StrictPow`
  is on. ([#10](https://github.com/ohstr/nmilat/pull/10))

## [0.2.6]

AltZap moves to its own package.

### Changed

- Moves AltZap from `nip57` to `nipAZ`. This breaks imports of
  `nip57.AltZap*`. ([#2](https://github.com/ohstr/nmilat/pull/2))

### Added

- Lets `AltZapReceiptParams` set the receipt's `r`/`R`/`a`/`e` tags directly.
  ([#2](https://github.com/ohstr/nmilat/pull/2))

## [0.2.5]

A goroutine leak fix.

### Fixed

- Fixes `relay/client.Connection` leaking a goroutine when nobody reads its
  errors. ([#1](https://github.com/ohstr/nmilat/pull/1))

## [0.2.4]

Full Blossom support.

### Added

- Expands `nipB7` to the full Blossom protocol (BUD-01 through BUD-12). ([b479e6d](https://github.com/ohstr/nmilat/commit/b479e6d))
- Adds `nipB7/client`, an HTTP client for Blossom servers. ([b479e6d](https://github.com/ohstr/nmilat/commit/b479e6d))

### Removed

- Removes the unused `nip11.DelegationConfig`. ([781d85b](https://github.com/ohstr/nmilat/commit/781d85b))

## [0.2.3]

NIP-43 membership, agent auth and owner attestation.

### Added

- Adds NIP-43 membership: join, leave and invites, and a
  `membership_required` gate. ([5d29cc2](https://github.com/ohstr/nmilat/commit/5d29cc2))
- Adds `nipOA` for NIP-OA owner attestations. ([d102626](https://github.com/ohstr/nmilat/commit/d102626))
- Adds `nipAA` for NIP-AA agent auth, with virtual membership for agent keys.
  ([f315db9](https://github.com/ohstr/nmilat/commit/f315db9))
- Lets one connection hold several authenticated pubkeys. ([c50bdae](https://github.com/ohstr/nmilat/commit/c50bdae))
- Adds `relay.RegisterLetteredNIP` for NIPs such as B0 and B7. ([154c1b2](https://github.com/ohstr/nmilat/commit/154c1b2))

### Fixed

- Advertises lettered NIPs as strings in NIP-11. ([154c1b2](https://github.com/ohstr/nmilat/commit/154c1b2))
- Fixes `EventStore.Close` racing in-flight tasks, NIP-42 checking the event's
  own relay tag, PoW thresholds, disabled search, and the reader ignoring
  cancellation. ([6d69bb2](https://github.com/ohstr/nmilat/commit/6d69bb2))

## [0.2.2]

The first public release.

### Added

- Adds `nip01` and 28 NIPs, an embeddable bbolt-backed relay engine, a relay
  client and profile search. ([14592fa](https://github.com/ohstr/nmilat/commit/14592fa))
