# WasteCo Backend (`wasteco-be`)

Backend service and REST API for **WasteCo**, an application designed to prevent and manage household food waste through intelligent pantry inventory tracking, AI-powered zero-waste recipe recommendations (Google Gemini), surplus food donation matching, and environmental & economic impact analytics.

---

## 1. Tech Stack
- **Language & Runtime**: Go 1.26+
- **HTTP Framework**: [Fiber v3](https://github.com/gofiber/fiber/v3)
- **High-Performance JSON**: [Sonic](https://github.com/bytedance/sonic)
- **Database & ORM**: MySQL 8.0 with [GORM](https://gorm.io/)
- **Validation**: [Go Playground Validator v10](https://github.com/go-playground/validator/v10)
- **Logging**: [Zerolog](https://github.com/rs/zerolog) + [Lumberjack](https://github.com/natefinch/lumberjack)
- **AI Integration**: Google Gemini API (`gemini-flash-lite-latest`)
- **Object Storage**: Supabase Storage (`donations` bucket)
- **Authentication**: Stateless JWT (`golang-jwt/jwt/v5`) + SHA-256 API Key Verification

---

## 2. Quick Start with Docker Compose

Running the entire WasteCo backend stack (API server + MySQL database + automatic migrations + seed data) requires only a single command:

```bash
docker compose up --build -d
```

### What Docker Compose Starts:
1. **`wasteco-mysql`**: MySQL 8.0 server with persistent volume `wasteco_mysql_data`.
   - Exposed internally to the Docker network on port `3306`.
   - Exposed to host machine on port `3307` (configurable via `DB_HOST_PORT` in `.env`) to prevent conflicts with any pre-existing local MySQL instances.
   - Includes automatic health checks via `mysqladmin ping`.
2. **`wasteco-api`**: WasteCo Go backend HTTP service on port `8000`.
   - Automatically waits for `wasteco-mysql` health check to pass before boot.
   - Automatically executes database schema migrations (`/app/migrate`).
   - Automatically executes database seeders (`/app/seed`).
   - Starts HTTP server on port `8000`.

### Checking Status & Logs:
```bash
# Check container status
docker compose ps

# View live application logs
docker compose logs -f wasteco-api

# Stop all containers
docker compose down
```

---

## 3. Environment Configuration

Copy `.env.example` to `.env` if not already present:

```bash
cp .env.example .env
```

| Key | Description | Default / Example |
| :--- | :--- | :--- |
| `APP_NAME` | Name of application | `WasteCo` |
| `APP_ENV` | Environment mode (`development` or `production`) | `development` |
| `APP_HOST` | Host IP to bind server to | `0.0.0.0` |
| `APP_PORT` | Port exposed by server | `8000` |
| `API_KEY` | Global client secret passed in `x-api-key` header | `<api_key>` |
| `DB_NAME` | MySQL database name | `wasteco` |
| `DB_HOST` | Database host (`localhost` for host, `wasteco-db` for Docker) | `localhost` |
| `DB_PORT` | Database internal port | `3306` |
| `DB_USER` | MySQL username | `root` |
| `DB_PASS` | MySQL password | `root` |
| `DB_HOST_PORT`| MySQL port exposed to host machine | `3307` |
| `JWT_SECRET_KEY` | HMAC secret key for signing JWT tokens | `<secret_key>` |
| `JWT_EXPIRED_TIME` | JWT token validity duration | `2h` |
| `SUPABASE_URL` | Supabase project URL | `https://<id>.supabase.co` |
| `SUPABASE_SECRET`| Supabase service role secret | `<service_role_key>` |
| `SUPABASE_BUCKET`| Supabase bucket name for donation images | `donations` |
| `GEMINI_API_KEY` | Google Gemini API Key | `<gemini_api_key>` |
| `GEMINI_MODEL` | Gemini model variant | `gemini-flash-lite-latest` |

---

## 4. Test Credentials & Default Data

The seeder automatically inserts initial test users and sample data:

| User | Email | Password | Role / Area |
| :--- | :--- | :--- | :--- |
| Affan Maulana | `affan@wasteco.id` | `password123` | Dinoyo, Lowokwaru, Malang |
| Budi Santoso | `budi@wasteco.id` | `password123` | Dinoyo, Malang |
| Siti Aminah | `siti@wasteco.id` | `password123` | Lowokwaru, Malang |

### Authentication Headers:
All API requests require:
```http
x-api-key: <api_key>
```

Protected endpoints additionally require the Bearer JWT token from `POST /api/v1/auth/login`:
```http
Authorization: Bearer <jwt_access_token>
```

---

## 5. Running Manually (Without Docker)

If you prefer to run natively on your host machine:

```bash
# 1. Run database migrations
go run ./cmd/migration/migrate.go

# 2. Seed dummy data
go run ./cmd/seeder/seed.go

# 3. Start the server
go run ./cmd/api/main.go
```

To wipe and re-create all database tables:
```bash
go run ./cmd/migration/migrate.go -fresh
```

---

## 6. API Documentation

Detailed OpenAPI 3.1 specifications and Figma UI screen mappings are located in the [`docs/`](docs/) directory:
- [`docs/FIGMA_SPEC.md`](docs/FIGMA_SPEC.md): Complete Figma UI screen to API endpoint reference matrix.
- [`docs/BACKEND_ISSUES.md`](docs/BACKEND_ISSUES.md): Implementation tracker for all core backend modules.
- [`docs/wasteco_openapi.json`](docs/wasteco_openapi.json): Standard OpenAPI 3.1 specification (importable into Postman, Swagger UI, or Apidog).
