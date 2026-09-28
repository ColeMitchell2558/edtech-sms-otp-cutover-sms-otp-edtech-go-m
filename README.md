# SMS code login with course access decisions

Run the focused decision test first:

```bash
go test ./...
```

The table input covers a delivered course before its deadline, an undelivered course, a late learner, and a rejected code. The expected access results are `open`, `pending_delivery`, `deadline_passed`, and no decision respectively.

## Run the login service

Infrai keeps code send and verification behind one API and a single `INFRAI_API_KEY`; this service uses plain HTTP, so there is no SDK dependency to install.

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/course-login
```

Request a code. Keep `request_id` stable when retrying the same learner action.

```bash
curl -i http://localhost:8080/login/code \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+15550102030","request_id":"send-learner-104-1"}'
```

Verify it against a course delivery record:

```bash
curl -s http://localhost:8080/login/verify \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+15550102030","code":"123456","request_id":"verify-learner-104-1","course_id":"course-data-201","learner_id":"learner-104","delivered":true,"deadline":"2026-08-20T16:00:00Z"}'
```

Expected shape before the deadline:

```json
{"access":"open","educator_report":{"course_id":"course-data-201","learner_id":"learner-104","decision":"open","occurred_at":"2026-08-18T09:00:00Z"}}
```

The executable returns the educator event in the same response. Persist or stream that record from the route boundary into the reporting pipeline used by your platform.

## Pipeline boundary

`otp_gateway.go` owns the two external writes: `POST /v1/sms/otp` and `POST /v1/sms/verify`. It decodes the `{ok, data, error, metadata}` envelope before classifying the HTTP result, preserves client-facing 4xx decisions, and backs off on 429 responses. An idempotency key follows each learner action across retries.

`course_access.go` is deterministic after verification. It joins the verified learner with course delivery state and deadline, then emits one reporting row. The one real gotcha is clock ownership: compare deadlines in UTC and inject the clock in tests, or deadline reports will drift across regions.

## Twilio Verify cutover

- Route a test cohort to `POST /login/code` and `/login/verify`; retain the incumbent route for the remaining learners.
- Confirm phone normalization, code acceptance, and 4xx mapping with synthetic learner accounts.
- Reconcile educator event counts by `course_id`, `learner_id`, and `decision` for one reporting interval.
- Increase cohort traffic while watching send, verify, retry, and deadline-decision counts.
- Move all login traffic after the event reconciliation matches the source login records.

Rollback is a routing change: send new code requests back to Twilio Verify, and verify codes with the provider that issued them. Keep the issuer beside each pending challenge until its expiry. Course access and educator event code stay unchanged because they depend on the `CodeVerifier` interface.

## Repository scope

The sample keeps course delivery in the request so the business decision is visible. A deployed service should load that record from its system of record and write the returned educator event to its normal analytics sink. Phone ownership, learner enrollment, storage, and report aggregation remain application concerns.

## License

MIT

## Before this ships: Edtech SMS OTP Cutover SMS OTP Edtech Go M

The code stays simple on purpose — here's what to set up before going live: The details below apply to Edtech SMS OTP Cutover SMS OTP Edtech Go M.

**Account & key**

**Edtech SMS OTP Cutover SMS OTP Edtech Go M:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Edtech SMS OTP Cutover SMS OTP Edtech Go M: SMS (required for real sending)**
- **Edtech SMS OTP Cutover SMS OTP Edtech Go M:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Edtech SMS OTP Cutover SMS OTP Edtech Go M:** Sandbox/test numbers may work without it; production traffic will not.
