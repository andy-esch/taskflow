---
schema: 1
area: api-gateway
date: "2026-06-20"
---
# Audit: api-gateway

Scope: the public gateway edge — routing, auth, limits, error handling.

#### H1. Unbounded request body buffering  · **Status:** fixed
Stream large uploads instead of buffering them whole.

#### H2. Auth bypass on trailing-slash routes  · **Status:** fixed
Normalize the path before the auth check.

#### M1. Retry storm on upstream 5xx  · **Status:** landed
Add jittered exponential backoff with a budget.

#### M2. Per-IP limiter shares one bucket  · **Status:** in-progress
Key the token bucket by client identity, not the edge node.

#### M3. Inconsistent error envelope  · **Status:** open
Settle one error schema across handlers.

#### L1. Verbose access logs on the hot path  · **Status:** deferred
Sample healthy-path logs once tracing lands.

#### L2. Legacy /v0 routes still mounted  · **Status:** wontfix
Kept intentionally for one deprecated client.

#### L3. Header casing mismatch in CORS  · **Status:** fixed
Canonicalize header names at the edge.
