# syntax=docker/dockerfile:1.6
# Multi-stage build for Next.js web application
# BuildKit required: DOCKER_BUILDKIT=1

#############################################
# STAGE 1: Dependencies
#############################################
FROM node:22-alpine AS deps

# Security: Install dumb-init for proper signal handling
RUN apk add --no-cache libc6-compat dumb-init

WORKDIR /app

# Copy package files for layer caching
COPY package.json package-lock.json* ./

# Install dependencies with npm (repo uses package-lock.json)
RUN npm ci --legacy-peer-deps --omit=dev

#############################################
# STAGE 2: Builder
#############################################
FROM node:22-alpine AS builder

WORKDIR /app

# Copy dependencies from deps stage
COPY --from=deps /app/node_modules ./node_modules

# Copy source code
COPY . .

# Build arguments for environment variables
ARG NEXT_PUBLIC_API_URL
ARG NEXT_PUBLIC_WS_URL
ARG NEXT_PUBLIC_APP_VERSION

ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL}
ENV NEXT_PUBLIC_WS_URL=${NEXT_PUBLIC_WS_URL}
ENV NEXT_PUBLIC_APP_VERSION=${NEXT_PUBLIC_APP_VERSION}
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production

# Build the application
RUN npm run build

#############################################
# STAGE 3: Production Image
#############################################
FROM node:22-alpine AS runner

# Security: Install dumb-init for proper signal handling
RUN apk add --no-cache dumb-init ca-certificates

# Create non-root user
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

WORKDIR /app

# Set environment
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

# Copy built application
COPY --from=builder --chown=nextjs:nodejs /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# Copy necessary files
COPY --from=builder --chown=nextjs:nodejs /app/package.json ./package.json

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --spider -q http://localhost:3000/api/health || exit 1

# Use dumb-init for proper signal handling
ENTRYPOINT ["dumb-init", "--"]

# Start the server
CMD ["node", "server.js"]
