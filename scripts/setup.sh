#!/usr/bin/env bash
set -euo pipefail

# ──────────────────────────────────────────────────────────────
# cloud-ide-mount — Linux/macOS Setup Script
# ──────────────────────────────────────────────────────────────

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SSH_DIR="$HOME/.ssh"
SSH_KEY="$SSH_DIR/codespaces.auto"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║  cloud-ide-mount — Linux/macOS Setup     ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════╝${NC}"
echo ""

# ─── Helpers ──────────────────────────────────────────────────────
check_cmd() {
    local name="$1" required="${2:-true}"
    if command -v "$name" &>/dev/null; then
        echo -e "  ${GREEN}✅ $name found: $(command -v "$name")${NC}"
        return 0
    fi
    if [ "$required" = "true" ]; then
        echo -e "  ${RED}❌ $name NOT found.${NC}"
        return 1
    fi
    echo -e "  ${YELLOW}⚠️  $name NOT found (optional).${NC}"
    return 2
}

check_version() {
    local name="$1" cmd="$2" min_ver="$3"
    local ver
    ver=$($cmd 2>&1 | head -1 | grep -oE '[0-9]+\.[0-9]+' | head -1 || echo "0")
    if [ "$(echo "$ver >= $min_ver" | bc 2>/dev/null || echo 0)" = "1" ] || [ "$ver" = "0" ]; then
        if [ "$ver" != "0" ]; then
            echo -e "  ${GREEN}✅ $name $ver (≥ $min_ver)${NC}"
        else
            echo -e "  ${YELLOW}⚠️  Cannot detect $name version${NC}"
        fi
        return 0
    fi
    echo -e "  ${YELLOW}⚠️  $name $ver (< $min_ver, upgrade recommended)${NC}"
    return 0
}

install_pkg() {
    local name="$1" install_cmd="$2"
    echo -e "  ${YELLOW}Installing $name...${NC}"
    if eval "$install_cmd"; then
        echo -e "  ${GREEN}✅ $name installed.${NC}"
        return 0
    fi
    echo -e "  ${RED}❌ Failed to install $name. Install manually.${NC}"
    return 1
}

detect_pkg_manager() {
    if command -v brew &>/dev/null; then
        echo "brew"
    elif command -v apt-get &>/dev/null; then
        echo "apt"
    elif command -v dnf &>/dev/null; then
        echo "dnf"
    elif command -v yum &>/dev/null; then
        echo "yum"
    elif command -v pacman &>/dev/null; then
        echo "pacman"
    else
        echo "unknown"
    fi
}

PKG_MANAGER=$(detect_pkg_manager)
echo -e "  ${MAGENTA}Detected package manager: $PKG_MANAGER${NC}"
echo ""

# ─── Prerequisites ────────────────────────────────────────────────
echo -e "  ${MAGENTA}─── Prerequisites ───${NC}"
ALL_GOOD=true

# Git
if ! check_cmd "git"; then ALL_GOOD=false; fi

# Go
if check_cmd "go"; then
    check_version "go" "go version" "1.21"
else
    echo -e "  ${YELLOW}  Install Go: https://go.dev/dl/ ${NC}"
    ALL_GOOD=false
fi

# gh (GitHub CLI)
if ! check_cmd "gh"; then
    case "$PKG_MANAGER" in
        brew) install_pkg "gh" "brew install gh" || ALL_GOOD=false ;;
        apt) install_pkg "gh" '(type -p wget >/dev/null || sudo apt-get install wget -y) \
              && sudo mkdir -p -m 755 /etc/apt/keyrings \
              && wget -qO- https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null \
              && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
              && sudo apt-get update && sudo apt-get install gh -y' || ALL_GOOD=false ;;
        dnf|yum) install_pkg "gh" "sudo dnf install gh -y" || ALL_GOOD=false ;;
        pacman) install_pkg "gh" "sudo pacman -S github-cli --noconfirm" || ALL_GOOD=false ;;
        *) echo -e "  ${YELLOW}  Install gh manually: https://cli.github.com/ ${NC}"; ALL_GOOD=false ;;
    esac
fi

# rclone
if ! check_cmd "rclone"; then
    case "$PKG_MANAGER" in
        brew) install_pkg "rclone" "brew install rclone" || ALL_GOOD=false ;;
        apt) install_pkg "rclone" "sudo apt-get install rclone -y" || ALL_GOOD=false ;;
        dnf|yum) install_pkg "rclone" "sudo dnf install rclone -y" || ALL_GOOD=false ;;
        pacman) install_pkg "rclone" "sudo pacman -S rclone --noconfirm" || ALL_GOOD=false ;;
        *) echo -e "  ${YELLOW}  Install rclone manually: https://rclone.org/downloads/ ${NC}"; ALL_GOOD=false ;;
    esac
fi

if [ "$ALL_GOOD" = false ]; then
    echo ""
    echo -e "  ${RED}❌ Some required dependencies are missing. Fix above and re-run.${NC}"
    exit 1
fi

# ─── SSH Key ──────────────────────────────────────────────────────
echo ""
echo -e "  ${MAGENTA}─── SSH Key ───${NC}"
if [ ! -f "$SSH_KEY" ]; then
    mkdir -p "$SSH_DIR"
    ssh-keygen -t rsa -b 4096 -f "$SSH_KEY" -N "" -C "codespaces-auto" 2>&1
    echo -e "  ${GREEN}✅ SSH key generated: $SSH_KEY${NC}"
else
    echo -e "  ${GREEN}✅ SSH key exists: $SSH_KEY${NC}"
fi

# ─── gh auth ──────────────────────────────────────────────────────
echo ""
echo -e "  ${MAGENTA}─── GitHub CLI Authentication ───${NC}"
if gh auth status &>/dev/null; then
    echo -e "  ${GREEN}✅ GitHub CLI authenticated.${NC}"
else
    echo -e "  ${YELLOW}  You need to log in to GitHub CLI.${NC}"
    echo -e "  ${YELLOW}  Run: gh auth login${NC}"
    echo -e "  ${YELLOW}  Then re-run this script.${NC}"
fi

# ─── Build ────────────────────────────────────────────────────────
echo ""
echo -e "  ${MAGENTA}─── Build ───${NC}"
cd "$REPO_ROOT"
if go build -o cloud-ide-mount . 2>&1; then
    echo -e "  ${GREEN}✅ Build successful: $REPO_ROOT/cloud-ide-mount${NC}"
else
    echo -e "  ${RED}❌ Build failed.${NC}"
    ALL_GOOD=false
fi

# ─── Tests ────────────────────────────────────────────────────────
echo ""
echo -e "  ${MAGENTA}─── Tests ───${NC}"
if go test -race -count=1 ./... 2>&1; then
    echo -e "  ${GREEN}✅ All tests pass.${NC}"
else
    echo -e "  ${RED}❌ Some tests failed.${NC}"
fi

# ─── Summary ─────────────────────────────────────────────────────
echo ""
echo -e "${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║  Setup complete!                         ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════╝${NC}"
echo ""
echo -e "  ${MAGENTA}Next steps:${NC}"
echo "    cs-mount list         # List codespaces"
echo "    cs-mount mount        # Mount a codespace"
echo "    cs-mount --help       # See all commands"
echo ""
