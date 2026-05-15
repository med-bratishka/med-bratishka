# Auth 2FA And Chat Notifications Plan

## Current Architecture

This repository contains a compact Go backend, not a multi-service deployment. The actual backend is in `backend/` and uses REST over `gorilla/mux`, PostgreSQL via `sqlx`, JWT via `github.com/golang-jwt/jwt/v5`, S3/MinIO for attachments, and generated Swagger models. No Redis, Kafka, RabbitMQ, Kubernetes manifests, NGINX config, or WebSocket gateway are present in the repo.

Current auth flow:

1. `POST /auth/register` creates a `users` row and immediately creates access/refresh rows in `auth_tokens`.
2. `POST /auth/login` finds a user by login/email/phone, checks bcrypt password, and creates access/refresh JWTs.
3. JWTs contain user id, role, purpose, session number, expiry, and a per-token secret.
4. `AuthMiddleware` validates access JWTs by parsing the token and checking the secret in `auth_tokens`.
5. `RefreshMiddleware` validates refresh JWTs the same way, then refresh revokes the whole session number and issues a new session.

Current chat flow:

1. `POST /chats` creates or finds a doctor-patient chat after checking bindings.
2. `POST /chats/{chat_id}/messages` validates membership, optionally uploads attachment, inserts `messages`, and updates `chats.updated_at`.
3. There is no backend unread state, notification delivery state, push pipeline, WebSocket endpoint, presence model, or event outbox before this change.

## Login Sequence

Existing users without 2FA:

```text
client -> /auth/login -> AuthHandler -> AuthService.Login
AuthService -> UsersRepository.GetByAccessParameter
AuthService -> bcrypt.Compare
AuthService -> auth_tokens(access, refresh)
AuthService -> client: AuthResponse
```

Users with 2FA enabled:

```text
client -> /auth/login(password)
AuthService -> validate password
AuthService -> trusted_devices lookup
AuthService -> auth_challenges(login_2fa)
client <- two_factor_required + challenge_id
client -> /auth/2fa/verify(code or recovery_code)
AuthService -> SELECT challenge FOR UPDATE
AuthService -> validate TOTP or one-time recovery code
AuthService -> consume challenge
AuthService -> optional trusted_devices insert
AuthService -> auth_tokens(access, refresh)
client <- AuthResponse
```

## Main Security Findings

- Default `JWT_SECRET` and new default 2FA encryption key are development placeholders and must be replaced in production.
- Previously `/auth/refresh`, `/auth/logout`, and `/auth/logout-all` were registered without auth middleware, so `GetUserFromContext` would be nil. This change wires access/refresh middleware explicitly.
- Session number allocation used `MAX(session_number)+1`; concurrent logins could collide. This change adds a PostgreSQL advisory transaction lock and creates both tokens in one transaction.
- Access/refresh token secrets are stored plaintext in PostgreSQL. Better long-term design: hash token secrets in `auth_tokens`, compare hashes, and rotate JWT signing keys with `kid`.
- Login has no distributed password-attempt limiter. 2FA challenge attempts are capped, but production password brute-force protection should use Redis or a DB-backed limiter keyed by user + IP.
- Chat attachment upload happens before DB commit. If DB commit fails after S3 upload, an orphan object can remain. Add cleanup or an attachment outbox/saga.

## 2FA Data Model

```text
users 1--1 user_totp_settings
users 1--N recovery_codes
users 1--N trusted_devices
users 1--N auth_challenges
users 1--N auth_audit_log

user_totp_settings(user_id, role, secret_ciphertext, enabled, confirmed_at, disabled_at)
recovery_codes(user_id, role, code_hash, used_at)
trusted_devices(user_id, role, token_hash, user_agent_hash, expires_at, revoked_at)
auth_challenges(id, user_id, role, purpose, failed_attempts, expires_at, consumed_at)
auth_audit_log(user_id, role, event_type, ip, user_agent, metadata)
```

TOTP secret storage:

- Generate 160-bit random TOTP secret.
- Return plaintext only once during setup.
- Store only AES-GCM ciphertext in `user_totp_settings.secret_ciphertext`.
- Use `TWO_FACTOR_ENCRYPTION_KEY` from a secret manager; rotate by adding key version columns or by decrypt-reencrypt background migration.

Recovery codes:

- Generate 10 one-time codes.
- Store bcrypt hashes only.
- Mark used codes with `used_at`.
- Regeneration invalidates all active codes.

Trusted devices:

- Client stores opaque random token.
- Server stores HMAC-SHA256 hash only.
- Token is bound to user, role, user-agent hash, expiry, and revocation state.

## Chat Notification Model

```text
messages 1--1 outbox_events(chat.message.created)
outbox_events 1--N chat_notification_deliveries
chats/users 1--1 chat_read_state
```

On message send, the same DB transaction now:

1. Inserts `messages`.
2. Inserts an idempotent `outbox_events` row.
3. Inserts recipient `chat_notification_deliveries`.
4. Increments recipient `chat_read_state.unread_count`.
5. Marks sender read state at the new message id.

Future event pipeline:

```text
PostgreSQL outbox poller -> Kafka/RabbitMQ topic chat.message.created
consumer -> WebSocket gateway fan-out by recipient_id
consumer -> push provider for offline users
consumer -> delivery status update
```

Redis strategy for the future WebSocket gateway:

- `presence:user:{id}` -> connection/session metadata with short TTL.
- `ws:user:{id}:nodes` -> gateway node ids for fan-out routing.
- `typing:chat:{chat_id}:user:{id}` -> TTL 3-5 seconds.
- `unread:user:{id}` -> optional cached aggregate, source of truth remains PostgreSQL.
- Rate limits: `rl:login:{ip}`, `rl:2fa:{user_id}`, `rl:ws:{user_id}`.

Reconnect/replay:

- Client stores last received event id or message id.
- On reconnect, call REST replay endpoint or WebSocket resume with cursor.
- Server queries durable `messages`/`chat_notification_deliveries`, then resumes live subscription.
- Consumers and clients must treat events as at-least-once and dedupe by `idempotency_key` or message id.

## Rollout

1. Deploy migrations first; they only add tables/indexes.
2. Deploy backend with 2FA disabled by default for existing users.
3. Configure strong `JWT_SECRET` and `TWO_FACTOR_ENCRYPTION_KEY`.
4. Enable 2FA setup endpoints in frontend behind feature flag.
5. Roll out login 2FA challenge handling to clients.
6. Start an outbox publisher service when Kafka/RabbitMQ/WebSocket infrastructure exists.
7. Add Redis-backed distributed rate limiting before opening 2FA to high-risk accounts.

Rollback:

- Backend rollback is safe while new tables remain unused.
- Do not drop 2FA tables if any users enabled TOTP; disable feature flag first.
- To rollback notifications, stop outbox publisher and keep durable rows for replay.

Testing checklist:

- Unit-test TOTP valid current/skew/invalid codes.
- Integration-test login without 2FA, setup, confirm, login challenge, verify, recovery code, trusted device, disable.
- Concurrent login test for session number uniqueness.
- Chat send test verifies message, outbox, delivery, unread state in one transaction.
- Security test invalid/expired/reused challenge and recovery code replay.
- Load test message fan-out separately from message persistence.

Observability:

- Metrics: login success/fail, 2FA required/success/fail, challenge attempts, trusted-device use, outbox pending age, outbox publish failures, unread updates, WebSocket connections when added.
- Logs: audit table for auth events; structured app logs include request id.
- Alerts: high failed login/2FA rate, old pending outbox events, DB deadlocks, auth token validation errors spike.
