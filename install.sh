#!/bin/bash
#
# AICO Installation Script
# Installs the latest pre-built binary from GitHub Releases
#
set -euo pipefail

# ==============================================================================
# Global Variables
# ==============================================================================
OS=""           # Operating system (darwin/linux)
ARCH=""         # Architecture (amd64/arm64)
VERSION=""      # Release version tag (e.g., v1.2.3)
TEMP_DIR=""     # Temporary directory for downloaded files

# ==============================================================================
# Cleanup Handler
# ==============================================================================

# cleanup removes temporary files on script exit
cleanup() {
  if [[ -n "${TEMP_DIR}" && -d "${TEMP_DIR}" ]]; then
    rm -rf "${TEMP_DIR}"
  fi
}

# Register cleanup function to be called on exit (success or failure)
trap 'cleanup' EXIT

# ==============================================================================
# Logging Functions
# ==============================================================================

# log_info prints an informational message to stdout
log_info() {
  echo "[INFO] $1"
}

# log_error prints an error message to stderr
log_error() {
  echo "[ERROR] $1" >&2
}

# log_success prints a success message to stdout
log_success() {
  echo "[SUCCESS] $1"
}

# ==============================================================================
# Platform Detection
# ==============================================================================

# detect_platform detects the operating system and architecture
# Sets global variables OS and ARCH
# Returns 0 on success, 1 on unsupported platform
detect_platform() {
  local os_raw
  local arch_raw

  # Detect OS
  os_raw=$(uname -s)
  case "$os_raw" in
    Darwin)
      OS="darwin"
      ;;
    Linux)
      OS="linux"
      ;;
    *)
      log_error "Unsupported operating system: $os_raw"
      log_error "Supported platforms:"
      log_error "  - darwin (macOS)"
      log_error "  - linux"
      return 1
      ;;
  esac

  # Detect architecture
  arch_raw=$(uname -m)
  case "$arch_raw" in
    x86_64)
      ARCH="amd64"
      ;;
    aarch64|arm64)
      ARCH="arm64"
      ;;
    *)
      log_error "Unsupported architecture: $arch_raw"
      log_error "Supported architectures:"
      log_error "  - x86_64 (amd64)"
      log_error "  - arm64/aarch64"
      log_error ""
      log_error "Supported combinations:"
      log_error "  - darwin/amd64 (macOS Intel)"
      log_error "  - darwin/arm64 (macOS Apple Silicon)"
      log_error "  - linux/amd64"
      log_error "  - linux/arm64"
      return 1
      ;;
  esac

  return 0
}

# ==============================================================================
# GitHub Release Version Fetching
# ==============================================================================

