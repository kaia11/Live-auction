# Backend Integration Test Report (No Docker)

This report records the full testing process executed on 2026-06-02, including environment blockers, fixes, API validation flow, and final results.

## Scope

- Backend runtime availability
- Redis dependency availability
- Authentication and authorization
- Auction session state-machine behavior
- Admin item/session operations
- Bid main flow
- Ranking and my-status synchronization
- Duplicate request idempotency

## Environment

- OS: Windows 10 (PowerShell)
- Repo: `D:/cuhk/bytan/Live-auction-main_v2/Live-auction-main`
- Backend service: `http://localhost:8080`
- Redis source: native Windows package (`Redis.Redis`)
- MySQL: not required for this run (backend continued with memory fallback)

## Full Process and Results

### 1) Docker/WSL route blocked on this machine

- `docker` command was unavailable.
- Docker Desktop installation repeatedly failed with Windows component-store related errors (`14098`).
- WSL installation and enabling `VirtualMachinePlatform` also failed with the same issue.
- Decision: proceed with a no-Docker path for backend testing.

Result:
- Docker path was abandoned for this test run.
- Native Redis path was selected.

### 2) Redis installed and reachable

- `winget search redis` identified `Redis.Redis`.
- Redis CLI check returned `PONG`.
- One manual foreground startup attempt reported a bind/start error, but reachable `PONG` confirmed an active Redis instance was already listening on `6379`.

Result:
- Redis dependency satisfied for backend startup.

### 3) Go toolchain setup issue, then fixed

- Initial `go` command failed (`CommandNotFoundException`).
- The previously assumed absolute Go path was not present.
- Go was installed/available and verified as `go1.26.3 windows/amd64`.

Result:
- Go runtime confirmed and usable.

### 4) First backend startup failed due to Redis command compatibility

Observed startup error:
- `panic: realtime bootstrap failed: ERR wrong number of arguments for 'hset' command`

Root cause:
- Installed Redis version was compatible with older command semantics.
- Project used multi-field `HSET`, which failed under this runtime.

Code fix applied:
- `Backend/internal/realtime/client.go`: multi-field hash write switched from `HSET` to `HMSET`.
- `Backend/internal/realtime/bid_lua.go`: multi-field `HSET` in Lua switched to `HMSET`.

Validation:
- Go test run succeeded after patch.

Result:
- Backend started successfully.

### 5) Backend service availability verified

Startup log confirmed:
- config loaded
- server started on `:8080`

API check:
- `GET /health` returned success payload with `service=auction-live-backend`, `port=8080`, `ok=true`.

Result:
- Service healthy and reachable.

### 6) Auth and permission matrix checks passed

Login checks:
- `POST /auth/login` for `viewer`, `anchor`, `admin` all succeeded.
- Returned roles matched expected identities.

Permission checks:
- Anonymous `GET /rooms` succeeded.
- Viewer `GET /users/me` succeeded.
- Anchor `GET /admin/stats/overview` succeeded.
- Viewer `GET /admin/stats/overview` returned `403 forbidden role` as expected.

Result:
- Authentication and role-based access control behaved correctly.

### 7) Session-state behavior validated (including expected failures)

Observed expected rejections:
- Bid attempt while session was not in `bidding` returned conflict (`session is not bidding`).
- Starting an already-ended session returned invalid state transition conflict.
- Bid attempt with invalid price step/minimum returned invalid bid price.

Interpretation:
- State-machine constraints were enforced correctly.

Result:
- Session lifecycle guardrails worked as designed.

### 8) Admin auction operations and bid flow passed end-to-end

Actions completed:
- Created new test item (`item-004`) with queued status.
- Activated next session (`session-004`) via queue-next.
- Started session successfully (`status=bidding`).
- Read current session and computed valid bid (`currentPrice + incrementStep`).
- Posted bid successfully:
  - `acceptedBidPrice=110`
  - `currentPrice=110`
  - `nextMinimumBid=120`
  - `isLeading=true`
- Fetched ranking and my-status:
  - ranking top entry was user `user-001`
  - my-status reflected `myHighestBid=110`, `myRank=1`, `isLeading=true`

Result:
- Core live-auction flow worked end-to-end.

### 9) Duplicate request idempotency validated

Goal:
- Re-submit the same bid payload with the same `requestId`.

Observations:
- Early attempts with curl formatting returned `400 invalid request body` (transport/body formatting issue, not business logic).
- Final duplicate verification with PowerShell web request path returned `HTTP 409`.

Interpretation:
- Duplicate request was rejected instead of being accepted as a new bid.
- This matches idempotency behavior expectations.

Result:
- Idempotency protection confirmed.

## Known Non-Blocking Notes

- `mysql init skipped error=mysql dsn is required` was logged during startup.
- This did not block this run because the backend can continue with in-memory data when MySQL is not configured.
- Some Chinese characters appeared as `????` in terminal output due to console encoding, not backend data corruption.

## Final Status

Overall result: PASS for no-Docker backend integration baseline.

Verified:
- Service health
- Auth and role checks
- State-machine enforcement
- Admin session operations
- Bid success path
- Ranking/my-status synchronization
- Duplicate request rejection

Not covered in this run:
- MySQL persistence verification
- WebSocket streaming assertions with dedicated client tooling
- Long-running load/performance verification

## Reusable Script

A reusable script was added for teammates:

- `Backend/scripts/integration_e2e_no_docker.ps1`

This script prints each request/response status and body, and outputs a final PASS/FAIL summary for quick verification.
