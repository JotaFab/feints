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

WORKDIR /app

# Copy Go module files first for caching
COPY go.mod ./
# Copy go.sum only if it exists
# COPY go.sum ./
RUN go mod download || true

# Install Air (hot reload)
RUN go install github.com/air-verse/air@latest

# Copy source files
COPY . .

RUN go mod tidy

CMD ["air", "-c", ".air.toml"]
