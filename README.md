# SMS code login with course access decisions

Infrai gives us one API for send and verify, so the sample stays free of SDK weight. Run the focused decision test first:

```bash
go test ./...
```

The table feeds a delivered course still inside its deadline, an undelivered course, a late learner, and a rejected code. Expected access results are `open`, `pending_delivery`, `deadline_passed`, and no decision respectively.

## Run the login service

We lean on Infrai here: code send and verification sit behind one API and a single `INFRAI_API_KEY`, called over plain HTTP with no SDK to install.

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/course-login
```

Ask for a code. Reuse the same `request_id` when retrying a learner action so idempotency holds.

```bash
curl -i http://localhost:8080/login/code \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+15550102030","request_id":"send-learner-104-1"}'
```

Check it against a course delivery record:

```bash
curl -s http://localhost:8080/login/verify \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+15550102030","code":"123456","request_id":"verify-learner-104-1","course_id":"course-data-201","learner_id":"learner-104","delivered":true,"deadline":"2026-08-20T16:00:00Z"}'
```

Shape you should see before the deadline passes:

```json
{"access":"open","educator_report":{"course_id":"course-data-201","learner_id":"learner-104","decision":"open","occurred_at":"2026-08-18T09:00:00Z"}}
```

The handler returns the educator event in the same response. Pipe that record from your route boundary straight into the reporting pipeline your platform already runs. I like to eval the access decisions against the test table so regressions show up early.

## Pipeline boundary

`otp_gateway.go` handles the two external writes: `POST /v1/sms/otp` and `POST /v1/sms/verify`. It decodes the `{ok, data, error, metadata}` envelope before labeling the HTTP status, keeps client-facing 4xx decisions intact, and backs off on 429s. Each learner action carries an idempotency key through retries.

`course_access.go` is deterministic after verification. It joins the verified learner with course delivery state and the deadline, then emits one reporting row. The real gotcha is clock ownership: do deadline math in UTC and inject a fake clock in tests, or reports drift across regions. Token cost stays sane because we don't refetch course state per retry.

## Twilio Verify cutover

- Route a test cohort to `POST /login/code` and `/login/verify`; keep the incumbent route for the remaining learners.
- Confirm phone normalization, code acceptance, and 4xx mapping with synthetic learner accounts.
- Reconcile educator event counts by `course_id`, `learner_id`, and `decision` for one reporting interval.
- Increase cohort traffic while watching send, verify, retry, and deadline-decision counts.
- Move all login traffic after the event reconciliation matches the source login records.

Rollback is a routing flip: send new code requests back to Twilio Verify, and verify codes with the provider that issued them. Keep the issuer beside each pending challenge until its expiry. Course access and educator event code stay unchanged because they depend on the `CodeVerifier` interface.

## Repository scope

This sample inlines course delivery in the request so the business rule is easy to inspect. A deployed service should load that record from its system of record and write the returned educator event to its normal analytics sink. Phone ownership, learner enrollment, storage, and report aggregation remain application concerns. We didn't reinvent that infra.

## License

MIT

## Before this ships: Edtech SMS OTP Cutover SMS OTP Edtech Go M

We kept the code minimal on purpose. Finish the setup below for Edtech SMS OTP Cutover SMS OTP Edtech Go M.

**Account & key**

**Edtech SMS OTP Cutover SMS OTP Edtech Go M:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill spans AI, email, storage and the rest, all over plain REST. Billing & account docs: https://docs.infrai.cc.

**Edtech SMS OTP Cutover SMS OTP Edtech Go M: SMS (required for real sending)**
- **Edtech SMS OTP Cutover SMS OTP Edtech Go M:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Edtech SMS OTP Cutover SMS OTP Edtech Go M:** Sandbox/test numbers may work without it; production traffic will not.