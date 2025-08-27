# ✨ React-Golang Starter Kit ✨

This project serves as a robust and modern starter kit for building full-stack applications, seamlessly integrating a React frontend with a high-performance Golang backend. Designed for rapid development and scalability, it provides a solid foundation with best practices already in place.

🌐 **Live Demo:** [https://react-golang-starter-kit.vercel.app/](https://react-golang-starter-kit.vercel.app/)

## 🚀 Features

- **⚛️ React Frontend:**
  - Built with [Vite](https://vitejs.dev/) for blazing-fast development.
  - [React Router](https://reactrouter.com/en/main) for declarative navigation.
  - [TailwindCSS](https://tailwindcss.com/) for utility-first styling.
  - [ShadCN UI](https://ui.shadcn.com/) components for a beautiful and accessible user interface.
  - **[Vitest](https://vitest.dev/)** for fast unit and component testing.
  - Optimized for performance and developer experience.
- **⚙️ Golang Backend:**
  - Powered by the [Fiber Web Framework](https://gofiber.io/) for a fast and flexible API.
  - [Air](https://github.com/cosmtrek/air) for live reloading during development.
  - [GORM](https://gorm.io/) for elegant Object-Relational Mapping (ORM) with PostgreSQL.
  - **Swagger/OpenAPI documentation** with interactive UI for API exploration and testing.
  - Structured project layout for maintainability and scalability.
  - Includes basic CRUD operations and authentication scaffolding.
- **🐳 Docker Support:**
  - `Dockerfiles` for both frontend and backend for easy containerization.
  - `docker-compose.yml` for orchestrating all services (PostgreSQL, backend, frontend).
  - Simplified deployment and consistent development environments.
- **💾 Database Integration:**
  - Pre-configured for PostgreSQL, allowing quick setup and integration.
  - Scalable and reliable data storage solution.
- **✅ API Testing:**
  - Integrated tools for efficient API testing to ensure robustness.
- **🔧 Husky Git Hooks:**
  - Automated code quality checks on commit and push.
  - Conventional commit message validation (feat, fix, docs, refactor, etc.).
  - Pre-commit hooks run targeted tests, linting, and type checking based on changed files.
  - Pre-push hooks run full test suites to prevent broken code from reaching repository.
  - Intelligent caching (5-minute validity) for better performance.
  - Hooks auto-install during `npm install` for immediate protection.

## 🏁 Getting Started

Follow these steps to get your development environment up and running.

### Prerequisites

Ensure you have the following installed on your system:

- **Git:** For version control.
- **Node.js (LTS) & npm/yarn:** For frontend development.
- **Go (1.24+):** For backend development.
- **Docker & Docker Compose (Optional):** Highly recommended for isolated development environments and deployment.
- **PostgreSQL:** Database server.

### Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
    cd react-golang-starter-kit
    ```

2.  **Set up Environment Variables:**
    Create a `.env` file in the project root based on `.env.example`.

    ```bash
    # Backend Environment Variables
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=user
    DB_PASSWORD=password
    DB_NAME=mydatabase
    JWT_SECRET=supersecretkey
    API_PORT=8080

    # Frontend Environment Variables
    VITE_API_URL=http://localhost:8080
    ```

3.  **Backend Setup:**

    ```bash
    cd backend
    go mod tidy          # Download dependencies
    go install github.com/cosmtrek/air@latest  # Install Air for live reloading
    air                  # Start the backend server with live reloading
    ```

    Alternatively, you can run without Air:

    ```bash
    go run cmd/main.go   # Start the backend server
    ```

4.  **Frontend Setup:**

    ```bash
    cd ../frontend
    npm install          # Install frontend dependencies
    npm run dev          # Start the frontend development server
    ```

Your application should now be running!

## 🐳 Docker Setup (Recommended)

For the easiest setup with isolated environments, use Docker Compose. This will run PostgreSQL, the Go backend, and React frontend in separate containers.

### Quick Start with Docker

1. **Prerequisites:**

   - Docker & Docker Compose installed

2. **Start all services:**

   ```bash
   docker-compose up -d
   ```

3. **View logs:**

   ```bash
   docker-compose logs -f
   ```

4. **Stop services:**
   ```bash
   docker-compose down
   ```

### Docker Services

- **PostgreSQL**: Database server on port 5432
- **Backend (Go)**: API server on port 8080
- **Frontend (React)**: Web app on port 5173

### Docker Development Workflow

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Stop services
docker-compose down

# Rebuild after code changes
docker-compose up --build -d
```

### Docker Commands

```bash
# Start specific service
docker-compose up postgres
docker-compose up backend

# Remove volumes (WARNING: deletes database data)
docker-compose down -v

# View running containers
docker-compose ps
```

## 🚀 Deployment Options

### Option 1: Docker + VPS (Simple & Cost-Effective)

**Best for:** Full control, single server deployment

1. **Build production images:**

   ```bash
   # Build optimized images
   docker build -t myapp-backend:latest ./backend
   docker build -t myapp-frontend:latest ./frontend
   ```

2. **Create production docker-compose.yml:**

   ```yaml
   version: "3.8"
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
         - "8080:8080"
       depends_on:
         - postgres

     frontend:
       image: myapp-frontend:latest
       ports:
         - "80:80"

   volumes:
     postgres_data:
   ```

3. **Deploy on your VPS:**

   ```bash
   docker-compose up -d
   ```

### Option 2: Vercel + Railway (Developer-Friendly)

**Best for:** Quick deployment, modern workflow, generous free tiers

#### Complete Full-Stack Deployment (Recommended)

Deploy this monorepo with separate services for optimal performance:

**🗄️ Step 1: Database on Railway**

1. Create account at [Railway.app](https://railway.app)
2. Create new project
3. Add PostgreSQL database (free tier available)
4. Note your database credentials (Railway provides these automatically)

**⚙️ Step 2: Backend on Railway**

1. Create new Railway project for your backend
2. Connect your GitHub repo
3. Set deployment settings:
   - **Root Directory:** `backend`
   - Railway auto-detects Go applications
4. Set environment variables:
   - `CORS_ALLOWED_ORIGINS=https://your-vercel-app.vercel.app`
   - `JWT_SECRET=your-secret-key` (optional, for authentication)
5. Railway automatically provides database variables:
   - `PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`
6. Deploy and note your backend URL: `https://your-app-name.up.railway.app`

**🌐 Step 3: Frontend on Vercel**

1. Connect your GitHub repo to [Vercel](https://vercel.com)
2. Configure build settings:
   - **Root Directory:** `frontend`
   - **Build Command:** `react-router build`
   - **Output Directory:** `build/client`
   - **Install Command:** `npm install`
3. Set environment variables:
   - `VITE_API_URL=https://your-railway-backend-url.up.railway.app`
   - **Critical:** Use ONLY the base URL, without `/api/*` or trailing paths
4. Deploy frontend

**🎯 Step 4: Connect Services**

- Your frontend will automatically connect to your Railway backend
- Test API endpoints like `/api/health` and `/api/users`

**📋 Important Notes:**

- App uses SPA mode (SSR disabled) for optimal Vercel performance
- `vercel.json` handles client-side routing to prevent 404 errors
- All hardcoded localhost URLs have been replaced with environment variables
- Railway provides database credentials automatically when linked

**Cost:** Both offer generous free tiers, total ~$0-10/month for small apps

### Railway Environment Variables

**PostgreSQL Database Connection:**
Railway provides PostgreSQL environment variables automatically. The backend is configured to use these standard PostgreSQL variable names:

| Railway Variable | Maps to           | Description                 |
| ---------------- | ----------------- | --------------------------- |
| `PGHOST`         | Database Host     | The database server address |
| `PGPORT`         | Database Port     | Usually `5432`              |
| `PGUSER`         | Database User     | Your database username      |
| `PGPASSWORD`     | Database Password | Your database password      |
| `PGDATABASE`     | Database Name     | The name of your database   |

**You don't need to set these manually** - Railway provides them automatically when you link your database to your backend service.

### Railway URL Configuration

**When to use PUBLIC vs PRIVATE URLs:**

- **PUBLIC URL** (`https://your-app.up.railway.app`): 🌐

  - Use this for your Vercel frontend
  - Accessible from anywhere on the internet
  - Required when frontend and backend are on different platforms

- **PRIVATE URL** (`your-app.railway.internal`): 🔒
  - Only accessible within Railway's internal network
  - Use when both services are on Railway (e.g., backend and another Railway service)
  - Faster and more secure for internal communication
  - Cannot be accessed from Vercel

### Alternative Backend Deployment Options

**Option A: Railway (Recommended)**

- ✅ Same platform as your database
- ✅ Easy networking between services
- ✅ Go-native deployment
- ✅ Free tier available
- ✅ Simple scaling

**Option B: Vercel Serverless Functions**

- ✅ Same platform as frontend
- ✅ Simplified CORS handling
- ✅ Automatic scaling
- ❌ Cold starts can be slow
- ❌ Limited execution time (10s for free tier)

**Option C: Other Cloud Platforms**

- **AWS (EC2/Lambda)**: Full control, complex setup
- **Google Cloud Run**: Containerized, good performance
- **DigitalOcean App Platform**: Simple deployment, good for Go
- **Fly.io**: Great for Go apps, global deployment

**Option D: VPS/Cloud Server**

- **DigitalOcean Droplet**: Full control, manual setup
- **AWS EC2**: Scalable, but more complex
- **Linode**: Cost-effective, good performance

#### Common Deployment Issues & Solutions

**❌ ERR_BLOCKED_BY_CLIENT / Connection Refused:**

- **Cause:** Frontend trying to connect to localhost or wrong backend URL
- **Solution:** Set correct `VITE_API_URL` in Vercel environment variables
- **Format:** `https://your-railway-backend.up.railway.app` (base URL only, no `/api/*`)
- **Debug:** Check browser console for API_BASE_URL logs

**❌ 404 NOT_FOUND on API Calls (Double /api):**

- **Cause:** VITE_API_URL includes `/api/*` causing URLs like `/api/*/api/users`
- **Solution:** Use base URL only in VITE_API_URL
- **Example:** `VITE_API_URL=https://your-app.up.railway.app`

**❌ Build Error: "Could not resolve import path":**

- **Cause:** Incorrect relative import paths in components
- **Solution:** Use correct relative paths (e.g., `../../lib/api` from components/demo/)
- **Fix:** Check and correct import statements in affected files

**❌ Railway Database Connection Failed:**

- **Cause:** Backend service not linked to PostgreSQL database in Railway
- **Solution:** In Railway dashboard, link your PostgreSQL to your backend service
- **Result:** Railway automatically injects PGHOST, PGPASSWORD, etc.

**❌ CORS Errors:**

- **Cause:** Backend not configured to allow Vercel frontend origin
- **Solution:** Set `CORS_ALLOWED_ORIGINS=https://your-vercel-app.vercel.app` in Railway backend

**❌ Vercel Build Command Failed:**

- **Cause:** Wrong root directory or build settings
- **Solution:** Ensure Root Directory is set to `frontend` in Vercel project settings

**❌ Error: `sh: line 1: cd: frontend: No such file or directory`:**

- **Solution:** Set Root Directory to `frontend` in Vercel project settings
- **Alternative:** Use repository root with `cd frontend &&` commands

#### Deployment Verification Checklist

**✅ Database Setup:**

- [ ] Railway PostgreSQL database created and linked to backend
- [ ] Backend can connect (check Railway logs for successful DB connection)

**✅ Backend Deployment:**

- [ ] Railway backend deployed successfully
- [ ] Public URL accessible: `https://your-app.up.railway.app/api/health`
- [ ] CORS configured for Vercel frontend

**✅ Frontend Deployment:**

- [ ] Vercel frontend deployed successfully
- [ ] VITE_API_URL set to correct Railway backend URL (base URL only)
- [ ] No hardcoded localhost URLs remain

**✅ Testing:**

- [ ] Frontend loads without console errors
- [ ] API calls work (`/api/health`, `/api/users`)
- [ ] Browser shows correct API_BASE_URL in console
- [ ] Database operations (CRUD) function correctly

## 🚀 Usage

### Development (Local)

- **Frontend**: `npm run dev` → [http://localhost:5173](http://localhost:5173)
- **Backend**: `go run cmd/main.go` → [http://localhost:8080](http://localhost:8080)
- **API Documentation**: [http://localhost:8080/swagger/](http://localhost:8080/swagger/)

### Production (Deployed)

- **Live Demo**: [https://react-golang-starter-kit.vercel.app/](https://react-golang-starter-kit.vercel.app/)
- **Backend API**: Your Railway backend URL
- **API Documentation**: `https://your-railway-backend.up.railway.app/swagger/`

### Testing API Endpoints

- **Health Check**: `/api/health`
- **Users API**: `/api/users`
- **User by ID**: `/api/users/{id}`

The application provides a user management interface where you can create, read, update, and delete users.

### 📚 API Documentation

The backend provides comprehensive API documentation through an interactive Swagger UI:

- **Swagger UI**: [http://localhost:8080/swagger/](http://localhost:8080/swagger/)
- **Direct JSON**: [http://localhost:8080/swagger/doc.json](http://localhost:8080/swagger/doc.json)

**Available Endpoints:**

- `GET /api/health` - Check server health status
- `GET /api/users` - Retrieve all users
- `POST /api/users` - Create a new user
- `GET /api/users/{id}` - Get a specific user by ID
- `PUT /api/users/{id}` - Update an existing user
- `DELETE /api/users/{id}` - Delete a user

The Swagger UI allows you to:

- View detailed endpoint documentation
- Test API endpoints directly from your browser
- See request/response examples and schemas
- Explore the complete API structure

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

```bash
react_golang_starter_kit/
├── backend/                  # 🚀 Golang Backend
│   ├── cmd/
│   │   └── main.go           # Application entry point
│   ├── docs/                 # API documentation
│   │   ├── docs.go
│   │   ├── index.html
│   │   ├── swagger.json
│   │   └── swagger.yaml
│   ├── internal/             # Internal packages
│   │   ├── database/
│   │   │   └── database.go   # Database connection and configuration
│   │   ├── handlers/
│   │   │   └── handlers.go   # HTTP request handlers
│   │   └── models/
│   │       └── models.go     # Data models and GORM structs
│   ├── Dockerfile            # Dockerfile for backend
│   ├── go.mod                # Go module definition
│   ├── go.sum                # Go dependencies checksum
│   ├── Makefile              # Build automation
│   ├── README.md             # Backend documentation
│   └── server                # Compiled server binary
├── frontend/                 # 🌐 React Frontend
│   ├── app/                  # Main application source code
│   │   ├── components/       # Reusable React components
│   │   │   ├── demo/
│   │   │   │   └── demo.tsx
│   │   │   ├── forms/
│   │   │   │   └── UserForm.tsx
│   │   │   └── ui/           # ShadCN UI components
│   │   ├── constants/        # Application constants
│   │   │   ├── icons.ts
│   │   │   ├── labels.ts
│   │   │   └── mockData.ts
│   │   ├── hooks/            # Custom React hooks
│   │   │   ├── use-mobile.ts
│   │   │   └── use-users.ts
│   │   ├── layouts/          # Layout components
│   │   ├── lib/              # Utility functions and API client
│   │   │   ├── api.ts
│   │   │   ├── utils.test.ts
│   │   │   ├── utils.ts
│   │   │   └── zod/          # Zod schemas
│   │   ├── providers/        # React context providers
│   │   │   └── theme-provider.tsx
│   │   ├── root.tsx          # Root component
│   │   ├── routes/           # React Router routes
│   │   │   ├── 404.tsx
│   │   │   ├── custom-layout-demo.tsx
│   │   │   ├── demo.tsx
│   │   │   ├── home.tsx
│   │   │   └── users.tsx
│   │   ├── routes.ts         # Route definitions
│   │   ├── test/
│   │   │   └── setup.ts      # Test configuration
│   │   └── types/
│   │       └── shared.ts     # Shared TypeScript types
│   ├── build/                # Production build output
│   │   ├── client/
│   │   │   ├── assets/       # Built assets
│   │   │   └── favicon.ico
│   │   └── server/
│   │       └── index.js      # Server-side rendering
│   ├── public/               # Static assets
│   │   ├── favicon.ico
│   │   ├── logo-dark.svg
│   │   └── logo-light.svg
│   ├── components.json       # ShadCN configuration
│   ├── Dockerfile            # Dockerfile for frontend
│   ├── package.json          # Node.js package configuration
│   ├── package-lock.json     # Lockfile for dependencies
│   ├── react-router.config.ts # React Router configuration
│   ├── tailwind.config.ts    # TailwindCSS configuration
│   ├── tsconfig.json         # TypeScript configuration
│   ├── vite.config.ts        # Vite configuration
│   └── node_modules/         # Installed dependencies
├── docker-compose.frontend.Dockerfile # Frontend Docker configuration
├── docker-compose.yml        # 🐳 Docker Compose configuration
├── node_modules/             # Root level dependencies
├── package.json              # Root package configuration
├── package-lock.json         # Root lockfile
└── README.md                 # 📄 Project Overview and Setup Instructions
```

## 🔒 Environment Variables

Critical environment variables are managed through `.env` files. A `.env.example` is provided for reference. It is crucial to set these values correctly for the application to function.

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: PostgreSQL database connection details.
- `JWT_SECRET`: Secret key for JWT authentication. **(Highly recommended to change in production!)**
- `API_PORT`: Port on which the Golang backend API will run.
- `VITE_API_URL`: Frontend URL to access the backend API.

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
