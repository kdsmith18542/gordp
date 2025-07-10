# Multi-stage build for GoRDP with GUI support
# Stage 1: Build the application
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Build GUI application
RUN make build-gui

# Stage 2: Qt Builder (for Qt GUI)
FROM ubuntu:22.04 AS qt-builder

# Install Qt6 and build dependencies
RUN apt-get update && apt-get install -y \
    qt6-base-dev \
    qt6-tools-dev \
    qt6-websockets-dev \
    qt6-charts-dev \
    qt6-declarative-dev \
    cmake \
    build-essential \
    pkg-config \
    git \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Build Qt GUI
RUN if [ -d "qt-gui" ]; then \
        cd qt-gui && \
        mkdir -p build && \
        cd build && \
        cmake .. -DCMAKE_BUILD_TYPE=Release && \
        make -j$(nproc); \
    fi

# Stage 3: Runtime image
FROM ubuntu:22.04

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    libqt6core6 \
    libqt6widgets6 \
    libqt6network6 \
    libqt6websockets6 \
    libqt6charts6 \
    libqt6declarative6 \
    libssl3 \
    libasound2 \
    pulseaudio \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -g 1001 gordp && \
    useradd -u 1001 -g gordp -m -s /bin/bash gordp

# Set working directory
WORKDIR /app

# Copy binaries from builder stages
COPY --from=builder /app/build/gordp /usr/local/bin/gordp
COPY --from=builder /app/build/gui/gordp-gui /usr/local/bin/gordp-gui

# Copy Qt GUI if built
COPY --from=qt-builder /app/qt-gui/build/bin/gordp-gui /usr/local/bin/gordp-qt-gui 2>/dev/null || echo "Qt GUI not built"

# Copy example configurations
COPY --from=builder /app/examples /app/examples

# Set ownership
RUN chown -R gordp:gordp /app

# Switch to non-root user
USER gordp

# Expose default RDP port (for documentation)
EXPOSE 3389

# Set entrypoint
ENTRYPOINT ["gordp"]

# Default command (show help)
CMD ["--help"]

# Labels
LABEL maintainer="kdsmith18542 <kdsmith18542@github.com>"
LABEL description="GoRDP - Production-grade RDP client in Go with Qt GUI"
LABEL version="latest"
LABEL org.opencontainers.image.source="https://github.com/kdsmith18542/gordp" 