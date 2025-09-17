FROM golang:1.25-alpine AS builder

WORKDIR /app

# Dependencias necesarias para compilar Go
RUN apk add --no-cache git build-base

# Copiar mod files primero (cache)
COPY . .
RUN go mod tidy

# Copiar código fuente

# Compilar binario
RUN go build -o /app/tmp/main ./cmd/main.go

FROM alpine:latest

# Install system dependencies (Python >= 3.11, ffmpeg, git, curl, wget, bash, build tools)
RUN apk add --no-cache \
    bash \
    wget \
    curl \
    git \
    ffmpeg \
    python3 \
    py3-pip \
    ca-certificates \
    build-base \
    && pip3 install --no-cache --break-system-packages --upgrade pip setuptools wheel \
    && update-ca-certificates

# Install Go
RUN wget https://go.dev/dl/go1.25.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.25.1.linux-amd64.tar.gz && \
    rm go1.25.1.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:/go/bin:/home/user/go/bin:$PATH"
ENV GOPATH="/go"
ENV PATH="$GOPATH/bin:$PATH"

# Install yt-dlp
RUN pip3 install --no-cache --break-system-packages -U "yt-dlp[default]"
# Install Air (hot reload)
RUN go install github.com/air-verse/air@latest

WORKDIR /app

COPY --from=builder /app/tmp/main .
COPY . .
# Copy source files
RUN go mod tidy



