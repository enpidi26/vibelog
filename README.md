# VibeLog

[![CI](https://github.com/enpidi26/vibelog/actions/workflows/ci.yml/badge.svg)](https://github.com/enpidi26/vibelog/actions/workflows/ci.yml)
[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue.svg)](LICENSE)
[![Go 1.26](https://img.shields.io/badge/go-1.26-00ADD8.svg)](https://go.dev/)

VibeLog is a self-hosted prompt journal for AI-assisted development. Save the
exact prompt you sent, record the model, tool, result, rating, and notes, then
duplicate an entry as the next attempt. Export filtered entries as Markdown or
JSON prompt packs.

The entire application ships as **one Go binary**: the Vue SPA is embedded, and
SQLite uses a pure-Go driver. There is no CGO, no gcc, no Node.js, and no
separate web server at runtime.

## Table of contents

- [Features](#features)
- [Quick start](#quick-start)
- [Configuration](#configuration)
- [Email and GitHub login](#email-and-github-login)
- [Backup, restore, and upgrade](#backup-restore-and-upgrade)
- [Reverse proxy and HTTPS](#reverse-proxy-and-https)
- [API](#api)
- [Architecture](#architecture)
- [Development](#development)
- [Security](#security)
- [Project status and limitations](#project-status-and-limitations)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Prompt journal**: exact prompt text, model, tool, outcome, rating 1-5,
  result notes, and up to 10 tags.
- **Iterations**: duplicate an entry as the next attempt with an incremented
  iteration number and a link to its parent entry.
- **Projects and filters**: group entries by project and filter by text,
  project, model, tool, outcome, or tag.
- **Summary**: totals, outcome breakdown, average rating, top models, tools,
  projects, and tags.
- **Prompt packs**: export the filtered history as Markdown or JSON. Export is
  not limited to the visible page.
- **Authentication**: password login with bcrypt, magic links, and optional
  GitHub OAuth.
- **Self-host ready**: versioned SQLite migrations, automatic JWT secret,
  health checks, rate limiting, graceful shutdown, and a backup command.
- **Single binary**: frontend assets are embedded at build time.

## Quick start

### Docker Compose (recommended)

```bash
git clone https://github.com/enpidi26/vibelog.git
cd vibelog
cp .env.example .env
```

Edit `.env` and set `BASE_URL` to the exact URL users will open:

```env
BASE_URL=http://localhost:8090
PORT=8090
```

Then start the stack:

```bash
docker compose up -d --build
docker compose logs -f vibelog
```

Open `BASE_URL`, create the first account, then lock registration:

```env
ALLOW_REGISTRATION=false
```

```bash
docker compose up -d
```

Data lives in the `vibelog-data` volume at `/data`. The container runs as a
non-root user.

### Linux binary and systemd

The install script downloads the latest release, verifies its checksum, creates
a `vibelog` system user, and installs a systemd unit. It never overwrites an
existing database or environment file.

```bash
curl -fsSLO https://raw.githubusercontent.com/enpidi26/vibelog/main/scripts/install.sh
less install.sh                     # inspect before running
VIBELOG_DRY_RUN=1 bash install.sh   # download and verify only
sudo bash install.sh
```

Configure the public URL and restart:

```bash
sudo nano /etc/vibelog/vibelog.env
sudo systemctl restart vibelog
journalctl -u vibelog -f
```

Magic links in `console` mode are written to the journal.

### Standalone binary

Download the archive for your platform from
[Releases](https://github.com/enpidi26/vibelog/releases), extract it, and run:

```bash
APP_ENV=production \
BASE_URL=https://vibelog.example.com \
DATABASE_PATH=/var/lib/vibelog/vibelog.db \
./vibelog
```

### Build from source

Prerequisites: Go 1.26+ and Node.js 22+. A C compiler is not required.

```bash
make build
./vibelog
```

`make build` builds the frontend, embeds it under `internal/web/dist`, and
produces `./vibelog`.

## Configuration

Configuration is environment-based. The defaults below are safe for local
development; production requires a real `BASE_URL` and a strong secret or an
automatically generated one.

| Variable | Default | Description |
| --- | --- | --- |
| `APP_ENV` | `development` | `production` rejects weak or missing secrets |
| `BASE_URL` | `http://localhost:8080` | Public URL used for magic links, OAuth callbacks, and CORS |
| `HOST` | `127.0.0.1` | Listen address; Docker sets `0.0.0.0` |
| `PORT` | `8080` | Listen port inside the process; compose maps the host port |
| `DATABASE_PATH` | `data/vibelog.db` | SQLite database file |
| `JWT_SECRET` | generated | Optional explicit secret, at least 32 characters in production |
| `JWT_SECRET_PATH` | next to the database | File used when `JWT_SECRET` is unset |
| `JWT_EXPIRY_HOURS` | `24` | Session lifetime, 1-720 hours |
| `ALLOW_REGISTRATION` | `true` | Set `false` to close sign-up |
| `EMAIL_DRIVER` | `console` | `console`, `smtp`, `resend`, or `off` |
| `EMAIL_FROM` | `VibeLog <no-reply@localhost>` | Sender address |
| `SMTP_HOST` | empty | SMTP server hostname |
| `SMTP_PORT` | `587` | SMTP port |
| `SMTP_USERNAME` | empty | SMTP username |
| `SMTP_PASSWORD` | empty | SMTP password |
| `SMTP_TLS` | `starttls` | `starttls`, `implicit`, or `none`; 465 defaults to implicit |
| `RESEND_API_KEY` | empty | Required when `EMAIL_DRIVER=resend` |
| `GITHUB_CLIENT_ID` | empty | GitHub OAuth app client ID |
| `GITHUB_CLIENT_SECRET` | empty | GitHub OAuth app client secret |
| `TRUSTED_PROXIES` | loopback | Comma-separated IPs/CIDRs allowed to set forwarded headers |
| `BACKUP_KEEP` | `7` | Number of backups kept by `vibelog backup` |

`FRONTEND_URL` is still accepted as a fallback for `BASE_URL` on installations
created before the rename.

In Docker Compose, `PORT` in `.env` is the host port. The container always
listens on `8080`.

## Email and GitHub login

Password login is always available. Email and GitHub login are optional.

### Console mode

```env
EMAIL_DRIVER=console
```

The link is printed to the server log. The login page warns users when this mode
is active. This is useful for local use and first-run setup, not for a public
instance with untrusted users.

### SMTP

```env
EMAIL_DRIVER=smtp
EMAIL_FROM=VibeLog <vibelog@example.com>
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=vibelog@example.com
SMTP_PASSWORD=secret
SMTP_TLS=starttls
```

Use `SMTP_TLS=implicit` for port 465. `SMTP_TLS=none` is intended only for a
trusted local relay and is rejected when credentials are set.

### Resend

```env
EMAIL_DRIVER=resend
EMAIL_FROM=VibeLog <vibelog@example.com>
RESEND_API_KEY=re_...
```

The sending domain must be verified in Resend.

### GitHub OAuth

1. Create a GitHub OAuth App.
2. Set the homepage URL to `BASE_URL`.
3. Set the callback URL to `{BASE_URL}/api/auth/github/callback`.
4. Add `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET`, then restart.

## Backup, restore, and upgrade

### Backup

```bash
docker compose exec vibelog vibelog backup /data/backups
```

The command uses SQLite `VACUUM INTO`, so it is consistent while the server is
running. It keeps the newest `BACKUP_KEEP` archives (default 7) and writes them
with mode `0600`.

Schedule it from the host:

```cron
0 3 * * * cd /srv/vibelog && docker compose exec -T vibelog vibelog backup /data/backups
```

### Restore

```bash
docker compose stop
docker compose run --rm --entrypoint sh vibelog -c \
  'cp /data/backups/vibelog-YYYYMMDD-HHMMSS.db /data/vibelog.db && rm -f /data/vibelog.db-wal /data/vibelog.db-shm'
docker compose up -d
```

Also keep `jwt_secret` if you want existing sessions to remain valid.

### Upgrade

```bash
cd /srv/vibelog
git pull
docker compose exec vibelog vibelog backup /data/backups
docker compose up -d --build
```

Database migrations run automatically at startup and are recorded in
`schema_migrations`. Back up before every upgrade.

## Reverse proxy and HTTPS

Put nginx, Caddy, or another TLS terminator in front of the container. A sample
nginx configuration is in [`deploy/nginx-vibelog.conf`](deploy/nginx-vibelog.conf).

When a proxy is used, set:

```env
BASE_URL=https://vibelog.example.com
TRUSTED_PROXIES=172.16.0.0/12
```

`TRUSTED_PROXIES` must contain only the proxy addresses. Do not use
`0.0.0.0/0`; clients could spoof their IP and bypass rate limiting.

## API

All endpoints under `/api` use `Authorization: Bearer <token>` unless noted.
Errors use a stable JSON shape:

```json
{"error":{"code":"not_found","message":"entry not found"}}
```

| Method | Endpoint | Description |
| --- | --- | --- |
| GET | `/health` | Liveness plus database check |
| GET | `/livez` | Liveness without a database check |
| GET | `/api/meta` | Enabled auth methods and registration state |
| POST | `/api/auth/register` | Create an account |
| POST | `/api/auth/login` | Password login |
| GET | `/api/auth/me` | Current profile |
| GET | `/api/auth/github` | Start GitHub OAuth |
| GET | `/api/auth/github/callback` | GitHub OAuth callback |
| POST | `/api/auth/email/request` | Request a magic link |
| POST | `/api/auth/email/verify` | Exchange a magic-link token for a JWT |
| GET, POST | `/api/projects` | List and create projects |
| PATCH, DELETE | `/api/projects/{id}` | Update and delete a project |
| GET, POST | `/api/entries` | List and create entries |
| GET, PATCH, DELETE | `/api/entries/{id}` | Read, update, and delete an entry |
| POST | `/api/entries/{id}/duplicate` | Create the next iteration |
| GET | `/api/entries/summary` | Summary, optionally filtered by project |
| GET | `/api/entries/export` | Export `format=md` or `format=json` |

List filters: `q`, `project_id`, `model`, `tool`, `outcome`, `tag`, and
`limit` (maximum 200). `project_id=0` selects entries without a project.
Exports include the full filtered history.

## Architecture

```text
cmd/vibelog/          entrypoint, backup command, usage
internal/
  auth/               bcrypt, JWT, random tokens
  config/             environment loading and validation
  database/           SQLite, migrations, entities, backups
  domain/             core types and validation
  email/              console, SMTP, and Resend drivers
  export/             Markdown prompt packs
  httpapi/            routes, middleware, handlers, SPA serving
  netutil/            forwarded-header parsing
  ratelimit/          in-memory per-IP limiter
  web/                embedded frontend
frontend/src/         Vue 3 SPA
deploy/               systemd unit and reverse-proxy example
docs/                 self-hosting, product, verification
scripts/              install.sh and run.sh
```

The code has no `main` package outside `cmd/`. Tests live next to the packages
they cover.

## Development

```bash
make test    # Go tests and frontend typecheck
make build   # build the SPA, embed it, and build ./vibelog
make run     # local single-binary run at http://127.0.0.1:8080
```

Useful commands:

```bash
go test -count=1 ./...
cd frontend && npm run typecheck && npm run build
```

Release tags matching `v*` trigger the release workflow, which builds Linux
(amd64/arm64), macOS (amd64/arm64), and Windows binaries with checksums.

## Security

- Passwords are hashed with bcrypt.
- JWTs are HS256 with a fixed issuer and locked algorithm.
- All SQL is parameterized and every query is scoped to the authenticated user.
- CORS is restricted to `BASE_URL`; CSP and standard security headers are set.
- Auth endpoints are rate limited to 20 requests per minute per IP, API
  endpoints to 300 per minute.
- Magic-link tokens are single-use, expire after 15 minutes, and are verified
  with `POST` so email scanners cannot consume them.
- Request bodies are limited to 1 MB and unknown JSON fields are rejected.
- JWT secrets are generated with mode `0600` when not supplied.

Report vulnerabilities privately as described in [`SECURITY.md`](SECURITY.md).
Do not open a public issue for an exploitable bug.

## Project status and limitations

VibeLog is ready for **self-hosted personal and small-team use**. A managed cloud
offering is planned but not released.

Current limitations:

- SQLite is single-node; there is no multi-instance mode.
- The dashboard shows the latest 100 entries. Search and export still cover the
  full history.
- There is no password-change, account-deletion, or profile page yet.
- Rate limits are stored in memory per process.
- There is no import command; Markdown and JSON export are available.
- SMTP, Resend, and GitHub OAuth require real credentials and have not been
  verified in CI.
- The Docker image is built locally from source; no prebuilt image is published.

## Contributing

Contributions are welcome. Read [`CONTRIBUTING.md`](CONTRIBUTING.md) before
opening a pull request. By contributing, you agree to the license terms and the
relicensing grant described there.

## License

VibeLog is licensed under the **GNU Affero General Public License v3.0 only**
(AGPL-3.0-only). See [`LICENSE`](LICENSE) and [`COPYRIGHT`](COPYRIGHT).

If you run a modified version as a network service, the AGPL requires you to
offer the corresponding source. A commercial license is available from the
project owner for use cases that cannot comply with the AGPL.

The VibeLog name, logo, and visual identity are not licensed with the code. See
[`TRADEMARK.md`](TRADEMARK.md).

Code published before the AGPL switch remains available under its original MIT
terms.
