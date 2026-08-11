#!/bin/bash
#
# Sub2API Ubuntu Server Setup Script
# Sub2API Ubuntu 服务器环境初始化脚本
#
# This script installs and configures:
#   - PostgreSQL (database)
#   - Redis (cache)
#   - Nginx (reverse proxy)
#   - sub2api system user and directories
#
# Usage / 用法:
#   sudo bash setup-ubuntu.sh                    # Interactive setup
#   sudo bash setup-ubuntu.sh --non-interactive   # Non-interactive (use defaults)
#

set -e

# ============================================================
# Configuration
# ============================================================
INSTALL_DIR="/opt/sub2api"
CONFIG_DIR="/etc/sub2api"
SERVICE_USER="sub2api"
DB_NAME="sub2api"
DB_USER="sub2api"
DB_PASS=""
REDIS_PASS=""
SERVER_PORT=8080
NON_INTERACTIVE=false

# ============================================================
# Color output
# ============================================================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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
        --non-interactive)
            NON_INTERACTIVE=true
            shift
            ;;
        --db-pass)
            DB_PASS="$2"
            shift 2
            ;;
        --redis-pass)
            REDIS_PASS="$2"
            shift 2
            ;;
        --help|-h)
            echo "Sub2API Ubuntu Server Setup Script"
            echo ""
            echo "Usage: sudo bash $0 [options]"
            echo ""
            echo "Options:"
            echo "  --non-interactive    Use default settings without prompts"
            echo "  --db-pass <pass>     Set PostgreSQL password"
            echo "  --redis-pass <pass>  Set Redis password"
            echo "  --help               Show this help message"
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# ============================================================
# Check root
# ============================================================
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        error "Please run as root (use sudo)"
        exit 1
    fi
}

# ============================================================
# Detect Ubuntu version
# ============================================================
detect_os() {
    if [ ! -f /etc/os-release ]; then
        error "This script is designed for Ubuntu. /etc/os-release not found."
        exit 1
    fi

    source /etc/os-release
    info "OS: $NAME $VERSION (Codename: $VERSION_CODENAME)"

    if [[ "$ID" != "ubuntu" ]]; then
        warn "This script is designed for Ubuntu, but detected: $ID"
        warn "Continuing anyway, but some commands may need adjustment."
    fi
}

# ============================================================
# Interactive configuration
# ============================================================
configure() {
    if [ "$NON_INTERACTIVE" = true ]; then
        # Generate random passwords if not provided
        if [ -z "$DB_PASS" ]; then
            DB_PASS=$(openssl rand -hex 16)
            info "Generated database password: $DB_PASS"
        fi
        if [ -z "$REDIS_PASS" ]; then
            REDIS_PASS=""
            info "Redis: no password (default)"
        fi
        return 0
    fi

    echo ""
    echo -e "${CYAN}=============================================="
    echo "  Sub2API Server Configuration"
    echo "==============================================${NC}"
    echo ""

    # Database password
    if [ -z "$DB_PASS" ]; then
        echo -e "${YELLOW}Set a password for the PostgreSQL database user '${DB_USER}'.${NC}"
        read -s -p "Database password (leave empty to auto-generate): " input_db_pass
        echo
        if [ -z "$input_db_pass" ]; then
            DB_PASS=$(openssl rand -hex 16)
            info "Auto-generated database password: $DB_PASS"
        else
            DB_PASS="$input_db_pass"
        fi
    fi

    echo ""

    # Redis password
    if [ -z "$REDIS_PASS" ]; then
        echo -e "${YELLOW}Set a password for Redis (recommended for production).${NC}"
        read -s -p "Redis password (leave empty for no password): " input_redis_pass
        echo
        REDIS_PASS="$input_redis_pass"
    fi

    echo ""
    info "Configuration summary:"
    info "  Database: $DB_NAME (user: $DB_USER)"
    info "  Redis: ${REDIS_PASS:+with password}${REDIS_PASS:-no password}"
    echo ""
}

# ============================================================
# Install system packages
# ============================================================
install_packages() {
    info "Updating package index..."
    apt-get update -qq

    info "Installing PostgreSQL, Redis, Nginx, and utilities..."
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
        postgresql \
        postgresql-contrib \
        redis-server \
        nginx \
        ufw \
        curl \
        wget \
        ca-certificates \
        openssl \
        >/dev/null 2>&1

    success "Packages installed: PostgreSQL, Redis, Nginx"
}

# ============================================================
# Configure PostgreSQL
# ============================================================
setup_postgresql() {
    info "Configuring PostgreSQL..."

    # Ensure PostgreSQL is running
    systemctl start postgresql
    systemctl enable postgresql >/dev/null 2>&1

    # Create database and user
    sudo -u postgres psql -c "CREATE USER \"$DB_USER\" WITH PASSWORD '$DB_PASS';" 2>/dev/null || \
        warn "Database user '$DB_USER' already exists"

    sudo -u postgres psql -c "CREATE DATABASE \"$DB_NAME\" OWNER \"$DB_USER\";" 2>/dev/null || \
        warn "Database '$DB_NAME' already exists"

    # Grant privileges
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE \"$DB_NAME\" TO \"$DB_USER\";" >/dev/null

    # Configure pg_hba.conf for password authentication
    PG_HBA=$(find /etc/postgresql -name "pg_hba.conf" 2>/dev/null | head -1)
    if [ -n "$PG_HBA" ]; then
        # Backup original
        cp "$PG_HBA" "${PG_HBA}.bak.$(date +%Y%m%d%H%M%S)" 2>/dev/null || true

        # Replace peer authentication with md5 for local connections
        sed -i "s/local   all             all                                     peer/local   all             all                                     md5/" "$PG_HBA"
        sed -i "s/host    all             all             127.0.0.1\/32            ident/host    all             all             127.0.0.1\/32            md5/" "$PG_HBA" 2>/dev/null || true

        systemctl restart postgresql
        success "PostgreSQL configured: user=$DB_USER, db=$DB_NAME"
    else
        warn "Could not find pg_hba.conf, manual configuration may be needed"
    fi
}

