# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

RouteX is a multi-user proxy admin platform: **Xray-core entry + SOCKS5/HTTP outbound pool + Go/React control plane**. Admins create users in the web UI → backend pushes them to Xray via gRPC → users connect via VLESS+Reality → traffic exits through a residential SOCKS5 chosen by the router service. The original requirements doc is in `需求.md`; the canonical 4-term vocabulary is **控制台 / 节点 / 出口 / 用户** (see [memory](/Users/zhenxianhuang/.claude/projects/-Users-zhenxianhuang-Desktop-CurrentWork-RouteX/memory/feedback_terminology.md)).

## Dev workflow — DO NOT bypass these

The user has a hard requirement: **everything starts/stops via `./start.sh` and `./stop.sh` from the repo root**. Do not invent new entry points or tell them to run `go run ./cmd/server` directly.

```bash
./start.sh    # docker (pg/redis/xray) + Go backend + React frontend
              # Ctrl+C exits log tail only; services keep running
./stop.sh     # actually stops everything (preserves volumes)
make reset    # nuke .env, keep DB data
make nuke     # nuke .env + DB volumes (destructive)
```

`start.sh` on first run: copies `.env.example` → `.env`, runs `docker run --rm teddysun/xray xray x25519` to generate the Reality keypair, sed-writes the PrivateKey into `deploy/xray/config.json` (gitignored), writes the PublicKey to `.env` as `REALITY_PUBLIC_KEY`.

## Ports — non-default range

The user moved every host-side port off the conventional defaults to avoid conflicts. Memorize this — older docs/memories may still show old ports:

| Port | Service |
|------|---------|
| 8890 | Vite dev (open browser here) |
| 8891 | Go backend API |
| 8892 | PostgreSQL |
| 8893 | Redis |
| 8894 | Xray gRPC API |
| 8895 | Xray VLESS+Reality inbound |

`.env.example` and `config.go` defaults are aligned with this range. `scripts/check-env.sh` warns when `.env` drifts from `.env.example`.

## Architecture you must understand

**One-Xray limitation.** The backend keeps a single `xray.Client` pointed at `XRAY_API_HOST:XRAY_API_PORT`. AddUser / RemoveUser / outbound swaps only hit that one Xray. The `nodes` table catalogs would-be entry nodes, but they're not remotely controlled — agents only report heartbeats. True multi-node = chunky deferred work.

**Routing path inside Xray.** `deploy/xray/config.json` declares three outbounds: `direct`, `block`, `proxy-out` (initially `freedom`). The routing rule sends `vless-in` → `proxy-out`. `service.RouterService` dynamically replaces `proxy-out` via `RemoveOutbound` + `AddSocksOutbound`/`AddHTTPOutbound` based on the healthiest active provider; falls back to freedom when none available.

**Schedulers (all in `server/internal/scheduler/`).**
- `TrafficCollector` — every `TRAFFIC_POLL_SECONDS` (60s default): for each enabled user, `xray.QueryUserTraffic(reset=true)` → write `user_traffic` row + UPDATE `used_traffic` → if over quota or expired, call `UserService.Disable`.
- `ProviderHealthChecker` — every `PROVIDER_HEALTH_SECONDS` (120s default): test each enabled provider via HTTPS GET through it (`cloudflare.com/cdn-cgi/trace`), record latency/error. After each tick, calls `RouterService.SyncBest` to swap `proxy-out` if needed. 5 consecutive failures auto-disables.
- `XrayWatcher` — every 15s: `Ping` Xray gRPC. On "down → up" transition, calls `UserService.SyncAll` (re-pushes all DB users to Xray) and `RouterService.SyncBest`. Handles `docker compose restart xray` cleanly.

**Two parallel auth schemes.** Admin endpoints use JWT (HS256 via `golang-jwt/v5`), middleware in `internal/middleware/jwt.go`. Node heartbeat uses `X-Node-Token` header matched against `nodes.auth_token`, middleware in `internal/middleware/node_auth.go`. Subscription endpoint uses the URL token itself (`/v1/sub/:token`, no auth header).

**Default admin bootstrap.** On startup, if `admins` table is empty, `auth.EnsureDefaultAdmin` creates `admin/<20-char random>` and logs the plaintext password with the literal prefix `BOOTSTRAP_ADMIN`. `start.sh` truncates `logs/server.log` on each run, so the line is **lost after restart** — user must `grep BOOTSTRAP_ADMIN logs/server.log` before stopping.

## Subscription content negotiation

`GET /v1/sub/:token` checks `Accept` header:
- `text/html` (browser) → renders `server/internal/api/templates/sub.html` (embed.FS) with QR + copy buttons + client download grid.
- otherwise (proxy clients) → returns base64-encoded vless:// list.

Same URL serves both. Per-node URLs use `Node.PublicHost/PublicPort`, falling back to `.env` `PUBLIC_HOST/PUBLIC_PORT` when blank. If no nodes exist in DB, a single "default" link is synthesized from env.

## File layout shortcuts

- `server/cmd/server/main.go` — wires all services + schedulers + router (~90 lines, read it first when adding cross-cutting features)
- `server/internal/api/router.go` — all routes registered here, including which group has which middleware
- `server/internal/xray/{client,user,stats,outbound}.go` — wrappers around xray-core's `HandlerService` / `StatsService` gRPC. **Xray v26 changed proto schemas**: `socks.ClientConfig.Server` is `*ServerEndpoint` (single, not slice), `ClientConfig.Version` is gone (always SOCKS5). Easy to get wrong.
- `web/src/components/ui/*` — hand-written shadcn primitives, not from CLI. Match the shadcn pattern but live in-tree.

## Verification before commit

- Backend: `cd server && go build ./...` (no test suite yet)
- Frontend: `cd web && npx tsc -b` (silent = pass; full prod build via `npm run build`)
- After running `tsc -b`, **delete** `web/vite.config.{js,d.ts}` and `web/*.tsbuildinfo` before committing — they're tsc-emitted artifacts gitignored at root level

## Deployment

Two paths exist:
- **Dev (this machine)** — `./start.sh`. Backend runs on host via `go run`, talks to Xray in docker via `127.0.0.1:8894`.
- **Prod (VPS, registry images)** — the server runs the stack from pulled images, no repo checkout. Push to `main`/`master` (or any tag) → GitHub Actions builds and pushes the single all-in-one image `registry.cn-hongkong.aliyuncs.com/mewtwo_zero/routex:latest`. On the server, `deploy/registry/restart.sh` pulls + recreates. The image packs nginx (frontend + `/v1` reverse-proxy to `127.0.0.1:8891`) + the Go backend under supervisord; postgres/redis/xray are stock images in `deploy/registry/docker-compose.yml`. See `deploy/registry/README.md`. Requires `ACR_USERNAME`/`ACR_PASSWORD` GitHub secrets.

## Git hygiene

Root-level images are auto-ignored (`/[a-z]*.png` etc.) because IDE clipboard pastes occasionally produce stray PNGs that got committed twice already. Don't add new `.png` at repo root without `git rm --cached`'ing if it sneaks in.

Commit messages: imperative `<type>(<scope>): <subject>` first line under ~70 chars, blank line, then a body that explains the *why* (not the *what*; the diff shows what). Co-authored-by trailer is the project convention.
