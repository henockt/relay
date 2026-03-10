# 1. Build frontend
FROM node:20-alpine AS web-builder
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
ENV NEXT_TELEMETRY_DISABLED=1
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
RUN npm run build

# 2. Build backend
FROM golang:1.25-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server

# 3. Final image
FROM node:20-alpine
WORKDIR /app

# Go binary
COPY --from=go-builder /app/server ./server

# Next.js standalone output
COPY --from=web-builder /web/.next/standalone ./web-standalone
COPY --from=web-builder /web/.next/static ./web-standalone/.next/static
COPY --from=web-builder /web/public ./web-standalone/public

EXPOSE 8080
CMD PORT=3000 node /app/web-standalone/server.js & exec /app/server