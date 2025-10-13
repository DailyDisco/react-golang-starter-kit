# Docker Setup Guide

Complete guide for Docker development and deployment with this project.

## Table of Contents
- [Quick Start](#quick-start)
- [Development](#development)
- [Production](#production)
- [Performance Optimizations](#performance-optimizations)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

### Development

```bash
# Start all services
docker compose up

# Start in background
docker compose up -d

# Rebuild and start
docker compose up --build

# View logs
docker compose logs -f

# Stop services
docker compose down
```

### Production

```bash
# Build production images
docker compose -f docker-compose.prod.yml build

# Start production
docker compose -f docker-compose.prod.yml up -d

# For staging: Use prod compose with staging .env file
# docker compose -f docker-compose.prod.yml --env-file .env.staging up -d
```

### Using the Build Script

```bash
# Make executable (first time only)
chmod +x docker-build.sh

# Build development
./docker-build.sh dev

# Build production
./docker-build.sh prod

# Build specific service
./docker-build.sh dev backend

# Clean unused resources
./docker-build.sh clean

# Show disk usage
./docker-build.sh stats
```

---

## Development

### Resource Limits

| Service  | CPU Limit | Memory Limit | CPU Reserve | Memory Reserve |
|----------|-----------|--------------|-------------|----------------|
| Postgres | 1.0       | 512M         | 0.5         | 256M           |
| Backend  | 2.0       | 1G           | 0.5         | 256M           |
| Frontend | 2.0       | 2G           | 0.5         | 512M           |

### Hot Reloading

- **Backend**: Uses Air for automatic Go code reloading
- **Frontend**: Uses Vite dev server with HMR (Hot Module Replacement)

### Workflow

```bash
# Start development environment
docker compose up -d

# View logs (follow mode)
docker compose logs -f backend
docker compose logs -f frontend

# Rebuild specific service
docker compose build backend
docker compose up -d backend

# Execute commands in container
docker compose exec backend go test ./...
docker compose exec frontend npm test
```

---

## Production

### Image Sizes

| Service | Target | Size | Notes |
|---------|--------|------|-------|
| Backend | production | ~25MB | Alpine + binary |
| Frontend| production | ~30MB | Nginx + static files |

### Production Deployment

```bash
# Enable BuildKit for faster builds
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Build production images
docker compose -f docker-compose.prod.yml build

# Start services
docker compose -f docker-compose.prod.yml up -d

# View logs
docker compose -f docker-compose.prod.yml logs -f

# Stop services
docker compose -f docker-compose.prod.yml down
```

### Environment Configuration

Production uses `.env` file with stricter settings:

```bash
# Production environment
LOG_LEVEL=info
DEBUG=false
RATE_LIMIT_ENABLED=true

# Staging: Use debug settings
LOG_LEVEL=debug
DEBUG=true
```

### Health Checks

All services include health checks:
- **Postgres**: `pg_isready` check every 30s
- **Backend**: HTTP check on `/health` endpoint
- **Frontend**: HTTP check on root path

---

## Performance Optimizations

### 1. BuildKit Cache Mounts (50-70% Faster Builds)

**Enable BuildKit:**
```bash
# Option 1: Current session
source .dockerbuildkit.env

# Option 2: Permanent (add to ~/.bashrc or ~/.zshrc)
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
```

**What it does:**
- Persistent caching of package downloads across builds
- Backend: Go module cache and build cache
- Frontend: npm cache

### 2. Multi-Stage Builds

**Backend Stages:**
- `development`: Full dev environment with Air hot-reloading
- `production`: Alpine-based (~25MB) with health checks

**Frontend Stages:**
- `development`: Full Node.js environment with Vite dev server
- `production`: Nginx Alpine serving static files (~30MB)

### 3. Optimized .dockerignore

Excludes:
- Test files and test data
- Documentation (*.md files)
- Development tools and configs
- Git history and CI/CD configs
- Build artifacts and cache directories

**Impact:** Faster build context transfer and smaller images

### 4. Layer Caching Strategy

- Dependencies installed before copying source code
- Less frequently changing layers placed earlier
- Named volumes for persistent caches

### Build Time Comparison

| Scenario | Without Optimization | With Optimization | Improvement |
|----------|---------------------|-------------------|-------------|
| Fresh build | 5-8 minutes | 5-8 minutes | - |
| Rebuild (no changes) | 2-3 minutes | 10-20 seconds | 85-90% |
| Rebuild (code changes) | 2-3 minutes | 30-45 seconds | 70-75% |
| Rebuild (deps change) | 5-8 minutes | 1-2 minutes | 60-75% |

---

## Troubleshooting

### Slow Builds?

```bash
# Check BuildKit is enabled
echo $DOCKER_BUILDKIT  # Should output: 1

# Enable if not set
source .dockerbuildkit.env

# View build cache usage
docker buildx du

# Clear and rebuild cache
docker builder prune -a
docker compose build --no-cache
```

### Out of Disk Space?

```bash
# Show usage
docker system df

# Clean safely (keeps recent builds)
docker image prune
docker builder prune --keep-storage 10GB

# Aggressive cleanup (removes everything!)
docker system prune -a --volumes
```

### Container Keeps Restarting?

```bash
# Check logs
docker compose logs [service-name]

# Check resource usage
docker stats

# View container status
docker compose ps

# Inspect container
docker inspect <container-id>
```

### Changes Not Reflecting?

```bash
# Rebuild without cache
docker compose build --no-cache [service]

# Or use the script
./docker-build.sh no-cache

# For frontend, clear Vite cache
docker compose exec frontend rm -rf node_modules/.vite
```

### Port Already in Use

```bash
# Find process using port
sudo lsof -i :8080  # Replace with your port

# Change port in .env file
API_PORT=8081  # Backend
# Or stop the conflicting service
```

### Permission Denied (Linux)

```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Logout and login again for changes to take effect
```

### Out of Memory Errors

**Symptoms:**
- Container killed with exit code 137
- "OOMKilled" in container inspect

**Solutions:**
```bash
# Check current limits
docker inspect <container-name> | grep -A 10 Memory

# Increase limits in docker-compose.yml
# Edit deploy.resources.limits.memory value

# Or increase Docker Desktop resources (Mac/Windows)
# Docker Desktop → Preferences → Resources → Memory
```

### BuildKit Not Working

```bash
# Check BuildKit is enabled
docker buildx version

# Check environment variables
echo $DOCKER_BUILDKIT
echo $COMPOSE_DOCKER_CLI_BUILD

# Restart Docker daemon
sudo systemctl restart docker  # Linux
# or restart Docker Desktop on Mac/Windows
```

---

## Monitoring and Maintenance

### Check Resource Usage

```bash
# Container resource usage
docker stats

# Disk usage by Docker
docker system df

# Detailed build cache info
docker buildx du --verbose
```

### Cleanup Strategies

**Regular maintenance (weekly):**
```bash
# Remove dangling images
docker image prune

# Remove build cache older than 24h
docker builder prune --keep-storage 10GB
```

**Deep cleanup (monthly):**
```bash
# Remove all unused images, containers, networks
docker system prune -a

# Remove all build cache
docker builder prune -a
```

---

## Best Practices

1. **Always use BuildKit** - Enable it globally for best performance
2. **Regular cleanup** - Run `docker builder prune` weekly
3. **Monitor resources** - Use `docker stats` to identify issues
4. **Use .dockerignore** - Keep build context small
5. **Named volumes** - Use for caches and data persistence
6. **Health checks** - Always configure for production
7. **Resource limits** - Set appropriate limits for all services
8. **Layer optimization** - Put rarely changing operations first

---

## Quick Reference Commands

```bash
# Enable BuildKit
source .dockerbuildkit.env

# Development
docker compose up --build
docker compose logs -f
docker compose down

# Production
docker compose -f docker-compose.prod.yml build
docker compose -f docker-compose.prod.yml up -d

# Monitoring
docker stats
docker system df
docker buildx du

# Cleanup
docker builder prune
docker system prune -a

# Rebuild without cache
docker compose build --no-cache

# Execute commands
docker compose exec backend sh
docker compose exec frontend sh
```

---

## Need Help?

Run the build script help:
```bash
./docker-build.sh help
```

Check logs for errors:
```bash
docker compose logs -f [service-name]
```
