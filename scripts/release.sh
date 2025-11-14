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
./dist/boba bump --dry-run

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
./dist/boba bump $TYPE --dry-run

# Confirmation
echo ""
read -p "Continue with release? [y/N]: " confirm

if [[ $confirm != "y" && $confirm != "Y" ]]; then
    echo "👋 Release cancelled"
    exit 0
fi

# Perform version bump
echo ""
echo "📈 Bumping version..."
./dist/boba bump $TYPE

# Create and push release tag
echo ""
echo "🏷️ Creating release tag..."
make tag VERSION=v$(./dist/boba bump $TYPE --dry-run | grep "Next version:" | awk '{print $3}')

echo ""
echo "✅ Release preparation complete!"
echo ""
echo "📋 Summary:"
echo "  - Version has been bumped"
echo "  - Changes have been committed"
echo "  - Ready to create release tag"
echo ""
echo "🚀 Next steps:"
echo "  1. Review the commits with: git log --oneline -5"
echo "  2. If everything looks good, push changes: git push origin main"
echo "  "
echo " 3. Create and push the release tag:"
echo "     make release-$TYPE"
echo "     or manually: git tag vVERSION && git push origin vVERSION"
