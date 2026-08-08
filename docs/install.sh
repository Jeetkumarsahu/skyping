#!/usr/bin/env sh
set -e

REPO="Jeetkumarsahu/skyping"
BINARY_NAME="skyping"
INSTALL_DIR="/usr/local/bin"
SERVICE_NAME="skyping"

BOLD="\033[1m"
GREEN="\033[32m"
CYAN="\033[36m"
RED="\033[31m"
RESET="\033[0m"

info()    { printf "${CYAN}==>${RESET} ${BOLD}%s${RESET}\n" "$1"; }
success() { printf "${GREEN}✓${RESET} %s\n" "$1"; }
error()   { printf "${RED}✗ Error:${RESET} %s\n" "$1" >&2; exit 1; }

detect_os() {
  OS="$(uname -s)"
  case "$OS" in
    Linux) OS="linux" ;;
    *) error "Unsupported OS: $OS. This installer is for Linux only." ;;
  esac
}

detect_arch() {
  ARCH="$(uname -m)"
  case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    armv7l)  ARCH="arm" ;;
    *)       error "Unsupported architecture: $ARCH" ;;
  esac
}

check_deps() {
  for cmd in curl tar; do
    command -v "$cmd" >/dev/null 2>&1 || error "'$cmd' is required but not installed."
  done
}

install_cloudflared() {
  if command -v cloudflared >/dev/null 2>&1; then
    success "cloudflared already installed"
    return
  fi

  CLOUDFLARED_URL="https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-${OS}-${ARCH}"
  TMP_DIR="$(mktemp -d)"

  info "Installing cloudflared tunnel helper (${OS}/${ARCH})..."
  curl -fsSL "$CLOUDFLARED_URL" -o "${TMP_DIR}/cloudflared" \
    || error "cloudflared download failed. URL: $CLOUDFLARED_URL"

  chmod +x "${TMP_DIR}/cloudflared"
  mv "${TMP_DIR}/cloudflared" "${INSTALL_DIR}/cloudflared"
  rm -rf "$TMP_DIR"

  success "cloudflared installed to ${INSTALL_DIR}/cloudflared"
}

fetch_latest_version() {
  info "Fetching latest release..."
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
  [ -z "$VERSION" ] && error "Could not fetch latest version. Check your internet connection."
  success "Latest version: $VERSION"
}

install_binary() {
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${OS}-${ARCH}.tar.gz"
  TMP_DIR="$(mktemp -d)"

  info "Downloading ${BINARY_NAME} ${VERSION} (${OS}/${ARCH})..."
  curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/${BINARY_NAME}.tar.gz" \
    || error "Download failed. URL: $DOWNLOAD_URL"

  tar -xzf "${TMP_DIR}/${BINARY_NAME}.tar.gz" -C "$TMP_DIR"

  if [ -w "$INSTALL_DIR" ]; then
    mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
  else
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
      SHELL_RC="$HOME/.bashrc"
      [ -f "$HOME/.zshrc" ] && SHELL_RC="$HOME/.zshrc"
      echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_RC"
      info "Added $INSTALL_DIR to PATH in $SHELL_RC"
    fi
  fi

  rm -rf "$TMP_DIR"
  success "Binary installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

verify() {
  if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    success "skyping installed successfully"
  else
    info "Binary at: ${INSTALL_DIR}/${BINARY_NAME}"
  fi
}

printf "\n${BOLD}Skyping Linux Installer${RESET}\n\n"

check_deps
detect_os
detect_arch
fetch_latest_version
install_binary
install_cloudflared
verify

printf "\n${BOLD}Done!${RESET} Start sharing:\n\n"
printf "  ${CYAN}skyping agent${RESET}          # start agent and print a secure link\n\n"
