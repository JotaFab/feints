FROM debian:bullseye


# Install system dependencies
RUN apt-get update && \
	apt-get install -y wget curl python3 python3-pip ffmpeg ca-certificates && \
	rm -rf /var/lib/apt/lists/*

# Install Go
RUN wget https://go.dev/dl/go1.25.1.linux-amd64.tar.gz && \
	tar -C /usr/local -xzf go1.25.1.linux-amd64.tar.gz && \
	rm go1.25.1.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:$PATH"

# Install yt-dlp
RUN pip3 install yt-dlp

# Create non-root user for app
RUN useradd -ms /bin/bash feints
USER feints
WORKDIR /home/feints/app

# Copy source files
COPY --chown=feints:feints . .

# Download Go dependencies
RUN go mod download

RUN go mod tidy
# Default command (customize as needed)
RUN go build /home/feints/app/cmd/bot/main.go
CMD ["./main"]