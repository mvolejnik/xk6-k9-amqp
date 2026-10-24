# CLAUDE.md

k6 `github.com/grafana/k6` extension (xk6) for AMQP 0-9-1 / RabbitMQ, built on `github.com/rabbitmq/amqp091-go`.
Module `github.com/mvolejnik/xk6-k9-amqp`, Go package `k9amqp`, all code in repo root (flat, ~1k LOC).
Main selling point vs. deprecated grafana/xk6-amqp: shared connection + channel pool (low channel churn) + active development.

## Files

| File | Purpose |
|---|---|
| `module.go` | Registers JS modules `k6/x/k9amqp` (root, exports `K9amqp`), `k6/x/k9amqp/queue` (`Queue`), `k6/x/k9amqp/exchange` (`Exchange`). Metrics registered per VU in `NewModuleInstance`. |
| `k9amqp.go` | `K9amqp.XClient` (JS `new k9amqp.Client(amqpOpts, poolOpts)`), `Client` methods `Publish`, `Get`, `Consume`, `Listen`, `Teardown`, metric reporting, tag helpers. |
| `amqp.go` | `AmqpClient` (connections) + `AmqpPool` (buffered chan of `*amqp.Channel`, round-robin over connections), option defaults (`init()`), `Connect()`. |
| `queue.go` / `exchange.go` | Declare/Delete/Bind/Unbind(/Purge) — each takes `(client *Client, opts XxxOptions)`. |
| `types.go` | All option/response structs. |
| `metrics.go` | k6 metrics: `amqp_pub_sent`, `amqp_pub_failed`, `amqp_pub_duration` (Trend, Time), `amqp_sub_received`, `amqp_sub_no_delivery`, `amqp_sub_failed`. No subscriber duration (by design). Tags: `endpoint`, `exchange`, `routing_key`, `scenario`. Builtin `data_sent`/`data_received` are also pushed (see below). |
| `index.d.ts` + `package.json` | Hand-written TypeScript typings for the JS API — **update when changing Go API/types**. |
| `examples/*.js` | Runnable k6 scripts (`simple.js` is the CI smoke test). Keep in sync with API. |
| `docker/` | `k6-test` (CI: RabbitMQ + locally built k6), `k6-test-release` (published image `mvolejnik/k6-with-amqp`), `k6-test-with-grafana` (Prometheus/Grafana, see `Prometheus.md`), `k6-with-amqp` (release image Dockerfile). |
| `README.md`, `Prometheus.md` | User docs; README also lists tested k6 versions. |

## Key design facts (non-obvious)

- **Global singleton**: `amqpClient` + `sync.Once` in `k9amqp.go` — one `AmqpClient` per k6 process, shared by all VUs. Options from the *first* `new Client(...)` win; later calls ignore their arguments (`new k9amqp.Client()` with no args is fine in setup/teardown).
- Pool sizing: `connections = channels_cache_size / channels_per_conn` (min 1). Defaults: `channels_per_conn=2`; host `localhost:5672`, vhost `/`, `guest/guest`.
- Publisher confirms are **on by default** (opt-out via pool option `no_confirm: true`): `AmqpPool.channel()` puts every new channel in confirm mode; `publish()` uses `PublishWithDeferredConfirmWithContext` + `WaitContext`. A nack returns `ErrPublishNack` and the channel goes back to the pool (it isn't closed).
- `data_sent`/`data_received` = bytes on the wire, counted globally by `countingConn` (wraps `net.Conn` via `amqp.Config.Dial` in `Connect()`, atomic counters on `AmqpClient`). The delta is pushed (`Swap(0)`) by whichever VU calls `Publish`/`Get`/`Consume`/`Teardown` next, so it isn't tracked per VU (on purpose). It includes `Listen` connections.
- Metric helpers: `endpointTags` (nil outside VU context, so it skips pushing), `sample`, `push`. `Listen` has no metrics (on purpose).
- Channel lifecycle pattern (copy-pasted in every op): `channels.get()` → op → `defer`: on success `put()` back to pool (closed if pool full), on error `channel.Close()`. Keep this pattern for new ops.
- `Listen` is the exception: opens its **own** connection+channel and consumes in a goroutine; returning an error from the JS listener stops it.
- `Consume` is non-blocking: drains up to `opts.Size` already-buffered deliveries, then cancels the consumer.
- JS naming: k6/sobek maps Go struct fields to **snake_case** (`ChannelsPerConn` → `channels_per_conn`, `AutoAck` → `auto_ack`) and methods to camelCase (`Publish` → `publish`, `Delete` → `delete`). `amqp.Publishing`/`amqp.Delivery` are passed through directly (responses keep Go-case fields like `Delivery.Body`, `Ok`, `Error`).
- Logging uses `log/slog`; errors are both returned (thrown in JS) and logged.
- `go.mod` has a leftover `replace .../k9amqp => ./k9amqp` (dir doesn't exist; harmless).
- No Go unit tests exist; verification = lint + xk6 build + running `examples/simple.js` against RabbitMQ.

## Commands

```bash
go build ./... && go vet ./...                 # quick compile check (go 1.25+ in go.mod; local/CI use 1.27.x)
golangci-lint run --timeout=5m                 # CI lint (v2.15.0); may not be installed locally
./build.sh                                     # go build + `xk6 build latest --with xk6-k9-amqp=.` -> ./k6 (gitignored, ~50MB)
# without xk6 installed:
docker run --rm -v "$PWD:/xk6" -w /xk6 -u $(id -u):$(id -g) grafana/xk6:latest build latest --with xk6-k9-amqp=.
# smoke test:
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:4.4.0-management-alpine
./k6 run examples/simple.js                    # env: AMQP_HOST/PORT/VHOST/USERNAME/PASSWORD, AMQP_CHANNELS_PER_CONN, AMQP_CACHE_SIZE
```

## CI / maintenance

- `.github/workflows/build.yaml` (push to master, PRs): golangci-lint → xk6 build → docker image → `docker/k6-test` compose → `k6 run simple.js`.
- `.github/workflows/xk6-build-release.yaml` (manual/weekly): builds a tagged release with pinned `XK6_VERSION` and default k6/extension versions — bump those defaults on release.
- Renovate (`renovate.json`) groups gomod, GitHub Actions, Go toolchain, Docker tags. Version-bump branches are named like `47-pipeline-versions-update`.
- When bumping k6: update `go.k6.io/k6/v2` in go.mod, `grafana/k6:<ver>` in `docker/*/Dockerfile*`, release workflow defaults, and the compatibility list in README.
- PRs target `master`.
