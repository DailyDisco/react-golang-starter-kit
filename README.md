# React-Golang Starter Kit

A modern, production-ready full-stack starter template combining **React 19** (with TanStack Router & Query) and **Go** (with Chi, GORM, JWT). Built for rapid SaaS development with Docker, featuring multi-tenant organizations, real-time WebSockets, i18n, Prometheus observability, Stripe billing, and comprehensive testing.

📚 **[Full Documentation](docs/README.md)**

---

## 🚀 Quick Start

Get up and running in under 5 minutes:

### Docker (Recommended)

```bash
# Clone and configure
git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
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

**Prerequisites:** Go 1.25+, Node.js (LTS), PostgreSQL

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

- ⚛️ **React 19** with TypeScript
- ⚡ **Vite** - Lightning-fast builds and HMR
- 🛣️ **TanStack Router** - Type-safe, file-based routing
- 🔄 **TanStack Query** - Powerful async state management
- 🎨 **TailwindCSS + ShadCN UI** - Beautiful, accessible components
- 🌐 **i18next** - Multi-language support (EN/ES included)
- 📡 **WebSocket** - Real-time notifications and data sync
- 🧪 **Vitest** - Fast, comprehensive testing

### Backend Stack

- 🐹 **Go 1.25+** with Chi router
- 🗄️ **GORM + PostgreSQL** - Powerful ORM and database
- 🔐 **JWT Authentication** - Secure token-based auth with 2FA
- 🏢 **Multi-Tenant Organizations** - Teams with roles and invitations
- 👥 **Role-Based Access Control (RBAC)** - 4 permission levels
- 📤 **File Upload System** - AWS S3 or database storage
- 🛡️ **Rate Limiting** - Configurable API protection
- 📧 **Email Service** - SMTP with HTML templates
- 💳 **Stripe Payments** - Subscriptions, Checkout, Customer Portal
- ⚡ **Background Jobs** - River (PostgreSQL-backed job queue)
- 📡 **WebSocket Server** - Real-time push notifications
- 🔄 **Database Migrations** - golang-migrate with CI validation
- 🤖 **AI Integration** - Gemini with chat, streaming, vision, embeddings

### Observability & DevOps

- 📊 **Prometheus Metrics** - HTTP, DB, cache, WebSocket, auth metrics
- 📈 **Grafana Dashboards** - Pre-configured monitoring
- 🐳 **Docker Compose** - Development and production ready
- 📦 **Multi-stage builds** - Optimized images
- 🔧 **Environment-based config** - Comprehensive .env support
- ✅ **CI/CD Ready** - GitHub Actions workflows included

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
│   │   ├── cache/       # Redis/memory caching
│   │   ├── config/      # Configuration management
│   │   ├── database/    # Database connection & migrations
│   │   ├── email/       # SMTP email service
│   │   ├── handlers/    # HTTP request handlers
│   │   ├── jobs/        # River background jobs
│   │   ├── middleware/  # Chi middleware
│   │   ├── models/      # GORM models (users, orgs)
│   │   ├── observability/ # Prometheus metrics
│   │   ├── ratelimit/   # Rate limiting logic
│   │   ├── services/    # Business logic layer
│   │   ├── storage/     # File storage (S3/DB)
│   │   ├── stripe/      # Stripe payments
│   │   └── websocket/   # Real-time WebSocket hub
│   ├── docs/            # Swagger documentation
│   └── scripts/         # Utility scripts
│
├── frontend/            # React application
│   ├── app/            # Application code
│   │   ├── components/ # Reusable UI components
│   │   ├── hooks/      # Custom React hooks
│   │   ├── i18n/       # Internationalization (EN/ES)
│   │   ├── layouts/    # Layout components
│   │   ├── lib/        # Utilities and helpers
│   │   ├── routes/     # TanStack Router pages
│   │   ├── services/   # API service layer
│   │   └── stores/     # Zustand state management
│   └── public/         # Static assets
│
├── grafana/            # Grafana dashboard configs
├── prometheus/         # Prometheus configuration
│
├── docker-compose.yml  # Development environment
├── docker-compose.prod.yml # Production environment
├── docker-compose.observability.yml # Monitoring stack
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
docker compose -f docker-compose.prod.yml up -d
```

**Time:** 30-60 minutes | **Cost:** $5-20/month

📖 **[Complete Deployment Guide →](docs/DEPLOYMENT.md)**

---

## 🧪 Testing

### Frontend Tests

```bash
cd frontend
npm test              # Watch mode
npm run test:fast     # Run once (CI)
npm run test:coverage # Coverage report
```

### Backend Tests

```bash
cd backend
go test ./...         # Run all tests
go test -cover ./...  # With coverage
```

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

### Code Quality

```bash
npm test              # Run frontend tests
npm run lint          # Check code formatting
npm run format        # Fix code formatting
make format-backend   # Format Go backend code
```

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

### Background Jobs (River)

PostgreSQL-backed job queue for reliable async processing:

- Email sending
- Webhook processing
- Configurable workers and retries

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

### File Upload System

Dual-backend storage supporting both AWS S3 and PostgreSQL with automatic fallback, secure uploads, and configurable size limits.

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

### Optional Features

All optional features (AWS S3, Redis, SMTP, payments, AI, analytics) are included in `.env.example` as commented sections. Simply uncomment and configure the features you need.

📖 **[Environment Configuration Guide →](.env.example)**

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

- **Documentation Issues?** Check the [Documentation Hub](docs/README.md)
- **Troubleshooting?** See [Deployment Guide](docs/DEPLOYMENT.md#troubleshooting-common-issues)
- **Feature Questions?** Review [Features Documentation](docs/FEATURES.md)
- **Found a Bug?** Open an issue on GitHub

---

Built with care for rapid full-stack SaaS development.
