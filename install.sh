#!/usr/bin/env bash
set -euo pipefail

APP=opencode-piko
REPO=friddle/opencode-piko-remote
INSTALL_DIR=${INSTALL_DIR:-$HOME/.${APP}/bin}

MUTED='\033[0;2m'
RED='\033[0;31m'
ORANGE='\033[38;5;214m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

usage() {
    cat <<EOF
${APP} Installer

Usage: install.sh [options]

Options:
    -h, --help              Display this help message
    -v, --version <version> Install a specific version (e.g. 0.1.0)
        --no-modify-path    Don't modify shell config files (.zshrc, .bashrc, etc.)

Examples:
    curl -fsSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash
    curl -fsSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash -s -- --version 0.1.0
EOF
}

requested_version=""
no_modify_path=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--version)
            if [[ -n "${2:-}" ]]; then
                requested_version="$2"
                shift 2
            else
                echo -e "${RED}Error: --version requires a version argument${NC}"
                exit 1
            fi
            ;;
        --no-modify-path)
            no_modify_path=true
            shift
            ;;
        *)
            echo -e "${ORANGE}Warning: Unknown option '$1'${NC}" >&2
            shift
            ;;
    esac
done

print_message() {
    local level=$1
    local message=$2
    local color=""
    case $level in
        info) color="${NC}" ;;
        warning) color="${ORANGE}" ;;
        error) color="${RED}" ;;
    esac
    echo -e "${color}${message}${NC}"
}

# ---- OS / Arch detection ----
raw_os=$(uname -s)
case "$raw_os" in
    Darwin*) os="darwin" ;;
    Linux*)  os="linux" ;;
    MINGW*|MSYS*|CYGWIN*)
        echo -e "${RED}Error: Windows is not supported. Use WSL instead.${NC}"
        exit 1
        ;;
    *)
        echo -e "${RED}Error: Unsupported OS: $raw_os${NC}"
        exit 1
        ;;
esac

arch=$(uname -m)
case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $arch${NC}"
        echo -e "${MUTED}Supported: amd64 (x86_64), arm64 (aarch64)${NC}"
        exit 1
        ;;
esac

# On Apple Silicon, prefer arm64 even under Rosetta
if [ "$os" = "darwin" ] && [ "$arch" = "amd64" ]; then
    rosetta_flag=$(sysctl -n sysctl.proc_translated 2>/dev/null || echo 0)
    if [ "$rosetta_flag" = "1" ]; then
        arch="arm64"
    fi
fi

combo="$os-$arch"
case "$combo" in
    linux-amd64|linux-arm64|darwin-amd64|darwin-arm64) ;;
    *)
        echo -e "${RED}Error: Unsupported OS/Arch: $os/$arch${NC}"
        exit 1
        ;;
esac

# ---- Dependencies ----
if ! command -v curl >/dev/null 2>&1; then
    echo -e "${RED}Error: 'curl' is required but not installed.${NC}"
    exit 1
fi

# ---- Resolve download URL + version ----
filename="${APP}-${os}-${arch}"

if [ -z "$requested_version" ]; then
    url="https://github.com/${REPO}/releases/latest/download/${filename}"
    specific_version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | sed -n 's/.*"tag_name": *"v\([^"]*\)".*/\1/p' | head -1)
    if [ -z "$specific_version" ]; then
        echo -e "${RED}Error: Failed to fetch latest version information${NC}"
        exit 1
    fi
else
    requested_version="${requested_version#v}"
    url="https://github.com/${REPO}/releases/download/v${requested_version}/${filename}"
    specific_version="$requested_version"
    http_status=$(curl -sI -o /dev/null -w "%{http_code}" \
        "https://github.com/${REPO}/releases/tag/v${requested_version}")
    if [ "$http_status" = "404" ]; then
        echo -e "${RED}Error: Release v${requested_version} not found${NC}"
        echo -e "${MUTED}Available releases: https://github.com/${REPO}/releases${NC}"
        exit 1
    fi
fi

# ---- Check if already installed ----
check_version() {
    if command -v "$APP" >/dev/null 2>&1; then
        installed_version=$("$APP" --version 2>/dev/null | head -1 || echo "")
        if [[ "$installed_version" == *"$specific_version"* ]]; then
            print_message info "${MUTED}Version ${NC}${specific_version}${MUTED} already installed${NC}"
            exit 0
        fi
        if [ -n "$installed_version" ]; then
            print_message info "${MUTED}Installed version: ${NC}${installed_version}"
        fi
    fi
}

