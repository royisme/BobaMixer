#!/bin/bash

# BobaMixer Development Helper Script
# Common development tasks and utilities

set -e

echo "🛠️ BobaMixer Development Helper"
echo "================================"

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -d "cmd/boba" ]; then
    echo "❌ Not in BobaMixer root directory. Please run from project root."
    exit 1
fi

# Function to show help
show_help() {
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Available commands:"
    echo "  build           Build the project (short for make build)"
    echo "  test            Run tests (short for make test)"
    echo "  check           Run all checks (lint, test, build)"
    echo "  clean           Clean build artifacts"
    echo "  install         Install to GOPATH/bin"
    echo "  version         Show current version"
    echo "  bump           Bump version (patch/minor/major/auto)"
    echo "  release         Create release tag"
    echo "  status          Show git status and version info"
    echo "  docs            Start documentation server"
    echo "  help            Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 build              # Build project"
    echo "  $0 bump patch        # Bump patch version"
    "  $0 bump minor        # Bump minor version"
    "  $0 bump auto         # Auto-detect and bump version"
    "  $0 release --auto     # Auto-release"
    ""
}

# Function to show status
show_status() {
    echo ""
    echo "📋 Project Status"
    echo "=============="
    echo ""
    "Git status:"
    git status --porcelain
    echo ""
    "Current version:"
    make version
    echo ""
    "Last 5 commits:"
    git log --oneline -5
}

# Function to start docs server
start_docs() {
    echo ""
    echo "📚 Starting documentation server..."
    cd website
    hugo server --buildDrafts --buildFuture --port 1313 --bind 127.0.0.1
}

# Main command handling
case "${1:-help}" in
    build)
        make build
        ;;
    test)
        make test
        ;;
    check)
        make check
        ;;
    clean)
        make clean
        ;;
    install)
        make install
        ;;
    version)
        make version
        ;;
    bump)
        if [ -z "$2" ]; then
            echo "❌ Please specify bump type: patch, minor, major, or auto"
            echo "   $0 bump patch  # Bump patch version"
            exit 1
        fi
        make build
        ./dist/boba bump "$2"
        ;;
    release)
        if [ "$2" = "--auto" ]; then
            make build
            ./dist/boba release --auto
        else
            echo "💡 Use './scripts/release.sh' for interactive release"
            echo "   or 'make release-patch/minor/major' for quick releases"
        fi
        ;;
    status)
        show_status
        ;;
    docs)
        start_docs
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo "❌ Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
