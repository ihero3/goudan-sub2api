#!/bin/bash
#
# Sub2API Build Script
# Sub2API 构建脚本
#
# Usage / 用法:
#   ./deploy/build.sh              # Build with embed (frontend embedded in binary)
#   ./deploy/build.sh --no-embed   # Build without embed (frontend served separately)
#   ./deploy/build.sh --clean      # Clean build artifacts before building
#
# Requirements / 构建依赖:
#   - Go 1.21+
#   - Node.js 18+ / pnpm 8+
#

set -e

# ============================================================
# Configuration
# ============================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
EMBED_DIST_DIR="$BACKEND_DIR/internal/web/dist"
OUTPUT_DIR="$PROJECT_ROOT/bin"

# Build flags
VERSION="${VERSION:-$(cat "$BACKEND_DIR/cmd/server/VERSION" 2>/dev/null || echo 'dev')}"
LDFLAGS="-s -w -X main.Version=$VERSION"
EMBED_TAG="embed"

# Flags
BUILD_EMBED=true
CLEAN_FIRST=false

# ============================================================
# Color output
# ============================================================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
error()   { echo -e "${RED}[ERROR]${NC} $1"; }

# ============================================================
# Parse arguments
# ============================================================
while [[ $# -gt 0 ]]; do
    case "$1" in
        --no-embed)
            BUILD_EMBED=false
            shift
            ;;
        --clean)
            CLEAN_FIRST=true
            shift
            ;;
        --help|-h)
            echo "Sub2API Build Script"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --no-embed   Build without embedding frontend (use Nginx to serve frontend)"
            echo "  --clean      Clean build artifacts before building"
            echo "  --help       Show this help message"
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# ============================================================
# Step 0: Check dependencies
# ============================================================
check_deps() {
    info "Checking build dependencies..."

    # Check Go
    if ! command -v go &>/dev/null; then
        error "Go is not installed. Please install Go 1.21+ first."
        echo "  Download: https://go.dev/dl/"
        echo "  Ubuntu:   sudo apt install -y golang-go"
        exit 1
    fi
    info "Go version: $(go version)"

    # Check Node.js and pnpm (only needed for embed build)
    if [ "$BUILD_EMBED" = true ]; then
        if ! command -v node &>/dev/null; then
            error "Node.js is not installed. Please install Node.js 18+ first."
            echo "  Download: https://nodejs.org/"
            echo "  Ubuntu:   curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash - && sudo apt install -y nodejs"
            exit 1
        fi
        info "Node.js version: $(node --version)"

        if ! command -v pnpm &>/dev/null; then
            error "pnpm is not installed. Please install pnpm 8+ first."
            echo "  Install:  npm install -g pnpm"
            echo "  Or:       curl -fsSL https://get.pnpm.io/install.sh | sh -"
            exit 1
        fi
        info "pnpm version: $(pnpm --version)"
    fi
}

# ============================================================
# Step 1: Clean (optional)
# ============================================================
clean_artifacts() {
    if [ "$CLEAN_FIRST" = true ]; then
        info "Cleaning build artifacts..."
        rm -rf "$OUTPUT_DIR"
        rm -rf "$FRONTEND_DIR/dist"
        rm -rf "$FRONTEND_DIR/node_modules/.vite"
        rm -rf "$EMBED_DIST_DIR"
        success "Clean complete"
    fi
}

# ============================================================
# Step 2: Build frontend
# ============================================================
build_frontend() {
    if [ "$BUILD_EMBED" = false ]; then
        warn "Skipping frontend build (--no-embed mode)"
        return 0
    fi

    info "Building frontend..."

    # Install dependencies if needed
    if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
        info "Installing frontend dependencies..."
        pnpm --dir "$FRONTEND_DIR" install --frozen-lockfile
    fi

    # Build
    pnpm --dir "$FRONTEND_DIR" run build

    if [ ! -d "$FRONTEND_DIR/dist" ]; then
        error "Frontend build failed: dist directory not found"
        exit 1
    fi

    success "Frontend build complete"

    # Copy dist to embed directory
    info "Copying frontend dist to embed directory..."
    mkdir -p "$EMBED_DIST_DIR"
    rm -rf "$EMBED_DIST_DIR"/*
    cp -r "$FRONTEND_DIR/dist/"* "$EMBED_DIST_DIR/"
    success "Frontend dist copied to $EMBED_DIST_DIR"
}

# ============================================================
# Step 3: Build backend
# ============================================================
build_backend() {
    info "Building backend..."

    mkdir -p "$OUTPUT_DIR"

    cd "$BACKEND_DIR"

    if [ "$BUILD_EMBED" = true ]; then
        info "Building with embed tag (frontend embedded)..."
        CGO_ENABLED=0 go build -tags "$EMBED_TAG" -ldflags="$LDFLAGS" -trimpath -o "$OUTPUT_DIR/sub2api" ./cmd/server
        success "Backend built with embedded frontend: $OUTPUT_DIR/sub2api"
    else
        info "Building without embed tag (frontend not embedded)..."
        CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -trimpath -o "$OUTPUT_DIR/sub2api" ./cmd/server
        success "Backend built (no embed): $OUTPUT_DIR/sub2api"
    fi

    cd "$PROJECT_ROOT"

    # Show binary info
    local binary_size
    binary_size=$(du -h "$OUTPUT_DIR/sub2api" | cut -f1)
    info "Binary size: $binary_size"
    info "Version: $VERSION"
}

# ============================================================
# Step 4: Summary
# ============================================================
print_summary() {
    echo ""
    echo "=============================================="
    success "Build Complete!"
    echo "=============================================="
    echo ""
    echo "  Binary:     $OUTPUT_DIR/sub2api"
    echo "  Version:    $VERSION"
    echo "  Embed:      $BUILD_EMBED"
    echo ""
    if [ "$BUILD_EMBED" = true ]; then
        echo "  The binary contains the embedded frontend."
        echo "  You can deploy it directly without Nginx for static files."
    else
        echo "  The binary does NOT contain the frontend."
        echo "  You need Nginx to serve frontend static files."
        echo "  Frontend dist: $FRONTEND_DIR/dist/"
    fi
    echo ""
    echo "  Next: Run deploy-ubuntu.sh to deploy to server."
    echo "=============================================="
}

# ============================================================
# Main
# ============================================================
main() {
    echo ""
    echo "=============================================="
    echo "       Sub2API Build Script"
    echo "=============================================="
    echo ""

    check_deps
    clean_artifacts
    build_frontend
    build_backend
    print_summary
}

main "$@"
