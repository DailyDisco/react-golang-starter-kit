# Documentation Hub

Welcome to the React-Golang Starter Kit documentation! This page provides an overview of all available documentation and guides you to the right resources.

## Quick Links

| Document | Purpose | Audience |
|----------|---------|----------|
| [Main README](../README.md) | Project overview and quick start | Everyone |
| [Deployment Guide](DEPLOYMENT.md) | Deploy to Vercel, Railway, VPS, AWS | Everyone |
| [Features Guide](FEATURES.md) | JWT, RBAC, Rate Limiting, File Upload | Developers |
| [Frontend Guide](FRONTEND_GUIDE.md) | React, TanStack Router, Testing | Frontend Developers |
| [Docker Setup](DOCKER_SETUP.md) | Docker development and deployment | DevOps, Developers |
| [Backend README](../backend/README.md) | Go backend architecture | Backend Developers |
| [Development](DEVELOPMENT.md) | History, roadmap, internal notes | Contributors, Maintainers |

---

## Getting Started

### For First-Time Users

1. **Start here:** [Main README](../README.md)
   - Quick start guide
   - Technology stack overview
   - Basic setup instructions

2. **Environment setup:** [Main README - Environment Configuration](../README.md#environment-configuration)
   - `.env.example` - All configuration (required + optional features)

3. **Choose your path:**
   - **Docker users:** [Docker Setup Guide](DOCKER_SETUP.md#quick-start)
   - **Local development:** [Main README - Local Development](../README.md#option-2-local-development)

### For Frontend Developers

1. **Frontend setup:** [Frontend Guide](FRONTEND_GUIDE.md#quick-start)
   - React + Vite + TanStack Router
   - File-based routing conventions
   - Testing with Vitest

2. **Component library:** [ShadCN UI](https://ui.shadcn.com/)
   - Pre-configured components
   - TailwindCSS styling

3. **State management:**
   - Server state: [TanStack Query](https://tanstack.com/query)
   - Client state: [Zustand](https://zustand.pm/)

### For Backend Developers

1. **Backend architecture:** [Backend README](../backend/README.md)
   - Project structure
   - Database setup
   - API endpoints

2. **Authentication & security:** [Features Guide - JWT](FEATURES.md#jwt-authentication--security)
   - JWT token management
   - Password hashing
   - Security best practices

3. **RBAC system:** [Features Guide - RBAC](FEATURES.md#role-based-access-control-rbac)
   - User roles and permissions
   - Protected endpoints
   - Role management

---

## Feature Documentation

### Security & Authentication

**[JWT Authentication & Security](FEATURES.md#jwt-authentication--security)**
- JWT secret management
- Token generation and validation
- Password security
- Migration guide

**[Rate Limiting](FEATURES.md#rate-limiting)**
- IP-based and user-based limits
- Configuration options
- Response headers
- Best practices

**[Role-Based Access Control (RBAC)](FEATURES.md#role-based-access-control-rbac)**
- 4 user roles (User, Premium, Admin, Super Admin)
- Permission-based architecture
- API endpoint access control
- Security features

### File Management

**[File Upload System](FEATURES.md#file-upload-system)**
- Dual storage (AWS S3 / PostgreSQL)
- Secure uploads
- CRUD operations
- Configuration guide

---

## Development Guides

### Docker Development

**[Docker Setup Guide](DOCKER_SETUP.md)**
- Quick start commands
- Development vs Production workflows
- Performance optimizations (BuildKit, caching)
- Resource limits and health checks
- Troubleshooting guide

### Frontend Development

**[Frontend Guide](FRONTEND_GUIDE.md)**
- Technology stack overview
- TanStack Router setup and file-based routing
- Docker development workflow
- Testing with Vitest
- Troubleshooting common issues

### Backend Development

**[Backend README](../backend/README.md)**
- Go project structure
- Database management with GORM
- API endpoint implementation
- Adding new features
- Railway deployment

---

## Deployment

### Deployment Guide

**[Complete Deployment Guide](DEPLOYMENT.md)** - Step-by-step deployment instructions

**Quick Deployment Options:**

1. **Vercel + Railway** (Recommended for beginners)
   - Frontend on Vercel
   - Backend + Database on Railway
   - ~15-30 minutes setup
   - [Step-by-step guide →](DEPLOYMENT.md#vercel--railway-step-by-step)

2. **Docker + VPS** (Full control)
   - Self-hosted solution
   - Production-ready Docker setup
   - ~30-60 minutes setup
   - [Step-by-step guide →](DEPLOYMENT.md#docker--vps-deployment)

3. **Alternative platforms:**
   - AWS (ECS/Fargate)
   - Google Cloud Run
   - Fly.io
   - Render
   - [Platform comparison →](DEPLOYMENT.md#alternative-deployment-platforms)

### Deployment Checklist

**[Full checklist →](DEPLOYMENT.md#deployment-checklist)**

- [ ] Database created and accessible
- [ ] Backend deployed and health check passes
- [ ] Frontend deployed and loads
- [ ] Environment variables configured
- [ ] CORS settings allow frontend origin
- [ ] Authentication flow works
- [ ] Rate limiting configured

---

## Testing

### Frontend Testing

**[Frontend Guide - Testing](FRONTEND_GUIDE.md#testing)**
- Vitest configuration
- Running tests
- Writing component tests
- Coverage reports

### Backend Testing

**[Backend README - Testing](../backend/README.md#testing)**
- Go test commands
- Test coverage
- Package-specific tests

---

## Troubleshooting

### Common Issues

**Frontend:**
- [TanStack Router Issues](FRONTEND_GUIDE.md#route-generation-issues)
- [Docker Problems](FRONTEND_GUIDE.md#docker-issues)
- [Build Issues](FRONTEND_GUIDE.md#build-issues)

**Backend:**
- [Database Connection](../backend/README.md#troubleshooting)
- [Port Conflicts](../backend/README.md#troubleshooting)
- [Module Errors](../backend/README.md#troubleshooting)

**Docker:**
- [Slow Builds](DOCKER_SETUP.md#slow-builds)
- [Disk Space](DOCKER_SETUP.md#out-of-disk-space)
- [Container Restarts](DOCKER_SETUP.md#container-keeps-restarting)
- [Permission Issues](DOCKER_SETUP.md#permission-denied-linux)

**Deployment:**
- [CORS Errors](DEPLOYMENT.md#-cors-errors)
- [Database Connection Failed](DEPLOYMENT.md#-database-connection-failed)
- [API Returns 404](DEPLOYMENT.md#-api-returns-404)
- [Environment Variables](DEPLOYMENT.md#-environment-variables-not-working)

---

## Contributing

### For Contributors

**[Development Documentation](DEVELOPMENT.md)**
- Project roadmap and completed features
- Planned enhancements
- Historical changes and simplifications
- Development priorities

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run tests: `npm test` (frontend) or `go test ./...` (backend)
6. Commit with conventional commits
7. Push and create a Pull Request

### Commit Message Format

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add user profile page
fix: resolve authentication bug
docs: update deployment guide
chore: upgrade dependencies
```

---

## External Resources

### Official Documentation

- [React Documentation](https://react.dev/)
- [TanStack Router](https://tanstack.com/router/latest)
- [TanStack Query](https://tanstack.com/query/latest)
- [Go Documentation](https://golang.org/doc/)
- [Chi Router](https://go-chi.io/)
- [GORM](https://gorm.io/docs/)
- [Docker Documentation](https://docs.docker.com/)
- [Vite](https://vitejs.dev/)

### UI & Styling

- [TailwindCSS](https://tailwindcss.com/docs)
- [ShadCN UI](https://ui.shadcn.com/)
- [Lucide Icons](https://lucide.dev/)

### Testing

- [Vitest](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/react)
- [Go Testing](https://golang.org/pkg/testing/)

---

## Recommended Reading Order

### New to the Project?

1. [Main README](../README.md) - Start here
2. [Features Guide](FEATURES.md) - Understand capabilities
3. [Frontend Guide](FRONTEND_GUIDE.md) or [Backend README](../backend/README.md) - Based on your role
4. [Docker Setup](DOCKER_SETUP.md) - If using Docker

### Planning Production Deployment?

1. [Deployment Guide](DEPLOYMENT.md) - Complete deployment walkthrough
2. [Docker Setup - Production](DOCKER_SETUP.md#production)
3. [Features Guide - JWT Security](FEATURES.md#jwt-authentication--security)
4. [Features Guide - Rate Limiting](FEATURES.md#rate-limiting)

### Want to Contribute?

1. [Development Documentation](DEVELOPMENT.md)
2. [Backend README](../backend/README.md)
3. [Frontend Guide](FRONTEND_GUIDE.md)
4. [Main README - Testing](../README.md#testing)

---

## Need Help?

### Documentation Not Clear?

- Check the [troubleshooting sections](#troubleshooting) above
- Review the [Main README FAQ](../README.md#troubleshooting-guides)
- Search through the specific guides

### Found a Bug or Issue?

- Check existing documentation for solutions
- Review the [development roadmap](DEVELOPMENT.md#project-roadmap)
- Create an issue with detailed information

### Want to Suggest Improvements?

- Review [future recommendations](DEVELOPMENT.md#future-recommendations)
- Create a Pull Request with your improvements
- Update relevant documentation

---

## Documentation Structure

```
react_golang_starter_kit/
├── README.md                    # Main entry point
├── backend/
│   └── README.md               # Backend-specific guide
└── docs/
    ├── README.md               # This file - Documentation hub
    ├── DEPLOYMENT.md           # Deployment guides (Vercel, Railway, VPS)
    ├── FEATURES.md             # JWT, RBAC, Rate Limiting, Files
    ├── FRONTEND_GUIDE.md       # React, TanStack Router, Testing
    ├── DOCKER_SETUP.md         # Docker development & deployment
    └── DEVELOPMENT.md          # History, roadmap, contributors
```

---

## Keep Documentation Updated

When making changes to the project:

1. ✅ Update relevant documentation files
2. ✅ Keep code examples accurate
3. ✅ Update version numbers if applicable
4. ✅ Test all commands and procedures
5. ✅ Add troubleshooting tips for new issues

Good documentation makes the project accessible to everyone. Thank you for keeping it up to date!