# ---- Download with progress ----
download_with_progress() {
    if [ -t 2 ]; then
        exec 4>&2
    else
        exec 4>/dev/null
    fi
    printf "${MUTED}Downloading ${NC}${filename}${MUTED} ...${NC}\n" >&4
    curl -# -fSL -o "$1" "$url"
    local ret=$?
    exec 4>&-
    return $ret
}

download_and_install() {
    print_message info "\n${MUTED}Installing ${NC}${APP} ${MUTED}version: ${NC}${specific_version} ${MUTED}(${os}/${arch})${NC}"
    mkdir -p "$INSTALL_DIR"
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT

    if ! download_with_progress "$tmp_dir/$APP"; then
        echo ""
        echo -e "${RED}Error: Download failed.${NC}"
        echo -e "${MUTED}URL: $url${NC}"
        echo -e "${MUTED}The binary for ${os}/${arch} may not exist in this release.${NC}"
        rm -rf "$tmp_dir"
        exit 1
    fi

    mv "$tmp_dir/$APP" "$INSTALL_DIR/$APP"
    chmod 755 "$INSTALL_DIR/$APP"
}

check_version
download_and_install

# ---- PATH setup ----
add_to_path() {
    local config_file=$1
    local command=$2

    if grep -Fxq "$command" "$config_file"; then
        print_message info "${MUTED}Command already exists in ${NC}$config_file${MUTED}, skipping.${NC}"
    elif [[ -w $config_file ]]; then
        {
            echo ""
            echo "# opencode-piko"
            echo "$command"
        } >> "$config_file"
        print_message info "${MUTED}Added ${NC}${APP}${MUTED} to \$PATH in ${NC}${config_file}"
    else
        print_message warning "Manually add the directory to $config_file (or similar):"
        print_message info "  $command"
    fi
}

XDG_CONFIG_HOME=${XDG_CONFIG_HOME:-$HOME/.config}

current_shell=$(basename "$SHELL")
case $current_shell in
    fish)
        config_files="$HOME/.config/fish/config.fish"
        ;;
    zsh)
        config_files="${ZDOTDIR:-$HOME}/.zshrc ${ZDOTDIR:-$HOME}/.zshenv $XDG_CONFIG_HOME/zsh/.zshrc $XDG_CONFIG_HOME/zsh/.zshenv"
        ;;
    bash)
        config_files="$HOME/.bashrc $HOME/.bash_profile $HOME/.profile $XDG_CONFIG_HOME/bash/.bashrc $XDG_CONFIG_HOME/bash/.bash_profile"
        ;;
    ash|sh)
        config_files="$HOME/.ashrc $HOME/.profile /etc/profile"
        ;;
    *)
        config_files="$HOME/.bashrc $HOME/.bash_profile $XDG_CONFIG_HOME/bash/.bashrc $XDG_CONFIG_HOME/bash/.bash_profile"
        ;;
esac

if [[ "$no_modify_path" != "true" ]]; then
    config_file=""
    for file in $config_files; do
        if [[ -f $file ]]; then
            config_file=$file
            break
        fi
    done

    if [[ -z $config_file ]]; then
        print_message warning "No config file found for $current_shell. Manually add to PATH:"
        print_message info "  export PATH=$INSTALL_DIR:\$PATH"
    elif [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        case $current_shell in
            fish)
                add_to_path "$config_file" "fish_add_path $INSTALL_DIR"
                ;;
            *)
                add_to_path "$config_file" "export PATH=$INSTALL_DIR:\$PATH"
                ;;
        esac
    fi
fi

if [ -n "${GITHUB_ACTIONS-}" ] && [ "${GITHUB_ACTIONS}" == "true" ]; then
    echo "$INSTALL_DIR" >> "$GITHUB_PATH"
    print_message info "${MUTED}Added ${NC}$INSTALL_DIR${MUTED} to \$GITHUB_PATH${NC}"
fi

# ---- Done ----
echo ""
echo -e "${GREEN}${APP} v${specific_version} installed successfully!${NC}"
echo ""
echo -e "${MUTED}Binary location:${NC} ${INSTALL_DIR}/${APP}"
echo ""
echo -e "${MUTED}Restart your shell or run:${NC}"
echo -e "  export PATH=$INSTALL_DIR:\$PATH"
echo ""
echo -e "${MUTED}Then start with:${NC}"
echo ""
echo -e "  ${APP} /path/to/project \\"
echo -e "    --name=my-dev \\"
echo -e "    --remote=piko-server.example.com:8088 \\"
echo -e "    --pass=your-password"
echo ""
echo -e "${MUTED}For more info:${NC} https://github.com/${REPO}"
echo ""
