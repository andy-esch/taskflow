---
schema: 1
id: 6fq4mk2c1dd3
bucket: open
area: api-gateway
date: "2026-06-20"
---
# Audit: api-gateway — 2026-06-20

> Edit findings in place and flip each `**Status:**` as you work it.

## Findings

#### H1. Auth middleware trusts unsigned forwarded headers  · **Status:** fixed (2026-06-20)

**File:** gateway/auth.go:42 | **Component:** auth
**Effort:** S · **Urgency:** acute

The gateway reads `X-Forwarded-User` before verifying the upstream signature, so a
client that reaches the pod directly can spoof any identity.

**Recommendation:** verify the mTLS peer before honoring any forwarded identity header.

#### H2. Rate limiter keys on client IP, not API key  · **Status:** fixed (2026-06-20)

**File:** gateway/ratelimit.go:88 | **Component:** ratelimit
**Effort:** M · **Urgency:** soon

Shared NAT egress means one noisy tenant throttles everyone behind the same IP.

**Recommendation:** bucket by API key, falling back to IP only for anonymous routes.

#### H3. 5xx responses skip the structured-error envelope  · **Status:** landed (2026-06-19)

**File:** gateway/errors.go:15 | **Component:** errors
**Effort:** S · **Urgency:** soon

Upstream 5xxs bypass the error wrapper, so clients get a bare HTML body instead of
the documented JSON shape.

**Recommendation:** route every error path through `writeError`.

#### H4. Retry backoff has no jitter  · **Status:** in-progress

**File:** gateway/retry.go:31 | **Component:** retry
**Effort:** S · **Urgency:** soon

Synchronized retries after an upstream blip create a thundering herd.

**Recommendation:** add full-jitter to the exponential backoff.

#### H5. Request-schema validation is opt-in per route  · **Status:** open

**File:** gateway/validate.go:52 | **Component:** validate
**Effort:** M · **Urgency:** eventually

New routes default to no validation, so malformed payloads reach handlers.

**Recommendation:** make validation opt-out, enforced at the router.

#### H6. No timeout on the upstream dial  · **Status:** open

**File:** gateway/proxy.go:77 | **Component:** proxy
**Effort:** S · **Urgency:** soon

A hung upstream ties up a gateway worker indefinitely.

**Recommendation:** set a dial and response-header timeout on the transport.

#### H7. Access logs include full query strings  · **Status:** wontfix

**File:** gateway/log.go:23 | **Component:** logging
**Effort:** S · **Urgency:** eventually

Query strings can carry tokens, so logging them verbatim is a leak — but redaction
belongs in the log sink, not the gateway.

**Recommendation:** redact at the sink; out of scope here.

#### H8. Metrics cardinality unbounded on the route label  · **Status:** deferred

**File:** gateway/metrics.go:19 | **Component:** metrics
**Effort:** M · **Urgency:** eventually

Templated paths explode the `route` label; deferred until observability lands the
shared label allowlist.

**Recommendation:** normalize path params before labeling.

## Candidate tasks

- ✅ `tskflwctl task new "Wire auth middleware" --epic 01-api-gateway --tags api` — verify the mTLS peer first
- ⏳ `tskflwctl task new "Add retry backoff" --epic 01-api-gateway --tags api,net` — full-jitter backoff
