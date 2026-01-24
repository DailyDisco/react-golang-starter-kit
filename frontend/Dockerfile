# ============================================
# React Frontend Dockerfile
# Multi-stage build for development and production
# ============================================

# Development stage
FROM node:22-alpine AS development

# Install curl for healthchecks
RUN apk add --no-cache curl

# Set working directory
WORKDIR /app

# Copy package files for better caching
COPY package*.json ./

# Install dependencies with BuildKit cache mount for faster rebuilds
RUN --mount=type=cache,target=/root/.npm \
    npm ci --no-audit --no-fund

# Copy source code
COPY . .

# Copy environment variables for build
ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL
ENV NODE_ENV=development

# Expose ports
EXPOSE 5173 4173

# Default command for development (can be overridden by docker-compose)
# CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]

# Production dependencies stage
FROM node:22-alpine AS production-dependencies

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install only production dependencies with BuildKit cache
RUN --mount=type=cache,target=/root/.npm \
    npm ci --omit=dev --no-audit --no-fund

# Build stage
FROM node:22-alpine AS build

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install all dependencies (including dev dependencies) with BuildKit cache
RUN --mount=type=cache,target=/root/.npm \
    npm ci --no-audit --no-fund

# Copy source code
COPY . .

# Copy environment variables for build
ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL
ENV NODE_ENV=production

# Build the application with optimizations
RUN npm run build

# Production stage - minimal image holding built static files
# Use: docker build --target production -o ./dist ./frontend
# Or simply run: npm run build
FROM scratch AS production
COPY --from=build /app/dist /
