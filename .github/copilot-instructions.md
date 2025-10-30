Short, actionable guide for AI coding agents working on ManifoldTradingBot

- Project layout (big picture)
  - Two parallel implementations live in the repo:
    - Go implementation (legacy/primary bot): `go/` — contains `main.go`, `ManifoldApi/*` (API wrappers), `ModuleVelocity/*` (bot logic), `utils/*` (websocket, redis, helpers).
    - TypeScript/Deno implementation (newer tooling): `ts/` — Deno-based modules under `ts/src/` that mirror many Go API modules.

- Key files to read first
  - `go/main.go` — shows the Go runtime entry, how modules are enabled, and queue monitoring.
  - `ts/src/main.ts` — main entry point for the TS bot, showing how modules are initialized and started.

- Runtime & developer workflows (how to run/check TS code)
  - Deno is configured inside the `ts` folder. Run all deno commands using that working directory.

- Environment & permissions
  - Required environment variables visible in code: `MANIFOLD_API_KEY`, `MANIFOLD_API_DEBUG` (zod-validated in `ts/src/env.ts`).
  - Runtime flags typically required: `--allow-net` (API + websocket), `--allow-env` (read env vars). Add `--allow-read` / `--allow-write` only if tests or scripts need FS access.
