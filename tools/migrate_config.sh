#!/bin/bash

# Configuration Migration Script for Redis-Runner
# This script helps migrate legacy configuration files to the new unified format

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
MIGRATION_TOOL="./tools/config_migration"
BACKUP_DIR="./conf/backup"
VERBOSE=false
DRY_RUN=false
AUTO_CONFIRM=false

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

show_usage() {
    cat << EOF
Configuration Migration Script for Redis-Runner

USAGE:
    $0 [OPTIONS] <config_file>

OPTIONS:
    -h, --help          Show this help message
    -v, --verbose       Enable verbose output
    -d, --dry-run       Show what would be migrated without making changes
    -y, --yes           Auto-confirm all prompts
    -b, --backup-dir    Backup directory (default: ./conf/backup)
    -t, --tool          Path to migration tool (default: ./tools/config_migration)
    -f, --format        Output format: yaml or json (default: yaml)
    -p, --protocol      Protocol to migrate: redis, http, kafka, auto (default: auto)

EXAMPLES:
    # Migrate a single Redis configuration
    $0 conf/old-redis.yaml

    # Migrate with verbose output and custom backup directory
    $0 -v -b /tmp/backup conf/legacy.yaml

    # Dry run to see what would be changed
    $0 -d conf/old-config.json

    # Batch migrate all config files
    $0 -y conf/*.yaml

EOF
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check if migration tool exists
    if [ ! -f "$MIGRATION_TOOL" ]; then
        log_error "Migration tool not found: $MIGRATION_TOOL"
        log_info "Building migration tool..."
        
        if ! go build -o "$MIGRATION_TOOL" tools/config_migration.go; then
            log_error "Failed to build migration tool"
            exit 1
        fi
        
        log_success "Migration tool built successfully"
    fi
    
    # Check if backup directory exists
    if [ ! -d "$BACKUP_DIR" ]; then
        log_info "Creating backup directory: $BACKUP_DIR"
        mkdir -p "$BACKUP_DIR"
    fi
    
    log_success "Dependencies check completed"
}

detect_config_type() {
    local config_file="$1"
    
    # Check file extension
    case "${config_file##*.}" in
        yaml|yml)
            echo "yaml"
            return
            ;;
        json)
            echo "json"
            return
            ;;
    esac
    
    # Check file content
    if head -n 5 "$config_file" | grep -q "^[[:space:]]*{"; then
        echo "json"
    else
        echo "yaml"
    fi
}

analyze_config() {
    local config_file="$1"
    
    log_info "Analyzing configuration file: $config_file"
    
    # Detect format
    local format=$(detect_config_type "$config_file")
    log_info "Detected format: $format"
    
    # Check for legacy patterns
    local legacy_patterns=0
    
    # Check for old Redis format
    if grep -q "host:" "$config_file" && grep -q "port:" "$config_file"; then
        log_warn "Found legacy Redis host/port configuration"
        legacy_patterns=$((legacy_patterns + 1))
    fi
    
    # Check for old cluster flag
    if grep -q "cluster: true" "$config_file"; then
        log_warn "Found legacy Redis cluster configuration"
        legacy_patterns=$((legacy_patterns + 1))
    fi
    
    if [ $legacy_patterns -eq 0 ]; then
        log_info "Configuration appears to be in new format already"
        return 1
    else
        log_warn "Found $legacy_patterns legacy configuration patterns"
        return 0
    fi
}

migrate_config() {
    local config_file="$1"
    local output_format="${2:-yaml}"
    local protocol="${3:-auto}"
    
    log_info "Starting migration for: $config_file"
    
    # Analyze configuration
    if ! analyze_config "$config_file"; then
        if [ "$AUTO_CONFIRM" = false ]; then
            read -p "File appears to be in new format. Continue anyway? (y/N): " -r
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                log_info "Skipping migration for: $config_file"
                return 0
            fi
        fi
    fi
    
    # Build migration command
    local cmd="$MIGRATION_TOOL -input \"$config_file\" -format \"$output_format\""
    
    if [ "$protocol" != "auto" ]; then
        cmd="$cmd -protocol \"$protocol\""
    fi
    
    if [ "$VERBOSE" = true ]; then
        cmd="$cmd -verbose"
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN - Would execute: $cmd"
        return 0
    fi
    
    # Execute migration
    log_info "Executing migration..."
    if eval "$cmd"; then
        log_success "Migration completed for: $config_file"
        
        # Move original to backup
        local backup_file="$BACKUP_DIR/$(basename "$config_file").$(date +%Y%m%d_%H%M%S).backup"
        cp "$config_file" "$backup_file"
        log_info "Original file backed up to: $backup_file"
        
    else
        log_error "Migration failed for: $config_file"
        return 1
    fi
}

migrate_batch() {
    local config_files=("$@")
    local total=${#config_files[@]}
    local succeeded=0
    local failed=0
    
    log_info "Starting batch migration for $total files"
    
    for config_file in "${config_files[@]}"; do
        if [ ! -f "$config_file" ]; then
            log_warn "File not found: $config_file"
            failed=$((failed + 1))
            continue
        fi
        
        if migrate_config "$config_file"; then
            succeeded=$((succeeded + 1))
        else
            failed=$((failed + 1))
        fi
        
        echo ""
    done
    
    log_info "Batch migration summary:"
    log_success "  Succeeded: $succeeded"
    if [ $failed -gt 0 ]; then
        log_error "  Failed: $failed"
    fi
    log_info "  Total: $total"
}

show_migration_report() {
    echo ""
    echo "=== Migration Report ==="
    echo ""
    
    # List migrated files
    log_info "Migrated configuration files:"
    find conf -name "*.new.yaml" -o -name "*.new.json" | while read -r file; do
        echo "  - $file"
    done
    
    echo ""
    
    # List backup files
    log_info "Backup files created:"
    find "$BACKUP_DIR" -name "*.backup" | while read -r file; do
        echo "  - $file"
    done
    
    echo ""
    log_info "Next steps:"
    echo "  1. Review migrated files (*.new.yaml or *.new.json)"
    echo "  2. Test configurations with enhanced commands"
    echo "  3. Replace original files when satisfied"
    echo "  4. Clean up backup files when no longer needed"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -y|--yes)
            AUTO_CONFIRM=true
            shift
            ;;
        -b|--backup-dir)
            BACKUP_DIR="$2"
            shift 2
            ;;
        -t|--tool)
            MIGRATION_TOOL="$2"
            shift 2
            ;;
        -f|--format)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        -p|--protocol)
            PROTOCOL="$2"
            shift 2
            ;;
        -*)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
        *)
            CONFIG_FILES+=("$1")
            shift
            ;;
    esac
done

# Check if config files provided
if [ ${#CONFIG_FILES[@]} -eq 0 ]; then
    log_error "No configuration files specified"
    show_usage
    exit 1
fi

# Main execution
main() {
    echo "=== Redis-Runner Configuration Migration ==="
    echo ""
    
    # Check dependencies
    check_dependencies
    echo ""
    
    # Expand glob patterns
    EXPANDED_FILES=()
    for pattern in "${CONFIG_FILES[@]}"; do
        for file in $pattern; do
            if [ -f "$file" ]; then
                EXPANDED_FILES+=("$file")
            fi
        done
    done
    
    if [ ${#EXPANDED_FILES[@]} -eq 0 ]; then
        log_error "No configuration files found"
        exit 1
    fi
    
    # Show summary
    log_info "Migration settings:"
    echo "  Files: ${EXPANDED_FILES[*]}"
    echo "  Output format: ${OUTPUT_FORMAT:-yaml}"
    echo "  Protocol: ${PROTOCOL:-auto}"
    echo "  Backup dir: $BACKUP_DIR"
    echo "  Dry run: $DRY_RUN"
    echo ""
    
    # Confirm execution
    if [ "$AUTO_CONFIRM" = false ] && [ "$DRY_RUN" = false ]; then
        read -p "Continue with migration? (y/N): " -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Migration cancelled"
            exit 0
        fi
    fi
    
    # Execute migration
    migrate_batch "${EXPANDED_FILES[@]}"
    
    # Show report
    if [ "$DRY_RUN" = false ]; then
        show_migration_report
    fi
    
    log_success "Migration process completed"
}

# Run main function
main