# ============================================================
# Configure Redis
# ============================================================
setup_redis() {
    info "Configuring Redis..."

    # Ensure Redis is running
    systemctl start redis-server
    systemctl enable redis-server >/dev/null 2>&1

    if [ -n "$REDIS_PASS" ]; then
        # Set Redis password
        REDIS_CONF="/etc/redis/redis.conf"

        if [ -f "$REDIS_CONF" ]; then
            # Backup
            cp "$REDIS_CONF" "${REDIS_CONF}.bak.$(date +%Y%m%d%H%M%S)"

            # Set password
            if grep -q "^requirepass" "$REDIS_CONF"; then
                sed -i "s/^requirepass .*/requirepass $REDIS_PASS/" "$REDIS_CONF"
            else
                echo "requirepass $REDIS_PASS" >> "$REDIS_CONF"
            fi

            systemctl restart redis-server
            success "Redis configured with password"
        else
            warn "Redis config not found at $REDIS_CONF"
        fi
    else
        success "Redis configured (no password)"
    fi
}

# ============================================================
# Create system user and directories
# ============================================================
setup_user_and_dirs() {
    info "Setting up sub2api user and directories..."

    # Create system user
    if id "$SERVICE_USER" &>/dev/null; then
        info "User '$SERVICE_USER' already exists"
    else
        useradd -r -s /bin/sh -d "$INSTALL_DIR" "$SERVICE_USER"
        success "Created system user: $SERVICE_USER"
    fi

    # Create directories
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$INSTALL_DIR/data"
    mkdir -p "$INSTALL_DIR/data/logs"
    mkdir -p "$INSTALL_DIR/data/public"
    mkdir -p "$CONFIG_DIR"

    # Set ownership
    chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"
    chown -R "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR"

    success "Directories created: $INSTALL_DIR, $CONFIG_DIR"
}

# ============================================================
# Configure firewall
# ============================================================
setup_firewall() {
    info "Configuring firewall (UFW)..."

    # Allow SSH
    ufw allow OpenSSH >/dev/null 2>&1 || true

    # Allow HTTP and HTTPS
    ufw allow 80/tcp >/dev/null 2>&1 || true
    ufw allow 443/tcp >/dev/null 2>&1 || true

    # Enable UFW (non-interactive)
    if ! ufw status | grep -q "Status: active"; then
        echo "y" | ufw enable >/dev/null 2>&1 || true
        success "Firewall enabled (ports 80, 443, SSH allowed)"
    else
        success "Firewall already active"
    fi
}

# ============================================================
# Save configuration for deploy script
# ============================================================
save_config() {
    info "Saving configuration..."

    cat > "$CONFIG_DIR/.deploy-env" << EOF
# Auto-generated by setup-ubuntu.sh
# Do not edit manually - use deploy-ubuntu.sh to reconfigure
DB_NAME="$DB_NAME"
DB_USER="$DB_USER"
DB_PASS="$DB_PASS"
DB_HOST="localhost"
DB_PORT="5432"
REDIS_HOST="localhost"
REDIS_PORT="6379"
REDIS_PASS="$REDIS_PASS"
SERVER_PORT="$SERVER_PORT"
INSTALL_DIR="$INSTALL_DIR"
CONFIG_DIR="$CONFIG_DIR"
EOF

    chmod 600 "$CONFIG_DIR/.deploy-env"
    chown "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR/.deploy-env"

    success "Configuration saved to $CONFIG_DIR/.deploy-env"
}

# ============================================================
# Print completion
# ============================================================
print_completion() {
    echo ""
    echo "=============================================="
    success "Server Setup Complete!"
    echo "=============================================="
    echo ""
    echo "  Installed services:"
    echo "    - PostgreSQL  (database: $DB_NAME, user: $DB_USER)"
    echo "    - Redis       (${REDIS_PASS:+password protected}${REDIS_PASS:-no password})"
    echo "    - Nginx       (reverse proxy, will be configured by deploy script)"
    echo "    - UFW Firewall (ports 80, 443, SSH)"
    echo ""
    echo "  Directories:"
    echo "    Install:  $INSTALL_DIR"
    echo "    Config:   $CONFIG_DIR"
    echo ""
    echo "  Configuration saved: $CONFIG_DIR/.deploy-env"
    echo ""
    echo "=============================================="
    echo "  Next Steps"
    echo "=============================================="
    echo ""
    echo "  1. Build the application on your dev machine:"
    echo "     ./deploy/build.sh"
    echo ""
    echo "  2. Upload binary to server and deploy:"
    echo "     scp bin/sub2api user@server:/tmp/"
    echo "     scp deploy/deploy-ubuntu.sh user@server:/tmp/"
    echo "     ssh user@server 'sudo bash /tmp/deploy-ubuntu.sh'"
    echo ""
    echo "  Or build directly on the server (requires Go + Node.js)."
    echo "=============================================="
}

# ============================================================
# Main
# ============================================================
main() {
    echo ""
    echo "=============================================="
    echo "  Sub2API Ubuntu Server Setup"
    echo "=============================================="
    echo ""

    check_root
    detect_os
    configure
    install_packages
    setup_postgresql
    setup_redis
    setup_user_and_dirs
    setup_firewall
    save_config
    print_completion
}

main "$@"
