# React-Golang Starter Kit

A modern, production-ready full-stack starter template combining **React 19** (with TanStack Router & Query) and **Go 1.25** (with Chi, GORM, JWT). Built for rapid SaaS development with Docker, featuring multi-tenant organizations, real-time WebSockets, i18n, Prometheus observability, Stripe billing, and comprehensive CI/CD with smart change detection.

📚 **[Full Documentation](docs/README.md)**

---

## 📋 Quick Reference

| I want to...                    | Command / Link                                        |
| ------------------------------- | ----------------------------------------------------- |
| Start development               | `make dev`                                            |
| Run all tests                   | `make test`                                           |
| Run frontend tests only         | `cd frontend && npm run test:fast`                    |
| Run backend tests only          | `cd backend && go test ./...`                         |
| View logs                       | `make logs`                                           |
| Reset database                  | `make db-reset`                                       |
| Check service health            | `make health`                                         |
| See all commands                | `make help`                                           |
| Deploy to production            | `make prod`                                           |
| Start monitoring stack          | `make observability-up`                               |

📖 **[Complete Makefile Reference →](#-makefile-reference)**

---

## 🚀 Quick Start

Get up and running in under 5 minutes:

### Docker (Recommended)

```bash
# Clone and configure
git clone https://github.com/DailyDisco/react-golang-starter-kit.git
cd react-golang-starter-kit
cp .env.example .env
# Edit .env and set JWT_SECRET (required)

# Start all services
docker compose up -d

# View logs
docker compose logs -f
```

**Your app is now running:**

- Frontend: [http://localhost:5193](http://localhost:5193)
- Backend API: [http://localhost:8080](http://localhost:8080)
- API Health: [http://localhost:8080/health](http://localhost:8080/health)

📖 **[Complete Docker Guide →](docs/DOCKER_SETUP.md)**

### Local Development

**Prerequisites:** Go 1.25+, Node.js 20+ (LTS), PostgreSQL 17+

```bash
# Configure environment
cp .env.example .env
# Edit .env with database credentials and JWT_SECRET

# Terminal 1: Backend
cd backend
go mod tidy
go run cmd/main.go

# Terminal 2: Frontend
cd frontend
npm install
npm run dev
```

---

## ✨ Features

### Frontend Stack

- ⚛️ **React 19** with TypeScript (strict mode)
- ⚡ **Vite 7** - Lightning-fast builds and HMR
- 🛣️ **TanStack Router** - Type-safe, file-based routing
- 🔄 **TanStack Query** - Server state management
- 📊 **TanStack Table** - Powerful data tables with sorting, filtering, pagination
- 🐻 **Zustand** - Client state management
- 🎨 **TailwindCSS 4 + ShadCN UI** - 65+ beautiful, accessible components
- 🎬 **Framer Motion** - Smooth animations and transitions
- 📈 **Recharts** - Charts and data visualization
- 🎠 **Embla Carousel** - Touch-friendly carousels
- 🌐 **i18next** - Multi-language support (EN/ES included)
- 📡 **WebSocket** - Real-time notifications and data sync
- 🔔 **Sonner** - Toast notifications with animations
- 🧪 **Vitest + Playwright** - Unit and E2E testing

### Backend Stack

- 🐹 **Go 1.25+** with Chi router
- 🗄️ **GORM + PostgreSQL 17** - Powerful ORM and database
- 🔐 **JWT Authentication** - Secure token-based auth with 2FA (TOTP)
- 🏢 **Multi-Tenant Organizations** - Teams with roles and invitations
- 👥 **Role-Based Access Control (RBAC)** - 4 permission levels
- 📤 **File Upload System** - AWS S3 or database storage with fallback
- 🛡️ **Rate Limiting** - Configurable IP and user-based protection
- 📧 **Email Service** - SMTP with styled HTML templates
- 💳 **Stripe Payments** - Subscriptions, Checkout, Customer Portal
- ⚡ **Background Jobs** - River (PostgreSQL-backed job queue)
- 📡 **WebSocket Server** - Real-time push notifications
- 🔄 **Database Migrations** - golang-migrate with CI validation
- 🤖 **AI Integration** - Gemini with chat, streaming, vision, embeddings
- 📝 **Structured Logging** - zerolog with JSON output
- 🚨 **Error Tracking** - Sentry integration with stack traces
- 🔑 **API Key Management** - Encrypted key storage with CRUD operations
- 🚩 **Feature Flags** - Runtime toggles with user overrides and plan enforcement
- 📋 **Audit Logging** - Compliance-ready action tracking
- 📦 **Data Export** - GDPR-compliant user data export with background jobs
- 🗑️ **Account Deletion** - Grace period workflow with cancellation

### Observability & DevOps

- 📊 **Prometheus Metrics** - HTTP, DB, cache, WebSocket, auth metrics
- 📈 **Grafana Dashboards** - Pre-configured monitoring
- 🚨 **Sentry** - Error tracking for frontend and backend
- 📉 **Vercel Analytics** - Frontend performance analytics
- ⚡ **Web Vitals** - Core Web Vitals monitoring
- 🎛️ **Feature Flags** - Runtime feature toggles via environment variables
- 🐳 **Docker Compose** - Development and production ready
- 📦 **Multi-stage builds** - Optimized Alpine & Distroless images
- 🔧 **Environment-based config** - Comprehensive .env support
- ✅ **GitHub Actions CI/CD** - Smart change detection, parallel checks, E2E tests

📖 **[Detailed Features Guide →](docs/FEATURES.md)**

---

## 📂 Project Structure

```text
react-golang-starter-kit/
├── backend/              # Go API server
│   ├── cmd/             # Application entry point
│   ├── migrations/      # SQL database migrations
│   ├── internal/        # Private application code
│   │   ├── ai/          # Gemini AI service
│   │   ├── auth/        # JWT authentication & OAuth
│   │   ├── cache/       # Dragonfly/memory caching
│   │   ├── config/      # Configuration management
│   │   ├── database/    # Database connection & migrations
│   │   ├── email/       # SMTP email service
│   │   ├── handlers/    # HTTP request handlers (14 modules, 100+ endpoints)
│   │   ├── jobs/        # River background jobs (export, cleanup, retention)
│   │   ├── middleware/  # Chi middleware (12 middleware implementations)
│   │   ├── models/      # GORM models (27+ models)
│   │   ├── observability/ # Prometheus metrics
│   │   ├── pagination/  # Pagination utilities
│   │   ├── ratelimit/   # Rate limiting (4 tiers: IP/User/Auth/API)
│   │   ├── repository/  # Data access layer (interfaces + implementations)
│   │   ├── response/    # Standardized response helpers
│   │   ├── sanitize/    # Input sanitization utilities
│   │   ├── services/    # Business logic layer (9 core services)
│   │   ├── storage/     # File storage (S3/DB)
│   │   ├── stripe/      # Stripe payments
│   │   ├── validation/  # Request validation logic
│   │   └── websocket/   # Real-time WebSocket hub
│   ├── docs/            # Swagger documentation
│   └── scripts/         # Utility scripts
│
├── frontend/            # React application
│   ├── app/            # Application code
│   │   ├── components/ # Reusable UI components (65+ including ShadCN)
│   │   ├── hooks/      # Custom React hooks (25+ hooks)
│   │   │   ├── queries/    # TanStack Query hooks
│   │   │   └── mutations/  # Mutation hooks with cache invalidation
│   │   ├── i18n/       # Internationalization (EN/ES)
│   │   ├── layouts/    # Layout components (Admin, Settings, Dashboard)
│   │   ├── lib/        # Utilities (query keys, guards, optimistic updates)
│   │   ├── routes/     # TanStack Router pages (file-based)
│   │   ├── services/   # API service layer
│   │   └── stores/     # Zustand state management (5 stores)
│   └── public/         # Static assets
│
├── docker/             # Docker compose files
│   ├── compose.yml     # Base services
│   ├── compose.dev.yml # Development overrides
│   ├── compose.prod.yml # Production (blue-green)
│   ├── compose.test.yml # Test environment
│   └── compose.observability.yml # Monitoring stack
│
├── infra/              # Infrastructure configs
│   ├── grafana/        # Grafana dashboards
│   └── prometheus/     # Prometheus configuration
│
├── scripts/            # Deployment and utility scripts
└── Makefile            # Development commands
```

📖 **[Backend Architecture →](backend/README.md)** | **[Frontend Development →](docs/FRONTEND_GUIDE.md)**

---

## 📚 Documentation

### Getting Started

- **[Complete Documentation Hub](docs/README.md)** - Start here for all guides
- **[Docker Setup Guide](docs/DOCKER_SETUP.md)** - Development and production Docker
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Deploy to Vercel, Railway, VPS, AWS

### Development

- **[Frontend Development](docs/FRONTEND_GUIDE.md)** - React, Vite, TanStack, Testing
- **[Backend Development](backend/README.md)** - Go architecture, GORM, API design
- **[Features Documentation](docs/FEATURES.md)** - Auth, RBAC, File uploads, Rate limiting

### Configuration

- **[Environment Variables](.env.example)** - All configuration (required + optional features)

---

## 🚀 Deployment

Deploy to production in minutes with your preferred platform:

### Vercel + Railway (Easiest)

1. Create PostgreSQL on [Railway.app](https://railway.app)
2. Deploy backend to Railway (`backend/` folder)
3. Deploy frontend to [Vercel](https://vercel.com) (`frontend/` folder)
4. Set `VITE_API_URL` in Vercel to Railway backend URL

**Time:** 15-30 minutes | **Cost:** $0-10/month

### Docker + VPS (Most Control)

```bash
docker compose -f docker/compose.yml -f docker/compose.prod.yml up -d
```

**Time:** 30-60 minutes | **Cost:** $5-20/month

### Blue-Green Deployment (Zero Downtime)

The production Docker setup supports blue-green deployments for zero-downtime updates:

```bash
make prod              # Deploy with zero downtime (swaps blue/green)
make prod-status       # Check current deployment status
make rollback          # Rollback to previous environment
```

**How it works:**

1. New version deploys to inactive environment (blue or green)
2. Health checks validate the new deployment
3. Traffic switches to new environment
4. Previous environment kept for instant rollback

📖 **[Complete Deployment Guide →](docs/DEPLOYMENT.md)**

---

## 🧪 Testing

### Frontend Tests

```bash
cd frontend
npm test              # Watch mode
npm run test:fast     # Run once (CI)
npm run test:coverage # Coverage report
npm run test:e2e      # Playwright E2E tests
```

### Backend Tests

```bash
cd backend
go test ./...         # Run all tests
go test -cover ./...  # With coverage
go test -race ./...   # With race detection
```

### Integration Tests

Integration tests run against real PostgreSQL using testcontainers-go:

```bash
# Using Makefile (starts test DB automatically)
make test-integration

# Or manually with environment variables
cd backend
INTEGRATION_TEST=true go test ./internal/services/... -v
```

**Test coverage includes:** Organization CRUD, session management, settings, usage tracking, 2FA, file storage.

### CI Pipeline

The GitHub Actions workflow includes:

- **Change Detection** - Only runs affected services (frontend/backend)
- **Parallel Checks** - Lint, typecheck, format, security audit run concurrently
- **Test Coverage** - Unit tests with coverage reports
- **E2E Testing** - Playwright tests on Chromium
- **Migration Validation** - Tests up/down/up cycle on real PostgreSQL
- **Swagger Validation** - Ensures API docs stay current
- **Compatibility Matrix** - Node 22 and Go 1.25 verification on master

📖 **[Testing Guide →](docs/FRONTEND_GUIDE.md#testing)**

---

## 🛠️ Available Scripts

### Build Operations

```bash
./docker-build.sh dev         # Build development images
./docker-build.sh prod        # Build production images
./docker-build.sh clean       # Clean Docker resources
```

### Runtime Operations

```bash
make dev              # Start development environment
make prod             # Start production environment
make logs             # View logs from all services
make stop             # Stop all containers
make clean            # Clean up containers & volumes
```

### Database Migrations

```bash
cd backend
make migrate-up       # Run all pending migrations
make migrate-down     # Rollback last migration
make migrate-create name=add_feature  # Create new migration
make migrate-version  # Show current version
```

#### Migration Best Practices

**File Naming:** Migrations use sequential numbering: `000001_init.up.sql`, `000001_init.down.sql`

**Creating Migrations:**

```bash
# Creates 000004_add_user_preferences.up.sql and .down.sql
cd backend && make migrate-create name=add_user_preferences
```

**Writing Safe Migrations:**

- Always write both `up.sql` and `down.sql` (reversible)
- Use `IF NOT EXISTS` / `IF EXISTS` for idempotency
- Add indexes for columns used in WHERE, JOIN, ORDER BY
- Test locally before deploying: `make migrate-down && make migrate-up`

**Example Migration:**

```sql
-- 000004_add_user_preferences.up.sql
CREATE TABLE IF NOT EXISTS user_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    theme VARCHAR(20) DEFAULT 'system',
    notifications_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id ON user_preferences(user_id);

-- 000004_add_user_preferences.down.sql
DROP TABLE IF EXISTS user_preferences;
```

**CI Validation:** The GitHub Actions pipeline validates migrations by running `up → down → up` against a real PostgreSQL instance.

📖 **[Migration Patterns →](docs/FEATURES.md#database-migrations)**

### Code Quality

```bash
# Frontend
npm run lint          # ESLint check
npm run lint:fix      # Fix ESLint issues
npm run prettier:check # Check formatting
npm run format        # Fix formatting
npm run typecheck     # TypeScript check

# Backend
make format-backend   # Format Go code (go fmt)
make vet-backend      # Static analysis (go vet)
# golangci-lint and govulncheck run in CI
```

### Dependency Management

```bash
make sync-deps        # Sync all dependencies (npm install + go mod tidy)
make sync-frontend    # Sync frontend dependencies only
make sync-backend     # Sync backend dependencies only
```

### Test Coverage

```bash
make coverage         # Generate all coverage reports
make coverage-html    # Generate HTML coverage reports
make coverage-check   # Check coverage meets 70% threshold
make test-services-up # Start all test services (DB, Redis, LocalStack, Mailpit)
make test-services-down # Stop all test services
make test-clean       # Clean test artifacts
```

### Project Setup

```bash
make setup            # Initial setup - copy env file
make configure-features # Interactive feature configuration wizard
make init             # Initialize a new project from this template
```

---

## 📜 Makefile Reference

Complete list of available `make` targets organized by category:

#### Dev Environment

| Command          | Description                                      |
| ---------------- | ------------------------------------------------ |
| `make dev`       | Start development environment (with auto-seed)   |
| `make dev-fast`  | Start dev without rebuild (fastest)              |
| `make dev-fresh` | Start with fresh database                        |
| `make restart`   | Restart all services                             |
| `make logs`      | View logs from all services                      |
| `make tail`      | View last 100 lines of logs                      |
| `make status`    | Show status of all services                      |
| `make health`    | Check health of all services                     |

#### Production Deployment

| Command            | Description                        |
| ------------------ | ---------------------------------- |
| `make prod`        | Deploy with zero downtime          |
| `make prod-status` | Show deployment status             |
| `make rollback`    | Rollback to previous environment   |
| `make prod-build`  | Build production images            |

#### Database Management

| Command         | Description                     |
| --------------- | ------------------------------- |
| `make db-reset` | Reset database (deletes data)   |
| `make seed`     | Seed with test data             |
| `make shell-db` | Access PostgreSQL shell         |

#### Test Commands

| Command                     | Description                          |
| --------------------------- | ------------------------------------ |
| `make test`                 | Run all tests                        |
| `make test-backend`         | Run backend tests                    |
| `make test-frontend`        | Run frontend tests                   |
| `make test-integration`     | Run integration tests                |
| `make test-e2e`             | Run Playwright E2E tests             |
| `make test-services-up`     | Start test services (DB, Redis, etc) |
| `make test-services-down`   | Stop test services                   |
| `make coverage`             | Generate all coverage reports        |
| `make coverage-html`        | Generate HTML coverage reports       |
| `make coverage-check`       | Check 70% coverage threshold         |

#### Quality & Dependencies

| Command               | Description                  |
| --------------------- | ---------------------------- |
| `make format-backend` | Format Go code               |
| `make sync-deps`      | Sync all dependencies        |
| `make sync-frontend`  | Sync frontend dependencies   |
| `make sync-backend`   | Sync backend dependencies    |

#### Monitoring Stack

| Command                   | Description                    |
| ------------------------- | ------------------------------ |
| `make observability-up`   | Start Prometheus + Grafana     |
| `make observability-down` | Stop monitoring stack          |
| `make grafana-logs`       | View Grafana logs              |
| `make prometheus-logs`    | View Prometheus logs           |

#### External Deployments

| Command                | Description                   |
| ---------------------- | ----------------------------- |
| `make deploy-vercel`   | Deploy frontend to Vercel     |
| `make deploy-railway`  | Deploy backend to Railway     |
| `make frontend-build`  | Build frontend for deployment |

#### Setup & Cleanup

| Command                   | Description                        |
| ------------------------- | ---------------------------------- |
| `make setup`              | Initial setup (copy env file)      |
| `make configure-features` | Interactive feature wizard         |
| `make init`               | Initialize new project             |
| `make clean`              | Clean containers, volumes, images  |
| `make down`               | Stop all containers                |

---

## 📁 Scripts Reference

Utility scripts in the `scripts/` folder:

| Script                      | Description                                    |
| --------------------------- | ---------------------------------------------- |
| `docker-build.sh`           | Build Docker images (`dev`, `prod`, `clean`)   |
| `deploy-bluegreen.sh`       | Zero-downtime blue-green deployment            |
| `deploy-vercel.sh`          | Deploy frontend to Vercel                      |
| `deploy-railway.sh`         | Deploy backend to Railway                      |
| `configure-features.sh`     | Interactive feature configuration wizard       |
| `validate-env-prod.sh`      | Validate production environment variables      |
| `validate-swagger.sh`       | Validate Swagger/OpenAPI documentation         |

**Usage:**

```bash
# Build development images
./scripts/docker-build.sh dev

# Deploy with blue-green strategy
./scripts/deploy-bluegreen.sh

# Check deployment status
./scripts/deploy-bluegreen.sh --status

# Rollback deployment
./scripts/deploy-bluegreen.sh --rollback
```

---

## 📐 Architecture Decisions

Key architectural decisions are documented in `docs/decisions/` using the MADR format:

| Decision | Summary |
| -------- | ------- |
| [ADR-001](docs/decisions/001-database-migrations.md) | Database migrations with golang-migrate |
| [ADR-002](docs/decisions/002-background-jobs-river.md) | Background jobs with River (PostgreSQL-backed) |
| [ADR-003](docs/decisions/003-stripe-payments.md) | Stripe payments integration |

📖 **[All Architecture Decisions →](docs/decisions/)**

---

## 🔐 Core Features

### JWT Authentication & 2FA

Complete authentication system with registration, login, email verification, password reset, and TOTP two-factor authentication. Includes OAuth support for Google and GitHub.

**Key Endpoints:**

- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login (returns JWT)
- `GET /api/auth/me` - Get current user (authenticated)
- `POST /api/auth/2fa/enable` - Enable two-factor authentication

### Multi-Tenant Organizations

Full multi-tenancy support with isolated workspaces:

- **Organization Roles** - Owner, Admin, Member with hierarchical permissions
- **Team Management** - Invite members via email, manage roles
- **Organization Plans** - Free, Pro, Enterprise tiers
- **Data Isolation** - Complete tenant separation

**Key Endpoints:**

- `POST /api/organizations` - Create organization
- `GET /api/organizations/:slug` - Get organization details
- `POST /api/organizations/:id/invite` - Invite team member

### Role-Based Access Control (RBAC)

Four-tier permission system with granular access control:

- `user` - Basic profile management
- `premium` - Access to premium content (via Stripe subscription)
- `admin` - User and content management
- `super_admin` - Full system administration

### Dashboard & Admin

Pre-built admin and dashboard interfaces:

- **Dashboard Widgets** - Customizable widget system with activity feeds
- **Admin Panel** - User management, audit logs, system settings
- **Layout Components** - Reusable AdminLayout, SettingsLayout, DashboardLayout
- **Activity Tracking** - User action logging and display

### Real-Time WebSocket

Bi-directional real-time communication:

- **Notifications** - Push alerts to connected users
- **Data Sync** - Automatic cache invalidation on updates
- **Broadcasts** - System-wide announcements
- **Auto-Reconnect** - Exponential backoff reconnection

### Stripe Payments

Full subscription billing with Stripe integration:

- **Checkout Sessions** - Secure, hosted payment pages
- **Customer Portal** - Self-service subscription management
- **Webhooks** - Real-time subscription sync with role updates
- **Billing Pages** - `/pricing` and `/billing` routes included

### Internationalization (i18n)

Multi-language support with i18next:

- **Languages** - English and Spanish included
- **Namespaces** - common, auth, errors, validation
- **Detection** - Browser language auto-detection
- **Persistence** - localStorage preference saving

### Observability

Full monitoring stack with Prometheus and Grafana:

- **HTTP Metrics** - Request count, duration, in-flight
- **Database Metrics** - Query duration, connection pool
- **WebSocket Metrics** - Active connections, messages
- **Business Metrics** - Auth attempts, registrations, subscriptions

**Running the monitoring stack:**

```bash
make observability-up    # Start Prometheus + Grafana
make observability-down  # Stop monitoring stack
```

**Access URLs:**

- Prometheus: [http://localhost:9090](http://localhost:9090)
- Grafana: [http://localhost:3001](http://localhost:3001) (admin/admin)

### Background Jobs (River)

PostgreSQL-backed job queue for reliable async processing:

- **Data Export** - GDPR-compliant user data export processing
- **Email Sending** - Async email delivery with retries
- **Cleanup Jobs** - Expired sessions, stale data cleanup
- **Retention Jobs** - Data retention policy enforcement
- Configurable workers, retries, and scheduling

### API Key Management

Secure API key system for programmatic access:

- **Encrypted Storage** - Keys encrypted at rest with AES-256
- **Provider Association** - Link keys to external services (e.g., AI providers)
- **Key Rotation** - Generate new keys without service interruption
- **Usage Tracking** - Monitor API key usage and last used timestamps

**Key Endpoints:**

- `POST /api/api-keys` - Create new API key
- `GET /api/api-keys` - List user's API keys
- `DELETE /api/api-keys/:id` - Revoke API key

### Feature Flags

Runtime feature toggles with granular control:

- **User Overrides** - Enable/disable features per user
- **Plan Enforcement** - Restrict features to specific subscription plans
- **Runtime Toggles** - Change features without redeployment
- **Admin UI** - Manage flags through admin panel

**Key Endpoints:**

- `GET /api/feature-flags` - Get all flags for current user
- `PUT /api/admin/feature-flags/:key` - Update flag (admin)
- `POST /api/admin/feature-flags/:key/user-overrides` - Set user override

### Audit Logging

Compliance-ready action tracking:

- **User Actions** - Track login, logout, data changes
- **Admin Actions** - Log role changes, user management
- **Data Access** - Record sensitive data access
- **Filterable** - Query by user, action type, date range

**Key Endpoints:**

- `GET /api/admin/audit-logs` - Query audit logs (admin)

### Data Export & Account Management

GDPR-compliant data handling:

- **Data Export** - Request full data export (processed via background job)
- **Account Deletion** - Request account deletion with grace period
- **Cancellation** - Cancel pending deletion requests
- **Login History** - View complete login history

**Key Endpoints:**

- `POST /api/users/me/export` - Request data export
- `POST /api/users/me/delete-request` - Request account deletion
- `DELETE /api/users/me/delete-request` - Cancel deletion request
- `GET /api/users/me/login-history` - View login history

### AI Integration (Gemini)

Full-featured Gemini AI integration with multiple capabilities:

- **Text Chat** - Multi-turn conversations with system prompts
- **Streaming** - Real-time token streaming via SSE
- **Multi-Modal** - Image analysis and understanding
- **Embeddings** - Vector embeddings for semantic search
- **Function Calling** - Tool use and structured interactions
- **JSON Mode** - Structured output with schema validation
- **Safety Settings** - Configurable content filtering levels

**Key Endpoints:**

- `POST /api/ai/chat` - Text chat completion
- `POST /api/ai/chat/stream` - Streaming chat (SSE)
- `POST /api/ai/chat/advanced` - Function calling & JSON mode
- `POST /api/ai/analyze-image` - Image analysis
- `POST /api/ai/embeddings` - Generate embeddings

**Security:**

- JWT authentication required for all AI endpoints
- Separate rate limiting tier for AI (20 req/min default)
- Input validation with configurable limits
- User-provided API keys support

### Session Management

Comprehensive session and device tracking:

- **Multi-Device Sessions** - Track active sessions across devices
- **Device Fingerprinting** - Browser and OS detection via User-Agent
- **Session Limits** - Configurable max concurrent sessions per user
- **Session Revocation** - Revoke individual or all sessions

### User Preferences

User-configurable settings with server-side persistence:

- **Theme Preferences** - Light, dark, or system theme
- **Notification Settings** - Email and push notification preferences
- **Language Selection** - Preferred UI language
- **Dashboard Layout** - Customizable widget arrangements

### File Upload System

Dual-backend storage supporting both AWS S3 and PostgreSQL with automatic fallback, secure uploads, and configurable size limits.

### Rate Limiting

Four-tier protection system with configurable limits:

| Tier | Default Limit | Use Case                    |
| ---- | ------------- | --------------------------- |
| IP   | 100 req/min   | Anonymous traffic           |
| User | 200 req/min   | Authenticated users         |
| Auth | 10 req/min    | Login/register attempts     |
| API  | 20 req/min    | AI and expensive operations |

All tiers support burst allowances and custom overrides via environment variables.

📖 **[Complete Features Documentation →](docs/FEATURES.md)** | **[Architecture Decisions →](docs/decisions/)**

---

## 🔧 Environment Configuration

### Essential Variables (Required)

```bash
# Copy and configure
cp .env.example .env
```

**Critical settings:**

- `JWT_SECRET` - Generate with: `openssl rand -hex 32`
- `DB_PASSWORD` - Strong database password
- `VITE_API_URL` - Frontend API endpoint

### Environment File Variants

| File                   | Purpose              | Use When                |
| ---------------------- | -------------------- | ----------------------- |
| `.env.example`         | Development template | Local development       |
| `.env.prod.example`    | Production template  | Deploying to production |
| `.env.staging.example` | Staging template     | Pre-production testing  |

```bash
# Development
cp .env.example .env

# Production
cp .env.prod.example .env

# Staging
cp .env.staging.example .env
```

### Optional Features

All optional features are included in `.env.example` as commented sections. Simply uncomment and configure the features you need:

- **AWS S3** - Cloud file storage
- **Dragonfly/Redis** - High-performance caching
- **SMTP** - Email notifications
- **Stripe** - Payment processing
- **Gemini AI** - Chat, streaming, vision, embeddings
- **Sentry** - Error tracking (frontend + backend)
- **Vercel Analytics** - Frontend performance analytics

### Feature Flags

Runtime feature toggles via environment variables:

```bash
VITE_FEATURE_NEW_DASHBOARD_LAYOUT=true
VITE_FEATURE_BETA_ANALYTICS=false
VITE_FEATURE_ADVANCED_FILE_UPLOAD=true
VITE_FEATURE_DARK_MODE_TOGGLE=true
```

### Security Headers

Configurable security headers for production:

- **CSP** - Content Security Policy
- **HSTS** - HTTP Strict Transport Security
- **CSRF** - Cross-Site Request Forgery protection
- **Frame Options** - Clickjacking protection

### Auto-Seeding (Development)

Database automatically seeds with test data on startup in development mode:

```bash
AUTO_SEED=true              # Enable auto-seeding
SEED_ADMIN_PASSWORD=admin123!  # Admin user password
SEED_DEFAULT_PASSWORD=password123!  # Default user password
```

📖 **[Environment Configuration Guide →](.env.example)**

---

## ❓ FAQ

<details>
<summary><strong>Can I use MySQL or SQLite instead of PostgreSQL?</strong></summary>

No. PostgreSQL is required because:
- Migrations use PostgreSQL-specific syntax
- River (background jobs) requires PostgreSQL
- GORM configuration is PostgreSQL-optimized
- Some features use PostgreSQL-specific types (JSONB, arrays)

</details>

<details>
<summary><strong>How do I add a new language?</strong></summary>

1. Create translation files in `frontend/app/i18n/locales/{lang}/`
2. Copy structure from `en/` folder
3. Add language to `frontend/app/i18n/config.ts`
4. Update the language selector component

📖 See [Frontend Guide](docs/FRONTEND_GUIDE.md#internationalization) for details.

</details>

<details>
<summary><strong>How do I disable optional features?</strong></summary>

Comment out or remove the relevant environment variables in `.env`:
- **Stripe**: Remove `STRIPE_*` variables
- **AI**: Remove `GEMINI_API_KEY`
- **S3**: Remove `AWS_*` variables (falls back to DB storage)
- **Email**: Remove `SMTP_*` variables
- **Sentry**: Remove `SENTRY_DSN`

The application gracefully handles missing optional features.

</details>

<details>
<summary><strong>How do I change the port numbers?</strong></summary>

Edit `.env`:
```bash
API_PORT=8080        # Backend port
FRONTEND_PORT=5193   # Frontend port (Vite)
DB_PORT=5432         # PostgreSQL port
```

</details>

<details>
<summary><strong>How do I run only the backend or frontend?</strong></summary>

```bash
# Backend only
cd backend && go run cmd/main.go

# Frontend only
cd frontend && npm run dev

# Or with Docker
docker compose up -d backend postgres
docker compose up -d frontend
```

</details>

<details>
<summary><strong>How do I add a new API endpoint?</strong></summary>

1. Create handler in `backend/internal/handlers/`
2. Add route in `backend/cmd/main.go`
3. Create TypeScript types in `frontend/app/types/`
4. Add API service function in `frontend/app/services/`
5. Create query/mutation hooks in `frontend/app/hooks/`

📖 See [Backend README](backend/README.md) for patterns.

</details>

---

## 🔄 CI/CD Pipeline

The GitHub Actions workflow provides comprehensive automation:

### Change Detection

Smart detection runs only affected checks:
- Frontend changes → lint, typecheck, test, E2E
- Backend changes → vet, test, migration validation
- Both → full pipeline

### Pipeline Stages

| Stage | Checks | Runs On |
| ----- | ------ | ------- |
| Lint | ESLint, Prettier, go vet, golangci-lint | All PRs |
| Types | TypeScript strict mode | Frontend changes |
| Test | Vitest (unit), Go tests | All changes |
| E2E | Playwright on Chromium | Frontend changes |
| Security | npm audit, govulncheck | All PRs |
| Migrations | up → down → up validation | Backend changes |
| Swagger | OpenAPI spec validation | Backend changes |

### Compatibility Matrix

On `master` branch merges:
- Node.js 22 verification
- Go 1.25 verification
- Full E2E test suite

### Running Locally

```bash
# Run the same checks as CI
npm run quality          # Frontend lint + format + typecheck
npm run test:fast        # Frontend unit tests
cd backend && go vet ./... && go test ./...  # Backend checks
```

---

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes
4. Add tests if applicable
5. Run tests: `npm test` (frontend) or `go test ./...` (backend)
6. Commit with conventional commits: `feat: add new feature`
7. Push and create a Pull Request

📖 **[Development Guide →](docs/DEVELOPMENT.md)**

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🆘 Need Help?

### Quick Troubleshooting

| Issue | Solution |
| ----- | -------- |
| Port already in use | Change ports in `.env` or stop conflicting services |
| Database connection failed | Check PostgreSQL is running: `make health` |
| Docker permission denied | Add user to docker group: `sudo usermod -aG docker $USER` |
| npm install fails | Clear cache: `rm -rf node_modules && npm install` |
| Migrations fail | Check DB connection, run `make migrate-down` then `make migrate-up` |
| JWT errors | Regenerate: `openssl rand -hex 32` and update `.env` |

### Resources

- **Documentation Hub** → [docs/README.md](docs/README.md)
- **Deployment Issues** → [docs/DEPLOYMENT.md#troubleshooting](docs/DEPLOYMENT.md#troubleshooting-common-issues)
- **Feature Questions** → [docs/FEATURES.md](docs/FEATURES.md)
- **Frontend Guide** → [docs/FRONTEND_GUIDE.md](docs/FRONTEND_GUIDE.md)
- **Backend Guide** → [backend/README.md](backend/README.md)
- **Found a Bug?** → [Open an issue](https://github.com/DailyDisco/react-golang-starter-kit/issues)

---

Built with ❤️ for rapid full-stack SaaS development.
