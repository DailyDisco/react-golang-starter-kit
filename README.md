# React-Golang Starter Kit

A modern, production-ready full-stack starter template combining **React 19** (with TanStack Router & Query) and **Go** (with Chi, GORM, JWT). Built for rapid development with Docker, featuring authentication, RBAC, file uploads, and comprehensive testing.

🌐 **[Live Demo](https://react-golang-starter-kit.vercel.app/)** | 📚 **[Full Documentation](docs/README.md)**

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
- Frontend: [http://localhost:5173](http://localhost:5173)
- Backend API: [http://localhost:8080](http://localhost:8080)
- API Health: [http://localhost:8080/health](http://localhost:8080/health)

📖 **[Complete Docker Guide →](docs/DOCKER_SETUP.md)**

### Local Development

**Prerequisites:** Go 1.24+, Node.js (LTS), PostgreSQL

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
- 🧪 **Vitest** - Fast, comprehensive testing

### Backend Stack
- 🐹 **Go 1.24+** with Chi router
- 🗄️ **GORM + PostgreSQL** - Powerful ORM and database
- 🔐 **JWT Authentication** - Secure token-based auth
- 👥 **Role-Based Access Control (RBAC)** - 4 permission levels
- 📤 **File Upload System** - AWS S3 or database storage
- 🛡️ **Rate Limiting** - Configurable API protection

### DevOps & Production
- 🐳 **Docker** - Development and production ready
- 📦 **Multi-stage builds** - Optimized images
- 🔧 **Environment-based config** - Comprehensive .env support
- ✅ **CI/CD Ready** - GitHub Actions workflows included

📖 **[Detailed Features Guide →](docs/FEATURES.md)**

---

## 📂 Project Structure

```
react-golang-starter-kit/
├── backend/              # Go API server
│   ├── cmd/             # Application entry point
│   ├── internal/        # Private application code
│   │   ├── auth/        # JWT authentication
│   │   ├── handlers/    # HTTP request handlers
│   │   ├── middleware/  # Chi middleware
│   │   ├── models/      # GORM models
│   │   └── storage/     # File storage (S3/DB)
│   └── docs/            # Swagger documentation
│
├── frontend/            # React application
│   ├── app/            # Application code
│   │   ├── routes/     # TanStack Router pages
│   │   ├── components/ # Reusable UI components
│   │   ├── hooks/      # Custom React hooks
│   │   └── lib/        # Utilities and helpers
│   └── public/         # Static assets
│
├── docs/               # Documentation
│   ├── README.md       # Documentation hub
│   ├── FEATURES.md     # Feature documentation
│   ├── DEPLOYMENT.md   # Deployment guides
│   ├── DOCKER_SETUP.md # Docker configuration
│   └── FRONTEND_GUIDE.md # React development
│
└── docker-compose.yml  # Development environment
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

### Code Quality
```bash
npm test              # Run frontend tests
npm run lint          # Check code formatting
npm run format        # Fix code formatting
make format-backend   # Format Go backend code
```

---

## 🔐 Core Features

### JWT Authentication
Complete authentication system with registration, login, email verification, and password reset. Includes secure password hashing and token management.

**Key Endpoints:**
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login (returns JWT)
- `GET /api/auth/me` - Get current user (authenticated)

### Role-Based Access Control (RBAC)
Four-tier permission system with granular access control:
- `user` - Basic profile management
- `premium` - Access to premium content
- `admin` - User and content management
- `super_admin` - Full system administration

### File Upload System
Dual-backend storage supporting both AWS S3 and PostgreSQL with automatic fallback, secure uploads, and configurable size limits.

📖 **[Complete Features Documentation →](docs/FEATURES.md)**

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

**Built with ❤️ for rapid full-stack development**
