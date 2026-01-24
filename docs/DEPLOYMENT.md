# 🚀 Deployment Guide

This guide covers all deployment options for the React-Golang Starter Kit, from quick cloud deployments to self-hosted solutions.

## Quick Deployment Options

Choose your preferred deployment method:

### 🚀 Vercel + Railway (Recommended for Beginners)

**Best for:** Quick setup, modern workflow, generous free tiers

**Services:**
1. **Database**: PostgreSQL on [Railway.app](https://railway.app) (free tier available)
2. **Backend**: Go API on Railway
3. **Frontend**: React app on [Vercel](https://vercel.com)

**Cost:** ~$0-10/month | **Setup Time:** 15-30 minutes

[Jump to detailed guide →](#vercel--railway-step-by-step)

---

### 🐳 Docker + VPS

**Best for:** Full control, cost-effective for production, self-hosted

**Requirements:**
- VPS (DigitalOcean, Linode, AWS EC2, etc.)
- Docker & Docker Compose installed
- Domain name (optional, for SSL)

**Cost:** VPS hosting only (~$5-20/month) | **Setup Time:** 30-60 minutes

[Jump to detailed guide →](#docker--vps-deployment)

---

## Vercel + Railway (Step-by-Step)

### 1. Database Setup on Railway

1. Sign up at [Railway.app](https://railway.app)
2. Create a new project
3. Click **"Add Service"** → **"Database"** → **"PostgreSQL"**
4. Railway will automatically provision and provide connection details
5. Note the credentials (you'll need them for backend deployment)

**Connection Info Location:**
- Navigate to your PostgreSQL service
- Go to **"Variables"** tab
- Copy: `PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`

---

### 2. Backend Deployment on Railway

1. In your Railway project, click **"Add Service"** → **"GitHub Repo"**
2. Connect your GitHub repository
3. Configure build settings:
   - **Root Directory:** `backend`
   - Railway auto-detects Go and handles the build
4. Set environment variables (click **"Variables"** tab):

   ```bash
   DB_HOST=${PGHOST}                    # From PostgreSQL service
   DB_PORT=${PGPORT}                    # From PostgreSQL service
   DB_USER=${PGUSER}                    # From PostgreSQL service
   DB_PASSWORD=${PGPASSWORD}            # From PostgreSQL service
   DB_NAME=${PGDATABASE}                # From PostgreSQL service
   DB_SSLMODE=require                   # Enable SSL for production

   JWT_SECRET=your-secure-random-key    # Generate with: openssl rand -hex 32
   JWT_EXPIRATION_HOURS=24

   CORS_ALLOWED_ORIGINS=https://your-vercel-app.vercel.app
   API_PORT=8080

   RATE_LIMIT_ENABLED=true
   LOG_LEVEL=info
   GO_ENV=production
   ```

5. Deploy and wait for build to complete
6. Note your backend URL: `https://your-app.up.railway.app`
7. Test health endpoint: `https://your-app.up.railway.app/health`

**Pro Tip:** Railway provides automatic HTTPS, health checks, and zero-downtime deployments.

---

### 3. Frontend Deployment on Vercel

1. Go to [Vercel](https://vercel.com) and sign in with GitHub
2. Click **"Add New Project"** → Import your repository
3. Configure project settings:
   - **Framework Preset:** Vite
   - **Root Directory:** `frontend`
   - **Build Command:** `npm run build` (auto-detected)
   - **Output Directory:** `dist` (auto-detected)
   - **Install Command:** `npm install` (auto-detected)

4. Set environment variables:

   ```bash
   VITE_API_URL=https://your-railway-backend.up.railway.app
   NODE_ENV=production
   ```

5. Click **"Deploy"**
6. Wait for build and deployment to complete
7. Your app is live! Vercel provides a URL like: `https://your-app.vercel.app`

**Post-Deployment:**
- Update `CORS_ALLOWED_ORIGINS` in Railway backend with your Vercel URL
- Test authentication flow (register/login)
- Verify API calls work correctly

**Custom Domain (Optional):**
- In Vercel project settings → **"Domains"** → Add your domain
- Update DNS records as instructed by Vercel
- Update `CORS_ALLOWED_ORIGINS` in Railway with custom domain

---

## Docker + VPS Deployment

### Prerequisites

- VPS with Docker installed (Ubuntu 22.04+ recommended)
- SSH access to your VPS
- Domain name pointed to your VPS IP (optional, for SSL)

### 1. Prepare Your VPS

```bash
# SSH into your VPS
ssh user@your-vps-ip

# Update system packages
sudo apt update && sudo apt upgrade -y

# Install Docker and Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install docker-compose-plugin -y

# Verify installation
docker --version
docker compose version
```

### 2. Clone and Configure

```bash
# Clone your repository
git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
cd react-golang-starter-kit

# Create production environment file
cp .env.example .env

# Edit with your production settings
nano .env
```

**Important .env Settings for Production:**

```bash
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=produser                      # Change from default!
DB_PASSWORD=strong-password-here      # Generate strong password!
DB_NAME=proddb
DB_SSLMODE=disable                    # Or 'require' with SSL setup

# JWT Secret - CRITICAL!
JWT_SECRET=generate-secure-key-here   # openssl rand -hex 32
JWT_EXPIRATION_HOURS=24

# API Configuration
API_PORT=8080
VITE_API_URL=http://your-domain.com   # Or https:// with reverse proxy

# CORS - Update with your domain
CORS_ALLOWED_ORIGINS=http://your-domain.com,https://your-domain.com

# Production Settings
GO_ENV=production
NODE_ENV=production
DEBUG=false
LOG_LEVEL=info
LOG_PRETTY=false

# Rate Limiting
RATE_LIMIT_ENABLED=true

# Frontend
FRONTEND_PORT=80
```

### 3. Build and Deploy

```bash
# Build production images
docker compose -f docker-compose.prod.yml build

# Start services
docker compose -f docker-compose.prod.yml up -d

# View logs
docker compose -f docker-compose.prod.yml logs -f

# Check health
curl http://localhost:8080/health
curl http://localhost/
```

### 4. Frontend Deployment

The frontend is deployed **separately** from the backend. Build and deploy to any static hosting:

```bash
# Build frontend (from project root)
make frontend-build

# Or manually:
cd frontend
npm run build
# Deploy the dist/ folder
```

**Deployment Options:**

| Platform | Command / Method |
|----------|-----------------|
| Vercel | `vercel --prod` |
| Cloudflare Pages | Dashboard or `wrangler pages deploy dist` |
| AWS S3 + CloudFront | `aws s3 sync dist s3://your-bucket` |
| Netlify | Dashboard or `netlify deploy --prod` |
| Any web server | Copy `dist/*` to document root |

**Important:** Set `VITE_API_URL` to your backend URL before building:
```bash
VITE_API_URL=https://api.yourdomain.com npm run build
```

### 5. Reverse Proxy & SSL (Optional)

If self-hosting everything on a single server, you need a reverse proxy for:
- TLS termination (HTTPS)
- Routing `/api/*` to backend, `/*` to frontend

**Example with nginx:**

```bash
sudo apt install nginx certbot python3-certbot-nginx
```

Create `/etc/nginx/sites-available/app`:

```nginx
upstream backend {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name yourdomain.com;

    location /api/ {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        proxy_pass http://backend;
    }

    # Serve frontend static files directly or proxy to CDN
    location / {
        root /var/www/html;  # Copy frontend/dist/* here
        try_files $uri $uri/ /index.html;
    }
}
```

Enable and get SSL:
```bash
sudo ln -s /etc/nginx/sites-available/app /etc/nginx/sites-enabled/
sudo certbot --nginx -d yourdomain.com
sudo systemctl reload nginx
```

**Alternative: Cloudflare** - Use Cloudflare as your proxy (handles TLS automatically).

---

## Blue-Green Deployment (Zero Downtime)

For production deployments with zero downtime, use the blue-green deployment script.

### Features

| Feature | Description |
|---------|-------------|
| **Graceful Shutdown** | 30-second connection draining before stopping containers |
| **Deep Health Checks** | Verifies database and cache connectivity, not just HTTP 200 |
| **Fast Rollback** | ~30s rollback using cached images (vs ~3min rebuild) |
| **Auto-Rollback** | Automatically reverts if post-switch validation fails |
| **Post-Switch Validation** | 5 health checks after traffic switch to ensure stability |

### Quick Start

```bash
# Deploy with zero downtime
make prod

# Check deployment status
make prod-status

# Rollback to previous version (fast if image cached)
make rollback
```

### How It Works

1. **Tag Current Image** - Before deploying, the current image is tagged for fast rollback
2. **Deploy to Inactive** - New version deployed to inactive environment (blue or green)
3. **Deep Health Check** - Verifies DB + cache connectivity via `/health/ready`
4. **Traffic Switch** - Gracefully stop old env (30s drain), new env takes over
5. **Post-Switch Validation** - 5 health checks over 10 seconds
6. **Auto-Rollback** - If validation fails, automatically reverts to previous version

### Health Endpoints

| Endpoint | Purpose | Used By |
|----------|---------|---------|
| `/health` | Liveness check (HTTP 200) | Kubernetes, basic monitoring |
| `/health/ready` | Readiness check (DB + cache) | Blue-green deployments, orchestration |

**`/health/ready` Response:**
```json
{
  "status": "healthy",    // healthy, degraded, or unhealthy
  "database": "healthy",  // healthy or unhealthy
  "cache": "healthy"      // healthy, degraded, or unavailable
}
```

### Configuration

Environment variables (set in `.env.prod`):

```bash
BACKEND_PORT=8080        # Port exposed by active backend
DB_USER=produser         # Database credentials
DB_PASSWORD=...
```

Script configuration (in `scripts/deploy-bluegreen.sh`):

```bash
HEALTH_CHECK_RETRIES=15           # Number of health check attempts
HEALTH_CHECK_INTERVAL=2           # Seconds between checks
GRACEFUL_SHUTDOWN_TIMEOUT=30      # Connection drain timeout
POST_SWITCH_CHECKS=5              # Validation checks after switch
POST_SWITCH_THRESHOLD=3           # Minimum checks that must pass
```

### Manual Commands

```bash
# Full deployment script
./scripts/deploy-bluegreen.sh

# Switch traffic without rebuild (if inactive env is running)
./scripts/deploy-bluegreen.sh --switch

# Rollback to previous version
./scripts/deploy-bluegreen.sh --rollback

# Check current status
./scripts/deploy-bluegreen.sh --status

# Show help
./scripts/deploy-bluegreen.sh --help
```

### Troubleshooting

**Deployment fails at health check:**
```bash
# Check container logs
docker logs react-golang-backend-blue
docker logs react-golang-backend-green

# Test health endpoint manually
curl http://localhost:8080/health/ready
```

**Auto-rollback triggered:**
- Check if database is accessible
- Verify cache (Dragonfly/Redis) is running
- Review application logs for startup errors

**No rollback image available:**
- First deployment won't have a cached image
- Rollback will rebuild (slower but still works)

---

## Alternative Deployment Platforms

### Platform Comparison

| Platform                  | Backend         | Frontend      | Database     | Cost/Month | Setup Time | Best For                    |
| ------------------------- | --------------- | ------------- | ------------ | ---------- | ---------- | --------------------------- |
| **Railway + Vercel**      | ✅ Native Go    | ✅ Optimized  | ✅ Built-in  | $0-10      | 15-30 min  | Quick starts, prototypes    |
| **Docker + DigitalOcean** | ✅ Full control | ✅ Custom     | ✅ Managed   | $5-20      | 30-60 min  | Production, cost-effective  |
| **AWS (ECS/Fargate)**     | ✅ Scalable     | ✅ CloudFront | ✅ RDS       | $20-100+   | 60-120 min | Enterprise, auto-scaling    |
| **Google Cloud Run**      | ✅ Serverless   | ✅ Cloud CDN  | ✅ Cloud SQL | $10-50     | 45-90 min  | Serverless, pay-per-use     |
| **Fly.io**                | ✅ Go optimized | ✅ Global CDN | ✅ Built-in  | $5-30      | 20-40 min  | Edge computing, global apps |
| **Render**                | ✅ Auto-deploy  | ✅ Static CDN | ✅ Built-in  | $7-25      | 20-30 min  | Simple, Git-based deploys   |

### Quick Links for Other Platforms

- **AWS**: Use ECS with Fargate for backend, S3+CloudFront for frontend, RDS for database
- **Google Cloud**: Cloud Run for backend/frontend, Cloud SQL for PostgreSQL
- **Fly.io**: `flyctl launch` in backend and frontend directories
- **Render**: Connect GitHub repo, auto-detects Go and Node.js
- **Heroku**: Use buildpacks for Go and Node.js (higher cost)

---

## Troubleshooting Common Issues

### ❌ CORS Errors

**Symptoms:**
- Frontend shows CORS errors in browser console
- API calls return `Access-Control-Allow-Origin` errors

**Solution:**
```bash
# In backend environment variables, ensure:
CORS_ALLOWED_ORIGINS=https://your-frontend-domain.com,http://localhost:5173
```

Update the value with your actual frontend URL(s), comma-separated.

---

### ❌ Database Connection Failed

**Symptoms:**
- Backend logs show `connection refused` or `authentication failed`
- Health check endpoint returns errors

**Solutions:**

1. **Railway/Cloud**: Verify database environment variables:
   ```bash
   # Double-check these match your PostgreSQL service:
   DB_HOST=<correct-host>
   DB_PORT=5432
   DB_USER=<correct-user>
   DB_PASSWORD=<correct-password>
   DB_NAME=<correct-database>
   ```

2. **Docker**: Ensure services are on same network:
   ```bash
   docker compose -f docker-compose.prod.yml ps
   # Backend should show: postgres as hostname
   ```

3. **SSL Mode**: Try different SSL settings:
   ```bash
   DB_SSLMODE=disable    # For local/Docker
   DB_SSLMODE=require    # For cloud databases
   ```

---

### ❌ API Returns 404

**Symptoms:**
- Frontend works but API calls fail with 404
- `/api/auth/login` returns "Not Found"

**Solution:**
```bash
# VITE_API_URL should be base URL only (no /api/ suffix):
VITE_API_URL=https://your-backend.com           ✅ Correct
VITE_API_URL=https://your-backend.com/api       ❌ Wrong

# The /api prefix is added by the frontend code
```

---

### ❌ Vercel Build Fails

**Symptoms:**
- Vercel shows "Build Failed"
- Error: "No package.json found"

**Solution:**
1. In Vercel project settings → **"General"**
2. Set **Root Directory** to `frontend`
3. Redeploy

---

### ❌ Environment Variables Not Working

**Symptoms:**
- App uses default values instead of .env values
- JWT errors or authentication fails

**Solution:**

1. **Vercel**: Environment variables must be set in Vercel dashboard, not .env file
2. **Railway**: Set in Railway dashboard under "Variables" tab
3. **Docker**: Ensure .env file exists and docker-compose references it
4. **Frontend**: Variables must start with `VITE_` prefix:
   ```bash
   VITE_API_URL=...     ✅ Accessible in frontend
   API_URL=...          ❌ Not accessible in frontend
   ```

---

### ❌ High Memory Usage / OOM Errors

**Symptoms:**
- Docker containers crash with "Out of Memory"
- VPS becomes unresponsive

**Solution:**

1. **Reduce Docker resource limits** in docker-compose.prod.yml:
   ```yaml
   deploy:
     resources:
       limits:
         memory: 256M    # Reduce from 512M
   ```

2. **Increase VPS RAM**: Upgrade to larger instance

3. **Enable swap** on VPS:
   ```bash
   sudo fallocate -l 2G /swapfile
   sudo chmod 600 /swapfile
   sudo mkswap /swapfile
   sudo swapon /swapfile
   ```

---

## Deployment Checklist

Use this checklist to verify your deployment:

### Pre-Deployment
- [ ] Repository is up to date on GitHub
- [ ] All tests pass locally
- [ ] Environment variables documented
- [ ] Database migration plan ready (if applicable)

### Database
- [ ] Database created and accessible
- [ ] Credentials secured (not in code)
- [ ] Backups configured (for production)
- [ ] SSL enabled (for cloud databases)

### Backend
- [ ] Environment variables set correctly
- [ ] JWT_SECRET is strong and unique
- [ ] CORS allows frontend origin
- [ ] Health check endpoint accessible (`/health`)
- [ ] API endpoints respond correctly
- [ ] Rate limiting configured
- [ ] Logging configured (production mode)

### Frontend
- [ ] VITE_API_URL points to backend
- [ ] Build completes without errors
- [ ] Static assets load correctly
- [ ] Authentication flow works
- [ ] API calls succeed from browser
- [ ] No console errors in production

### Security
- [ ] SSL/HTTPS enabled (production)
- [ ] Default passwords changed
- [ ] Sensitive data not in environment variables
- [ ] Rate limiting active
- [ ] CORS properly restricted
- [ ] Database not publicly accessible

### Monitoring
- [ ] Application logs accessible
- [ ] Health checks configured
- [ ] Uptime monitoring setup (optional)
- [ ] Error tracking setup (optional)

---

## Post-Deployment

### Verify Deployment

```bash
# Check backend health
curl https://your-backend.com/health

# Test authentication
curl -X POST https://your-backend.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"Test123456"}'

# Check frontend
curl -I https://your-frontend.com
```

### Monitor Logs

**Railway:**
- Dashboard → Your Service → "Logs" tab

**Vercel:**
- Dashboard → Your Project → "Logs" tab

**Docker:**
```bash
docker compose -f docker-compose.prod.yml logs -f
docker compose -f docker-compose.prod.yml logs -f backend
docker compose -f docker-compose.prod.yml logs -f frontend
```

### Update Deployment

**Railway/Vercel:**
- Push to GitHub → Auto-deploys

**Docker VPS:**
```bash
cd react-golang-starter-kit
git pull
docker compose -f docker-compose.prod.yml build
docker compose -f docker-compose.prod.yml up -d
```

---

## Additional Resources

- [Docker Setup Guide](DOCKER_SETUP.md) - Detailed Docker configuration
- [Features Documentation](FEATURES.md) - Authentication, rate limiting, RBAC
- [Frontend Guide](FRONTEND_GUIDE.md) - React/Vite development
- [Backend README](../backend/README.md) - Go backend architecture

---

**Need Help?** Check the troubleshooting section above or open an issue on GitHub.