# fetch_latest_version fetches the latest release version from GitHub API
# Sets global variable VERSION
# Returns 0 on success, 1 on API error
fetch_latest_version() {
  local api_url="https://api.github.com/repos/micheam/ai-assistant-console/releases/latest"
  local response
  local http_code

  log_info "Fetching latest release information..."

  # Fetch release information from GitHub API with HTTP status code
  response=$(curl -fsSL --max-time 10 -w "\n%{http_code}" "$api_url" 2>&1) || {
    local curl_exit=$?
    # Try to extract HTTP code if available
    http_code=$(echo "$response" | tail -n 1 2>/dev/null)

    if [[ "$http_code" == "403" ]]; then
      log_error "GitHub API rate limit exceeded (HTTP 403)"
      log_error "Please wait a few minutes and try again, or set GITHUB_TOKEN environment variable"
    else
      log_error "Failed to fetch release information from GitHub API"
      if [[ -n "$http_code" && "$http_code" =~ ^[0-9]+$ ]]; then
        log_error "HTTP status code: $http_code"
      fi
      log_error "Please check your network connection and try again"
    fi
    return 1
  }

  # Extract HTTP status code (last line)
  http_code=$(echo "$response" | tail -n 1)
  # Remove HTTP status code from response
  response=$(echo "$response" | sed '$d')

  # Check if we got a valid response
  if [[ -z "$response" ]]; then
    log_error "Empty response from GitHub API"
    return 1
  fi

  # Extract tag_name from JSON response
  if command -v jq > /dev/null 2>&1; then
    # Use jq if available
    VERSION=$(echo "$response" | jq -r .tag_name)
  else
    # Fallback to grep and sed
    VERSION=$(echo "$response" | grep -o '"tag_name"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
  fi

  # Validate version was extracted
  if [[ -z "$VERSION" ]]; then
    log_error "Failed to extract version from API response"
    return 1
  fi

  log_info "Latest version: $VERSION"
  return 0
}

# ==============================================================================
# Secure File Download
# ==============================================================================

# download_file downloads a file from a URL to a local path
# Arguments:
#   $1 - URL to download from (must be HTTPS)
#   $2 - Output file path
# Returns 0 on success, 1 on download error
download_file() {
  local url="$1"
  local output="$2"

  # Validate HTTPS URL
  if [[ ! "$url" =~ ^https:// ]]; then
    log_error "Only HTTPS URLs are allowed for security reasons"
    log_error "Provided URL: $url"
    return 1
  fi

  log_info "Downloading $url..."

  # Download file with curl, capturing HTTP status code
  # -f: Fail silently on HTTP errors
  # -s: Silent mode (no progress bar)
  # -S: Show errors even in silent mode
  # -L: Follow redirects
  # -o: Output file
  # -w: Write out HTTP status code
  local http_code
  http_code=$(curl -fsSL -w "%{http_code}" -o "$output" "$url" 2>&1 | tail -n 1)
  local curl_exit=$?

  if [[ $curl_exit -eq 0 ]]; then
    log_info "Downloaded successfully"
    return 0
  else
    log_error "Failed to download file from $url"
    if [[ -n "$http_code" && "$http_code" =~ ^[0-9]+$ ]]; then
      log_error "HTTP status code: $http_code"
    fi
    log_error "Please check the URL and your network connection"
    return 1
  fi
}

# download_release_files downloads binary and checksum files from GitHub Releases
# Uses global variables: VERSION, OS, ARCH, TEMP_DIR
# Returns 0 on success, 1 on download error
download_release_files() {
  local base_url="https://github.com/micheam/ai-assistant-console/releases/download"
  local binary_filename="aico_${OS}_${ARCH}.tar.gz"
  local checksum_filename="${binary_filename}.sha256"
  local binary_url="${base_url}/${VERSION}/${binary_filename}"
  local checksum_url="${base_url}/${VERSION}/${checksum_filename}"

  log_info "Downloading release files for version ${VERSION}..."
  log_info "Platform: ${OS}/${ARCH}"

  # Download binary file
  if ! download_file "$binary_url" "${TEMP_DIR}/${binary_filename}"; then
    log_error "Failed to download binary file"
    return 1
  fi

  # Download checksum file
  if ! download_file "$checksum_url" "${TEMP_DIR}/${checksum_filename}"; then
    log_error "Failed to download checksum file"
    return 1
  fi

  log_success "Downloaded binary and checksum files successfully"
  return 0
}

# ==============================================================================
# Checksum Verification
# ==============================================================================

# verify_checksum verifies the SHA256 checksum of a binary file
# Arguments:
#   $1 - Binary file path
#   $2 - Checksum file path
# Returns 0 on success, 1 on mismatch or error, 2 on checksum command not found
verify_checksum() {
  local binary_file="$1"
  local checksum_file="$2"

  log_info "Verifying checksum..."

  # Check if checksum file exists
  if [[ ! -f "$checksum_file" ]]; then
    log_error "Checksum file not found: $checksum_file"
    log_error "Cannot verify binary integrity"
    return 1
  fi

  # Determine which checksum command to use
  local checksum_cmd=""
  if command -v shasum > /dev/null 2>&1; then
    checksum_cmd="shasum"
  elif command -v sha256sum > /dev/null 2>&1; then
    checksum_cmd="sha256sum"
  else
    log_error "WARNING: Neither 'shasum' nor 'sha256sum' command found"
    log_error "Cannot verify checksum"
    log_error ""
    # Ask user if they want to continue without verification
    read -p "Continue without verification? (y/N): " response
    if [[ "$response" =~ ^[Yy]$ ]]; then
      log_info "Skipping checksum verification (not recommended)"
      return 2
    else
      log_error "Installation aborted"
      return 1
    fi
  fi

  # Change to the directory containing the files for checksum verification
  local binary_dir
  binary_dir=$(dirname "$binary_file")
  local binary_name
  binary_name=$(basename "$binary_file")
  local checksum_name
  checksum_name=$(basename "$checksum_file")

  # Verify checksum
  cd "$binary_dir" || return 1

  if [[ "$checksum_cmd" == "shasum" ]]; then
    if shasum -a 256 -c "$checksum_name" > /dev/null 2>&1; then
      log_success "Checksum verification passed"
      cd - > /dev/null || return 1
      return 0
    else
      log_error "Checksum verification FAILED"
      log_error "The downloaded binary may be corrupted or tampered with"
      log_error "Please try downloading again or verify the source"
      cd - > /dev/null || return 1
      return 1
    fi
  else
    if sha256sum -c "$checksum_name" > /dev/null 2>&1; then
      log_success "Checksum verification passed"
      cd - > /dev/null || return 1
      return 0
    else
      log_error "Checksum verification FAILED"
      log_error "The downloaded binary may be corrupted or tampered with"
      log_error "Please try downloading again or verify the source"
      cd - > /dev/null || return 1
      return 1
    fi
  fi
}

# ==============================================================================
# Archive Extraction
# ==============================================================================

# extract_archive extracts a tar.gz archive to a directory
# Arguments:
#   $1 - Archive file path
#   $2 - Extract directory path
# Returns 0 on success, 1 on extraction error
extract_archive() {
  local archive_file="$1"
  local extract_dir="$2"

  log_info "Extracting archive..."

  # Check archive file exists
  if [[ ! -f "$archive_file" ]]; then
    log_error "Archive file not found: $archive_file"
    return 1
  fi

  # Check extract directory exists
  if [[ ! -d "$extract_dir" ]]; then
    log_error "Extract directory not found: $extract_dir"
    return 1
  fi

  # Check archive contents
  local file_count
  file_count=$(tar -tzf "$archive_file" 2>/dev/null | wc -l)

  if [[ $file_count -eq 0 ]]; then
    log_error "Archive appears to be empty or corrupted"
    return 1
  elif [[ $file_count -gt 1 ]]; then
    log_info "Warning: Archive contains multiple files ($file_count files)"
    log_info "Expected a single binary file"
  fi

  # Extract archive
  if tar -xzf "$archive_file" -C "$extract_dir"; then
    log_success "Archive extracted successfully"
    return 0
  else
    log_error "Failed to extract archive"
    log_error "The archive may be corrupted or in an unsupported format"
    return 1
  fi
}

# ==============================================================================
# Binary Installation
# ==============================================================================

# install_binary installs the binary to $HOME/.local/bin
# Arguments:
#   $1 - Source binary file path
# Returns 0 on success, 1 on installation error
install_binary() {
  local source_binary="$1"
  local install_dir="$HOME/.local/bin"
  local install_path="$install_dir/aico"

  log_info "Installing binary..."

  # Check source binary exists
  if [[ ! -f "$source_binary" ]]; then
    log_error "Source binary not found: $source_binary"
    return 1
  fi

  # Create install directory if it doesn't exist
  if [[ ! -d "$install_dir" ]]; then
    log_info "Creating installation directory: $install_dir"
    if ! mkdir -p "$install_dir"; then
      log_error "Failed to create installation directory"
      return 1
    fi
  fi

  # Check if binary already exists and show current version
  if [[ -f "$install_path" ]]; then
    log_info "Existing installation found"
    local current_version
    current_version=$("$install_path" --version 2>/dev/null || echo "unknown")
    log_info "Current version: $current_version"
    log_info "Updating to new version..."
  fi

  # Copy binary to installation directory
  if ! cp "$source_binary" "$install_path"; then
    log_error "Failed to copy binary to $install_path"
    log_error "Please check disk space and permissions"
    return 1
  fi

  # Set executable permissions
  if ! chmod +x "$install_path"; then
    log_error "Failed to set executable permissions on $install_path"
    return 1
  fi

  log_success "Binary installed successfully to $install_path"
  return 0
}

# ==============================================================================
# Main Installation Flow
# ==============================================================================

# main executes the complete installation workflow
# Returns 0 on success, 1 on error
main() {
  log_info "Installing aico..."
  echo ""

  # Create temporary directory for downloads
  TEMP_DIR=$(mktemp -d -t aico-install.XXXXXX) || {
    log_error "Failed to create temporary directory"
    return 1
  }

  # Step 1: Detect platform
  if ! detect_platform; then
    log_error "Platform detection failed"
    return 1
  fi

  # Step 2: Fetch latest version
  if ! fetch_latest_version; then
    log_error "Failed to fetch latest version"
    return 1
  fi

  # Step 3: Download release files
  if ! download_release_files; then
    log_error "Failed to download release files"
    return 1
  fi

  # Step 4: Verify checksum
  local binary_filename="aico_${OS}_${ARCH}.tar.gz"
  local checksum_filename="${binary_filename}.sha256"

  verify_checksum "${TEMP_DIR}/${binary_filename}" "${TEMP_DIR}/${checksum_filename}"
  local checksum_result=$?

  if [[ $checksum_result -eq 2 ]]; then
    # User chose to skip verification
    log_info "Continuing without checksum verification (user choice)"
  elif [[ $checksum_result -ne 0 ]]; then
    # Checksum verification failed
    log_error "Checksum verification failed"
    return 1
  fi

  # Step 5: Extract archive
  if ! extract_archive "${TEMP_DIR}/${binary_filename}" "${TEMP_DIR}"; then
    log_error "Failed to extract archive"
    return 1
  fi

  # Step 6: Install binary
  local install_path="$HOME/.local/bin/aico"
  if ! install_binary "${TEMP_DIR}/aico"; then
    log_error "Failed to install binary"
    return 1
  fi

  # Installation complete - show success message with version and path
  echo ""
  log_success "Successfully installed aico ${VERSION} to ${install_path}"
  echo ""

  # Check if $HOME/.local/bin is in PATH
  if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    log_info "Note: $HOME/.local/bin is not in your PATH"
    log_info "Add the following line to your ~/.bashrc or ~/.zshrc:"
    echo ""
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
  fi

  return 0
}

# Execute main function if script is run directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main
  exit $?
fi
