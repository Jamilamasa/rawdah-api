Rawdah API
==========

Rawdah is a backend API for a family-focused Islamic learning and habits app. It powers:
- Family accounts (parents, children, and trusted adult relatives)
- Tasks, rewards, and XP to gamify chores and learning
- Islamic content (hadiths, prophets, Quran lessons, quizzes)
- Messaging, requests, rants, and notifications
- WebSocket updates, web push notifications, and AI-assisted quiz generation

This document explains how to run the API locally, configure it for production, and deploy it on your own infrastructure.


Overview
--------

- Language: Go
- Framework: Gin
- Database: PostgreSQL
- Storage: Cloudflare R2 (S3-compatible)
- Email: Brevo (Sendinblue)
- Push: Web Push (VAPID)
- AI: OpenRouter (for quiz/content generation)

Main entrypoints:
- API server: [cmd/api/main.go](file:///Users/jamilamasa/Documents/programming/rawdah/rawdah-api/cmd/api/main.go)
- Migrations runner: [cmd/migrate/main.go](file:///Users/jamilamasa/Documents/programming/rawdah/rawdah-api/cmd/migrate/main.go)
- Configuration: [internal/config/config.go](file:///Users/jamilamasa/Documents/programming/rawdah/rawdah-api/internal/config/config.go)

The HTTP API is versioned under `/v1` and documented via the included Postman collection (`postman.json`).


Getting Started
---------------

### Prerequisites

- Go 1.22+
- PostgreSQL 13+ (local or hosted, e.g. Supabase)
- `make` (for the provided Makefile targets)
- Optional: Docker (for containerized deployment)


### Clone and configure

```bash
git clone https://github.com/rawdah/rawdah-api.git
cd rawdah-api

cp .env.example .env
```

Edit `.env` and set the values appropriate for your environment (see “Environment variables” below).


### Database setup and migrations

Create a PostgreSQL database and set `DATABASE_URL` in `.env`. For example:

```env
DATABASE_URL=postgresql://postgres:password@localhost:5432/rawdah
```

You can run migrations via the provided migrate command:

```bash
go run ./cmd/migrate
```

or, if you prefer `make`:

```bash
make migrate
```

Migrations are stored in [migrations/](file:///Users/jamilamasa/Documents/programming/rawdah/rawdah-api/migrations).


### Seeding Islamic content (optional)

There are helper scripts in [scripts/](file:///Users/jamilamasa/Documents/programming/rawdah/rawdah-api/scripts) to seed hadiths, prophets, and Quran content. These expect the relevant data files to be present or embedded according to your setup.

Typical usage:

```bash
go run ./scripts/seed_hadiths.go
go run ./scripts/seed_prophets.go
go run ./scripts/seed_quran.go
```

Run only the seeds you actually want; they are not required for the server to run.


Running the API Locally
-----------------------

With `.env` configured and migrations applied:

```bash
go run ./cmd/api
```

By default:
- The server listens on `PORT` (default `8080`)
- The health endpoint is `GET /health`
- The API root is `/v1`

Example signup request:

```bash
curl --location "http://localhost:8080/v1/auth/signup" \
  --header "Content-Type: application/json" \
  --data-raw '{
    "family_name": "test-family",
    "slug": "test-family",
    "name": "Parent Name",
    "parent_name": "Parent Name",
    "email": "parent@example.com",
    "password": "changeme123"
  }'
```

On success, you receive an access token, refresh cookie, and initial family/user details.


Testing
-------

The project includes basic tests and an isolation test suite.

Run all tests:

```bash
go test ./...
```

Isolation tests use a separate database (`TEST_DATABASE_URL`) and expect an empty schema:

```bash
TEST_DATABASE_URL=postgresql://postgres:password@localhost:5432/rawdah_test go test ./tests -run TestIsolation
```


Environment Variables
---------------------

The canonical reference is [.env.example](file:///Users/jamilamasa/Documents/programming/rawdah/rawdah-api/.env.example). This section explains what each group does and how to choose values.


### Core app

- `PORT`  
  Port the HTTP server listens on (e.g. `8080`).

- `ENV`  
  Environment name, typically `development` or `production`. Used to control logging and other behavior.

- `ALLOWED_ORIGINS`  
  Comma-separated list of origins allowed for CORS, e.g.:
  `https://app.rawdah.app,https://kids.rawdah.app,http://localhost:3001,http://localhost:3002`


### Database

- `DATABASE_URL`  
  PostgreSQL connection string. Example:
  `postgresql://postgres:password@localhost:5432/rawdah`

- `AUTO_MIGRATE`  
  If `true`, the app runs migrations automatically on startup. For production, you may prefer to run migrations explicitly and set this to `false`.

- `TEST_DATABASE_URL` (not in `.env.example` but used by tests)  
  Connection string for the isolation test database.


### Authentication (JWT)

- `JWT_ACCESS_SECRET`  
  Secret key for signing access tokens. Use a random string at least 64 characters long.

- `JWT_REFRESH_SECRET`  
  Secret key for signing refresh tokens. Must be a different random string, also at least 64 characters.

- `ACCESS_TOKEN_TTL`  
  Lifetime of access tokens (e.g. `15m`).

- `REFRESH_TOKEN_TTL`  
  Lifetime of refresh tokens (e.g. `168h` for 7 days).

- `CHILD_TOKEN_TTL`  
  Lifetime of child tokens (e.g. `4h`).

All TTL values are parsed using Go duration syntax (`1h`, `15m`, `24h`, etc.).


### Cloudflare R2 (file storage)

Used for user avatars, family logos, and other assets.

- `R2_ACCOUNT_ID`  
  Cloudflare account ID.

- `R2_ACCESS_KEY_ID`  
- `R2_SECRET_ACCESS_KEY`  
  Access keys for the R2 bucket.

- `R2_BUCKET`  
  Bucket name, e.g. `rawdah-assets`.

- `PRESIGN_EXPIRES_SECONDS`  
  Lifetime of pre-signed upload URLs in seconds (e.g. `600`).

If R2 is not configured, upload-related endpoints will return a 503-style error indicating that uploads are not available.


### Email (Brevo)

Used for account emails and notifications.

- `BREVO_API_KEY`  
  API key from Brevo.

- `BREVO_SENDER_EMAIL`  
  Sender email, e.g. `hello@rawdah.app`.

- `BREVO_SENDER_NAME`  
  Display name, e.g. `Rawdah`.

- `ADULT_PLATFORM_URL`  
- `KIDS_PLATFORM_URL`  
  Base URLs used in emails for adult and kids platforms (e.g. `https://app.rawdah.app` and `https://kids.rawdah.app`).


### AI (OpenRouter)

Used to generate quizzes and topic packs.

- `OPENROUTER_API_KEY`  
  OpenRouter API key.

- `OPENROUTER_MODEL`  
- `OPENROUTER_FALLBACK_MODEL`  
  Primary and fallback model IDs, e.g.:
  - `google/gemma-2-9b-it`
  - `mistralai/mistral-7b-instruct`

If AI credentials are missing, quiz-generation endpoints that rely on AI will return an error when called.


### Web Push (VAPID)

Used for web push notifications via the browser.

- `VAPID_PUBLIC_KEY`  
- `VAPID_PRIVATE_KEY`  
  VAPID key pair used to sign push notifications.

- `VAPID_SUBJECT`  
  Contact URI, e.g. `mailto:hello@rawdah.app`.

These values must match what your front-end uses to register push subscriptions.


### Cron / internal jobs

- `CRON_SECRET`  
  Shared secret used to protect any internal cron-style endpoints or tasks. Use a random string at least 64 characters long. When you add a cron job in your own infrastructure, send this secret as a header or query parameter, and have the server validate it before running the job.


API Overview
------------

The full API surface is captured in `postman.json`. Import this file into Postman or a similar client to explore all endpoints.

High-level groups:

- `/v1/auth`  
  Signup, signin, child signin, token refresh, password change.

- `/v1/family`  
  Family details, members, access control.

- `/v1/tasks`, `/v1/rewards`, `/v1/recurring-tasks`  
  Task and reward management for children, including recurring tasks.

- `/v1/hadiths`, `/v1/prophets`, `/v1/quran`, `/v1/lessons`, `/v1/quizzes`  
  Islamic content and quizzes (includes AI-generated quizzes).

- `/v1/messages`, `/v1/requests`, `/v1/rants`  
  Family communication features.

- `/v1/games`, `/v1/dashboard`  
  Game sessions and dashboard metrics.

- `/v1/notifications`, `/v1/push`  
  Notification listing and web push subscription endpoints.

- `/ws`  
  WebSocket endpoint for real-time events.

Most endpoints require a JWT access token in the `Authorization: Bearer <token>` header. The signup/signin endpoints return this token for you to use in subsequent requests.


Deployment
----------

You can deploy the Rawdah API in several ways. This section covers:
- Docker-based deployment
- Render-style deployment (using `render.yaml` as a reference)
- Bare-metal / VM deployment


### 1. Docker deployment

The repository includes a [Dockerfile](file:///Users/jamilamasa/Documents/programming/rawdah/rawdah-api/Dockerfile) that builds a minimal container image.

Build the image:

```bash
docker build -t rawdah-api .
```

Prepare a production `.env` file (or an environment in your orchestrator) with all required variables, then run:

```bash
docker run --rm \
  --env-file .env \
  -p 8080:8080 \
  rawdah-api
```

If your PostgreSQL database is also running in Docker, make sure `DATABASE_URL` points to the correct host (for example, `host.docker.internal` on Docker Desktop).

For production, you would typically:
- Build the image in CI
- Push it to a registry
- Run it in your orchestrator of choice (Kubernetes, ECS, Nomad, etc.)
- Configure readiness checks against `/health`
- Run migrations as a separate job or “release command” before starting the app


### 2. Render-style deployment

The repository includes a sample [render.yaml](file:///Users/jamilamasa/Documents/programming/rawdah/rawdah-api/render.yaml) with a basic configuration:

```yaml
services:
  - type: web
    name: rawdah-api
    env: go
    plan: starter
    buildCommand: go build -o rawdah-api ./cmd/api
    startCommand: ./rawdah-api
    healthCheckPath: /health
    envVars:
      - key: PORT
        value: 8080
      - key: ENV
        value: production
```

On Render or a similar platform:

1. Create a PostgreSQL database and note the connection URL.
2. Create a new web service from this repository or from a container image.
3. Set `PORT`, `ENV`, `DATABASE_URL`, and the rest of the environment variables in the dashboard.
4. Configure the health check to `/health`.
5. Add a deploy hook or separate job to run migrations (e.g. `go run ./cmd/migrate` or `make migrate`) before starting the service.


### 3. Bare-metal / VM deployment

On a Linux VM (Ubuntu/Debian/Rocky, etc.):

1. Install Go and PostgreSQL.
2. Clone the repository and build a binary:

   ```bash
   git clone https://github.com/rawdah/rawdah-api.git
   cd rawdah-api
   go build -o rawdah-api ./cmd/api
   ```

3. Copy the binary and `.env` to a directory like `/opt/rawdah-api`.
4. Create a systemd service, for example:

   ```ini
   [Unit]
   Description=Rawdah API
   After=network.target

   [Service]
   WorkingDirectory=/opt/rawdah-api
   ExecStart=/opt/rawdah-api/rawdah-api
   EnvironmentFile=/opt/rawdah-api/.env
   Restart=on-failure
   User=rawdah
   Group=rawdah

   [Install]
   WantedBy=multi-user.target
   ```

5. Reload systemd, enable, and start the service:

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable rawdah-api
   sudo systemctl start rawdah-api
   ```

6. Put the API behind a reverse proxy (Nginx, Caddy, Traefik) to terminate TLS and forward to the internal `PORT`.


CRON and background jobs
------------------------

The API is designed to be stateless and does not include an internal scheduler. Instead, you should:

1. Expose any cron-style endpoints you need (for example, to send reminders, clean up data, or recalculate metrics).
2. Protect them with `CRON_SECRET` (for example, by requiring a header `X-Cron-Secret` that matches the environment variable).
3. Use your infrastructure’s scheduler (Cron, Kubernetes CronJob, Render cron, etc.) to call those endpoints at the desired intervals.


Security and Production Notes
-----------------------------

- Do not commit `.env` or secrets to version control.
- Generate strong random strings for all secrets (`JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `CRON_SECRET`, etc.).
- Always run the API behind HTTPS in production.
- Ensure `ALLOWED_ORIGINS` is restricted to your actual front-end origins.
- Consider setting `AUTO_MIGRATE=false` in production and running migrations as a separate step.
- Keep your database backed up regularly.


Contributing
------------

If you want to extend the Rawdah API:

- Handlers live in [internal/handlers/](file:///Users/jamilamasa/Documents/programming/rawdah/rawdah-api/internal/handlers).
- Business logic lives in [internal/services/](file:///Users/jamilamasa/Documents/programming/rawdah/rawdah-api/internal/services).
- Data access lives in [internal/repository/](file:///Users/jamilamasa/Documents/programming/rawdah/rawdah-api/internal/repository).
- Shared models live in [internal/models/models.go](file:///Users/jamilamasa/Documents/programming/rawdah/rawdah-api/internal/models/models.go).

Follow existing patterns for routing, middleware, and error handling when adding new endpoints.

