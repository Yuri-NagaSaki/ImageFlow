FROM golang:1.22-alpine AS backend-builder

ENV GO111MODULE=on

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o imageflow

FROM oven/bun:1 AS frontend-builder

WORKDIR /app/frontend

COPY frontend/ /app/frontend/

RUN bun install --frozen-lockfile

RUN bun run build

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache \
    ca-certificates \
    libwebp-tools \
    libavif-apps \
    wget

RUN mkdir -p /app/static/images/metadata && \
    mkdir -p /app/static/images/original/landscape && \
    mkdir -p /app/static/images/original/portrait && \
    mkdir -p /app/static/images/landscape/webp && \
    mkdir -p /app/static/images/landscape/avif && \
    mkdir -p /app/static/images/portrait/webp && \
    mkdir -p /app/static/images/portrait/avif

COPY --from=backend-builder /app/imageflow /app/
COPY --from=backend-builder /app/config /app/config
COPY --from=frontend-builder /app/frontend/out /app/static
COPY --from=frontend-builder /app/frontend/public/favicon* /app/static/

ENV API_KEY=""
ENV STORAGE_TYPE="local"
ENV LOCAL_STORAGE_PATH="/app/static/images"
ENV S3_ENDPOINT=""
ENV S3_REGION=""
ENV S3_ACCESS_KEY=""
ENV S3_SECRET_KEY=""
ENV S3_BUCKET=""
ENV CUSTOM_DOMAIN=""
ENV MAX_UPLOAD_COUNT="20"
ENV IMAGE_QUALITY="80"
ENV WORKER_THREADS="4"
ENV COMPRESSION_EFFORT="6"
ENV FORCE_LOSSLESS="false"

EXPOSE 8686

CMD ["./imageflow"]


