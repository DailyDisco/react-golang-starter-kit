# ✨ React-Golang Starter Kit ✨

This project serves as a robust and modern starter kit for building full-stack applications, seamlessly integrating a React frontend with a high-performance Golang backend. Designed for rapid development and scalability, it provides a solid foundation with best practices already in place.

🌐 **Live Demo:** [https://react-golang-starter-kit.vercel.app/](https://react-golang-starter-kit.vercel.app/)

## 📋 Table of Contents

- [🚀 Quick Start](#-quick-start)
- [🚀 Features](#-features)
- [🏁 Getting Started](#-getting-started)
- [🔐 Authentication & Security](#-authentication--security)
- [🐳 Docker Setup](#-docker-setup)
- [🚀 Deployment](#-deployment)
- [📚 API Documentation](#-api-documentation)
- [🧪 Testing](#-testing)
- [📂 Project Structure](#-project-structure)
- [🔧 Configuration](#-configuration)
- [🔄 CI/CD Pipeline](#-ci/cd-pipeline)
- [🤝 Contributing](#-contributing)

## 🚀 Quick Start

> **New to the project?** Start here for the fastest setup!

### Option 1: Docker (Recommended)

```bash
git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
cd react-golang-starter-kit
docker-compose up -d
```

Your app will be running at [http://localhost:5173](http://localhost:5173)!

### Option 2: Local Development

```bash
git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
cd react-golang-starter-kit

# Backend
cd backend && go mod tidy && go run cmd/main.go

# Frontend (new terminal)
cd ../frontend && npm install && npm run dev
```

## 🚀 Features

### ⚛️ React Frontend

- **[Vite](https://vitejs.dev/)** - Blazing-fast development and optimized builds
- **[React Router](https://reactrouter.com/en/main)** - Declarative navigation and routing
- **[TailwindCSS](https://tailwindcss.com/)** - Utility-first CSS framework
- **[ShadCN UI](https://ui.shadcn.com/)** - Beautiful and accessible UI components
- **[Vitest](https://vitest.dev/)** - Fast unit and component testing
- **TypeScript** - Type-safe development experience

### ⚙️ Golang Backend

- **[Chi Router](https://go-chi.io/)** - Lightweight and fast HTTP router
- **[GORM](https://gorm.io/)** - Elegant ORM with PostgreSQL integration
- **[JWT Authentication](https://jwt.io/)** - Secure token-based authentication
- **[Rate Limiting](https://github.com/go-chi/httprate)** - API abuse protection
- **[Swagger/OpenAPI](https://swagger.io/)** - Interactive API documentation
- **[Air](https://github.com/cosmtrek/air)** - Live reloading during development

### 🛡️ Security & Performance

- **Password Hashing** - Bcrypt encryption for secure password storage
- **Rate Limiting** - Configurable request throttling by endpoint and user
- **CORS Protection** - Configurable cross-origin request handling
- **Input Validation** - Comprehensive request validation and sanitization
- **Environment-based Config** - Secure configuration management

### 🐳 DevOps & Deployment

- **Docker Support** - Containerized development and deployment
- **Multi-stage Builds** - Optimized production images
- **Git Hooks** - Automated code quality checks (Husky)
- **CI/CD Ready** - GitHub Actions workflows included
- **Environment Management** - `.env` file support with validation

## 🔐 Authentication & Security

### JWT Authentication

The backend includes a complete JWT (JSON Web Token) authentication system with the following features:

#### 🔑 Authentication Endpoints

| Method | Endpoint                           | Description                  | Auth Required |
| ------ | ---------------------------------- | ---------------------------- | ------------- |
| `POST` | `/api/auth/register`               | Register new user account    | ❌            |
| `POST` | `/api/auth/login`                  | User login with credentials  | ❌            |
| `GET`  | `/api/auth/me`                     | Get current user information | ✅            |
| `GET`  | `/api/auth/verify-email`           | Verify user email address    | ❌            |
| `POST` | `/api/auth/reset-password`         | Request password reset       | ❌            |
| `POST` | `/api/auth/reset-password/confirm` | Confirm password reset       | ❌            |

#### 🛡️ Security Features

- **Password Security**: Bcrypt hashing with configurable cost factor
- **JWT Tokens**: 24-hour expiration (configurable via `JWT_EXPIRATION_HOURS`)
- **Password Validation**: Minimum 8 characters, requires uppercase, lowercase, and digits
- **Email Verification**: Token-based email verification system
- **Password Reset**: Secure password reset flow with expiration tokens
- **Bearer Authentication**: Standard `Authorization: Bearer <token>` header

#### 📝 Example Usage

**Register User:**

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "SecurePass123"
  }'
```

**Login:**

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123"
  }'
```

**Access Protected Route:**

```bash
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Rate Limiting

The API includes comprehensive rate limiting to prevent abuse and ensure fair usage across different user types and endpoints.

#### ⚙️ Rate Limiting Configuration

Rate limits are configurable via environment variables and apply different rules based on endpoint types:

| Environment Variable         | Default | Description                                |
| ---------------------------- | ------- | ------------------------------------------ |
| `RATE_LIMIT_ENABLED`         | `true`  | Enable/disable rate limiting globally      |
| `RATE_LIMIT_IP_PER_MINUTE`   | `60`    | Requests per minute per IP                 |
| `RATE_LIMIT_IP_PER_HOUR`     | `1000`  | Requests per hour per IP                   |
| `RATE_LIMIT_USER_PER_MINUTE` | `120`   | Requests per minute per authenticated user |
| `RATE_LIMIT_USER_PER_HOUR`   | `2000`  | Requests per hour per authenticated user   |
| `RATE_LIMIT_AUTH_PER_MINUTE` | `5`     | Strict limit for auth endpoints            |
| `RATE_LIMIT_API_PER_MINUTE`  | `100`   | General API endpoint limits                |

#### 🏷️ Rate Limit Headers

When rate limited, the API returns these headers:

```
X-RateLimit-Limit: 60          # Maximum requests allowed
X-RateLimit-Remaining: 0       # Remaining requests in current window
X-RateLimit-Reset: 1693526400  # Unix timestamp when limit resets
Retry-After: 60                # Seconds to wait before retrying
```

#### 📊 Rate Limiting by Endpoint Type

- **Authentication Endpoints** (`/api/auth/*`): Strict IP-based limiting (5/minute)
- **Protected Endpoints**: User-based limiting (120/minute for authenticated users)
- **Public API Endpoints**: Mixed IP/user-based limiting (100/minute)
- **Global**: IP-based limiting applied to all requests (60/minute)

#### 🚫 Rate Limit Response

```http
HTTP/1.1 429 Too Many Requests
Content-Type: text/plain; charset=utf-8
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1693526400
Retry-After: 60

Rate limit exceeded. Too many requests from this IP address.
```

## 🏁 Getting Started

### Prerequisites

- **Git** - Version control
- **Node.js (LTS)** & **npm** - Frontend development
- **Go (1.24+)** - Backend development
- **Docker & Docker Compose** _(Recommended)_ - Isolated development environments
- **PostgreSQL** - Database server (if not using Docker)

### Setup Options

#### Option 1: Docker (Recommended)

```bash
git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
cd react-golang-starter-kit
docker-compose up -d
```

**Services:**

- **Frontend:** [http://localhost:5173](http://localhost:5173)
- **Backend API:** [http://localhost:8080](http://localhost:8080)
- **API Docs:** [http://localhost:8080/swagger/](http://localhost:8080/swagger/)

**Useful Docker Commands:**

```bash
docker-compose logs -f          # View logs
docker-compose down             # Stop services
docker-compose up --build -d    # Rebuild after changes
```

#### Option 2: Local Development

1. **Clone and setup:**

   ```bash
   git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
   cd react-golang-starter-kit
   cp .env.example .env
   ```

2. **Configure environment:** Edit `.env` with your database credentials

3. **Start services:**

   ```bash
   # Backend (with live reloading)
   cd backend && go mod tidy && air

   # Frontend (new terminal)
   cd ../frontend && npm install && npm run dev
   ```

## 🚀 Deployment

### Quick Deployment Guide

Choose your preferred deployment method:

#### 🚀 Vercel + Railway (Recommended for Beginners)

**Best for:** Quick setup, modern workflow, generous free tiers

1. **Database**: Create PostgreSQL on [Railway.app](https://railway.app) (free tier)
2. **Backend**: Deploy `backend/` folder to Railway
3. **Frontend**: Deploy `frontend/` folder to [Vercel](https://vercel.com)
4. **Connect**: Set `VITE_API_URL` in Vercel to your Railway backend URL

**Cost:** ~$0-10/month | **Time:** 15-30 minutes

#### 🐳 Docker + VPS

**Best for:** Full control, cost-effective for production

```bash
# Build production images
docker build -t myapp-backend:latest ./backend
docker build -t myapp-frontend:latest ./frontend

# Deploy with docker-compose
docker-compose up -d
```

**Cost:** VPS hosting only (~$5-20/month) | **Time:** 30-60 minutes

### Detailed Deployment Guides

#### Option 1: Vercel + Railway (Step-by-Step)

**🗄️ Database Setup:**

1. Sign up at [Railway.app](https://railway.app)
2. Create new project → Add PostgreSQL
3. Note the database credentials (auto-provided)

**⚙️ Backend Deployment:**

1. Create new Railway project
2. Connect your GitHub repository
3. Set build settings:
   - **Root Directory:** `backend`
   - **Environment Variables:**
     ```
     CORS_ALLOWED_ORIGINS=https://your-vercel-app.vercel.app
     JWT_SECRET=your-secret-key-here
     ```
4. Railway auto-detects Go and deploys
5. Note your backend URL: `https://your-app.up.railway.app`

**🌐 Frontend Deployment:**

1. Connect GitHub repo to [Vercel](https://vercel.com)
2. Configure build settings:
   - **Root Directory:** `frontend`
   - **Build Command:** `npm run build`
   - **Output Directory:** `dist`
3. Set environment variables:
   ```
   VITE_API_URL=https://your-railway-backend.up.railway.app
   ```
4. Deploy!

#### Option 2: Docker VPS Deployment

**Production-Ready Docker Setup:**

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: prod_db
      POSTGRES_USER: prod_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

  backend:
    image: myapp-backend:latest
    environment:
      - DB_HOST=postgres
      - DB_PASSWORD=${DB_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
    ports:
   - '8080:8080'
    depends_on:
      - postgres

  frontend:
    image: myapp-frontend:latest
    ports:
   - '80:80'
 environment:
   - VITE_API_URL=http://localhost:8080

volumes:
  postgres_data:
```

**Deploy Commands:**

```bash
# On your VPS
git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
cd react-golang-starter-kit

# Build images
docker-compose build

# Start services
docker-compose up -d

# Setup SSL (optional)
# docker run -it --rm --name certbot certbot certonly --webroot --webroot-path /var/www/html -d yourdomain.com
```

### Troubleshooting Common Issues

#### ❌ CORS Errors

**Solution:** Set `CORS_ALLOWED_ORIGINS=https://yourdomain.com` in backend environment

#### ❌ Database Connection Failed

**Solution:** Ensure Railway PostgreSQL is linked to your backend service

#### ❌ API Returns 404

**Solution:** Use base URL only in `VITE_API_URL` (no `/api/` suffix)

#### ❌ Vercel Build Fails

**Solution:** Ensure Root Directory is set to `frontend` in Vercel settings

### Alternative Deployment Platforms

| Platform                  | Backend         | Frontend      | Database     | Cost/Month | Setup Time |
| ------------------------- | --------------- | ------------- | ------------ | ---------- | ---------- |
| **Railway + Vercel**      | ✅ Native Go    | ✅ Optimized  | ✅ Built-in  | $0-10      | 15-30 min  |
| **Docker + DigitalOcean** | ✅ Full control | ✅ Custom     | ✅ Managed   | $5-20      | 30-60 min  |
| **AWS (ECS/Fargate)**     | ✅ Scalable     | ✅ CloudFront | ✅ RDS       | $20-100+   | 60-120 min |
| **Google Cloud Run**      | ✅ Serverless   | ✅ Cloud CDN  | ✅ Cloud SQL | $10-50     | 45-90 min  |
| **Fly.io**                | ✅ Go optimized | ✅ Global CDN | ✅ Built-in  | $5-30      | 20-40 min  |

### Deployment Checklist

- [ ] Database created and accessible
- [ ] Backend deployed and health check passes (`/api/health`)
- [ ] Frontend deployed and loads without errors
- [ ] Environment variables configured correctly
- [ ] CORS settings allow frontend origin
- [ ] API endpoints respond correctly
- [ ] Authentication flow works (register/login)
- [ ] Rate limiting configured appropriately

## 🧪 Testing

### Frontend (React with Vitest)

The frontend uses **Vitest** with **Happy DOM** for fast, reliable testing. Happy DOM is a lightweight alternative to jsdom that provides better performance.

#### Quick Test Commands

```bash
cd frontend

# Run tests once (CI mode)
npm run test:fast

# Run tests in watch mode (development)
npm test
# or
npm run test:dev

# Run tests with coverage report
npm run test:coverage

# Run tests with web UI (opens browser)
npm run test:ui
```

#### Test Environment Features

- ✅ **Happy DOM** - Fast, lightweight DOM implementation
- ✅ **Global test functions** - No need to import describe/it/expect
- ✅ **Hot reload** - Tests rerun automatically on file changes
- ✅ **Coverage reporting** - Built-in coverage with HTML reports
- ✅ **Web UI** - Visual test runner with detailed results

## 📜 Available Scripts

### Frontend Scripts

```bash
cd frontend
npm run dev          # Start development server
npm run build        # Build for production
npm run preview      # Preview production build
npm run typecheck    # Run TypeScript type checking

# Testing Scripts
npm test             # Run tests in watch mode
npm run test:fast    # Run tests once with basic output
npm run test:dev     # Run tests in watch mode (alias for npm test)
npm run test:coverage # Run tests with coverage report
npm run test:ui      # Run tests with web UI (opens browser)

npm run prettier:check # Check code formatting
npm run prettier:fix   # Fix code formatting
```

### Backend Scripts

```bash
cd backend
go run cmd/main.go   # Start server (without Air)
air                  # Start with live reloading
go mod tidy          # Install/update dependencies
go test ./...        # Run all tests
```

## 🔧 Troubleshooting

**Database connection failed:**

```bash
cd backend
# Make sure PostgreSQL is running
# Check your .env file has correct DB credentials
```

**Port already in use:**

```bash
# Kill process using port 8080 (backend) or 3000 (frontend)
kill -9 $(lsof -ti:8080)
```

**Air not found after installation:**

```bash
export PATH=$PATH:$(go env GOPATH)/bin
# Or restart your terminal
```

## 📂 Project Structure

### High-Level Overview

```
react_golang_starter_kit/
├── backend/                  # 🏗️ Go Backend API
├── frontend/                 # ⚛️ React Frontend App
├── docker-compose.yml        # 🐳 Docker orchestration
└── README.md                # 📖 This documentation
```

### Backend Structure (`/backend`)

| Directory/File            | Purpose                      |
| ------------------------- | ---------------------------- |
| **`cmd/main.go`**         | Application entry point      |
| **`internal/auth/`**      | JWT authentication system    |
| **`internal/ratelimit/`** | Rate limiting middleware     |
| **`internal/database/`**  | Database connection & config |
| **`internal/handlers/`**  | HTTP request handlers        |
| **`internal/models/`**    | Data models & GORM structs   |
| **`docs/`**               | Swagger API documentation    |
| **`Dockerfile`**          | Backend containerization     |
| **`go.mod`**              | Go module dependencies       |

### Frontend Structure (`/frontend`)

| Directory/File        | Purpose                      |
| --------------------- | ---------------------------- |
| **`app/`**            | Main application source      |
| ├── **`components/`** | Reusable React components    |
| ├── **`routes/`**     | React Router page components |
| ├── **`lib/`**        | Utilities & API client       |
| ├── **`hooks/`**      | Custom React hooks           |
| ├── **`types/`**      | TypeScript type definitions  |
| **`public/`**         | Static assets                |
| **`package.json`**    | Node.js dependencies         |
| **`vite.config.ts`**  | Vite build configuration     |

### Key Files You'll Work With

#### 🔧 Development

- **`backend/cmd/main.go`** - Start here for backend changes
- **`frontend/app/routes/`** - Add new pages here
- **`frontend/app/components/`** - Create reusable components here
- **`frontend/app/lib/api.ts`** - API integration layer

#### ⚙️ Configuration

- **`docker-compose.yml`** - Local development environment
- **`.env`** - Environment variables (create from `.env.example`)
- **`backend/internal/auth/`** - Authentication settings
- **`frontend/tailwind.config.ts`** - Styling configuration

#### 🚀 Deployment

- **`backend/Dockerfile`** - Backend container build
- **`frontend/Dockerfile`** - Frontend container build
- **`vercel.json`** - Vercel deployment config (if needed)

### Quick Navigation Tips

- **Adding a new page?** → `frontend/app/routes/`
- **Creating API endpoint?** → `backend/internal/handlers/`
- **Database model changes?** → `backend/internal/models/`
- **Authentication logic?** → `backend/internal/auth/`
- **Styling updates?** → `frontend/app/components/` or `tailwind.config.ts`

## 🔧 Configuration

### Environment Variables

All configuration is managed through `.env` files. Copy `.env.example` to `.env` and customize the values.

#### Required Variables

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_database_name

# Security
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:5173,https://yourdomain.com
```

#### Optional Variables

```bash
# JWT Configuration
JWT_EXPIRATION_HOURS=24

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_IP_PER_MINUTE=60
RATE_LIMIT_USER_PER_MINUTE=120

# API
API_PORT=8080

# Frontend
VITE_API_URL=http://localhost:8080
```

### Configuration Files

- **`.env`** - Environment variables (create from `.env.example`)
- **`docker-compose.yml`** - Local development environment
- **`backend/go.mod`** - Go dependencies
- **`frontend/package.json`** - Node.js dependencies
- **`frontend/tailwind.config.ts`** - Styling configuration

## 🔄 CI/CD Pipeline

This project includes comprehensive CI/CD workflows following industry best practices for React and Go development.

### Available Workflows

- **Complete CI** (`ci.yml`): Full pipeline with security scanning, linting, testing, and builds for both frontend and backend
- **React CI** (`react-ci.yml`): Frontend-focused workflow with Node.js matrix testing and coverage reporting
- **Go CI** (`go-ci.yml`): Backend-focused workflow with cross-platform builds and race detection

### Key Features

- **Security**: Automated vulnerability scanning for both npm and Go dependencies
- **Quality**: Linting, formatting, and type checking with Prettier, TypeScript, and golangci-lint
- **Testing**: Comprehensive test suites with coverage reporting via Codecov
- **Performance**: Caching, parallel execution, and artifact management
- **Cross-Platform**: Multi-Node.js testing and multi-platform Go builds

### Triggers

Workflows run on push to master, pull requests, and manual dispatch.

### Customization

Modify workflow files to adjust test commands, coverage thresholds, or security rules.

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/your-feature-name`).
3. Make your changes.
4. Commit your changes (`git commit -m 'feat: Add new feature'`).
5. Push to the branch (`git push origin feature/your-feature-name`).
6. Open a Pull Request.

Please ensure your code adheres to the existing style and conventions.

## 📄 License

This project is licensed under the MIT License - see the `LICENSE` file for details.
