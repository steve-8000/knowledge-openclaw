# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------------
# Dependencies
# ---------------------------------------------------------------------------
FROM node:20-alpine AS deps
WORKDIR /app
COPY apps/dashboard/package.json apps/dashboard/package-lock.json* ./
RUN npm ci --prefer-offline

# ---------------------------------------------------------------------------
# Build
# ---------------------------------------------------------------------------
FROM node:20-alpine AS builder
WORKDIR /app

ARG NEXT_PUBLIC_QUERY_API_URL=http://localhost:8081
ARG NEXT_PUBLIC_INGEST_API_URL=http://localhost:8080
ENV NEXT_PUBLIC_QUERY_API_URL=${NEXT_PUBLIC_QUERY_API_URL}
ENV NEXT_PUBLIC_INGEST_API_URL=${NEXT_PUBLIC_INGEST_API_URL}

COPY --from=deps /app/node_modules ./node_modules
COPY apps/dashboard/ .
RUN npm run build

# ---------------------------------------------------------------------------
# Runtime
# ---------------------------------------------------------------------------
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000
ENV PORT=3000

CMD ["node", "server.js"]
