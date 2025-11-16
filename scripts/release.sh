#!/bin/bash

# BobaMixer Release Script
# This script demonstrates the complete release workflow

set -e

echo "🚀 BobaMixer Release Script"
echo "========================"

# Build the project first
echo "📦 Building project..."
make build

# Show current version
echo ""
echo "📋 Current version information:"
make version

# Analyze what would be bumped
echo ""
echo "🔍 Analyzing changes since last tag..."
./dist/boba-maint bump --dry-run

# Interactive prompt for release type
echo ""
read -p "Select release type [1] Patch [2] Minor [3] Major [a] Auto-detect [q] Quit: " choice

case $choice in
  1)
    TYPE="patch"
    ;;
  2)
    TYPE="minor"
    ;;
  3)
    TYPE="major"
    ;;
  a)
    TYPE="auto"
    ;;
  q)
    echo "👋 Release cancelled"
    exit 0
    ;;
  *)
    echo "❌ Invalid choice"
    exit 1
    ;;
esac

echo ""
echo "🎯 Selected release type: $TYPE"

# Show what will happen
echo ""
echo "📝 Version bump preview:"
./dist/boba-maint bump $TYPE --dry-run

# Confirmation
echo ""
read -p "Continue with release? [y/N]: " confirm

if [[ $confirm != "y" && $confirm != "Y" ]]; then
    echo "👋 Release cancelled"
    exit 0
fi

# Perform version bump
echo ""
echo "🚢 Executing release..."
./dist/boba-maint release --part "$TYPE"

echo ""
echo "✅ Release preparation complete!"
echo ""
echo "📋 Summary:"
echo "  - Version has been bumped"
echo "  - Release commit + tag created"
echo "  - Changes pushed to origin"
echo ""
echo "🚀 GitHub Actions is now building and publishing the release."
