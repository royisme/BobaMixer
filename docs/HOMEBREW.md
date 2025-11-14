# Homebrew Setup Guide

This guide explains how to set up Homebrew formula updates for BobaMixer releases.

## Prerequisites

### 1. Create Homebrew Tap

You need to create a separate repository for your Homebrew formulas:

```bash
# Create a new repository named homebrew-tap
# Repository should be: https://github.com/royisme/homebrew-tap
```

The tap repository should be initialized with:
- README.md
- LICENSE (same as your main project)
- Empty .gitignore (optional)

### 2. GitHub Secrets Configuration

No additional secrets are required! The release workflow uses the standard `GITHUB_TOKEN` which has:

- **Contents**: read/write access to your repositories
- **Metadata**: read access
- **Pull requests**: read/write access

The `GITHUB_TOKEN` is automatically provided by GitHub Actions with the permissions specified in the workflow:

```yaml
permissions:
  contents: write
  packages: write
```

## How It Works

### Automated Process

1. **Release Trigger**: When you push a tag (e.g., `v1.2.3`)
2. **Build & Test**: Tests run, then goreleaser builds binaries
3. **GitHub Release**: Creates release with assets
4. **Homebrew Update**: Automatically updates formula in your tap

### Goreleaser Configuration

The `brews` section in `.goreleaser.yml` handles Homebrew formula generation:

```yaml
brews:
  - repository:
      owner: royisme          # Your GitHub username
      name: homebrew-tap      # Your tap repository name
      token: "{{ .Env.GITHUB_TOKEN }}"  # Uses GitHub Actions token
    homepage: "https://github.com/royisme/BobaMixer"
    description: "CLI tool for managing multiple AI providers..."
    license: MIT
    test: |
      system "#{bin}/boba version"
```

### Generated Formula Example

Goreleaser automatically creates a formula like this:

```ruby
class Boba < Formula
  desc "CLI tool for managing multiple AI providers with intelligent routing and cost tracking"
  homepage "https://github.com/royisme/BobaMixer"
  url "https://github.com/royisme/BobaMixer/archive/refs/tags/v1.2.3.tar.gz"
  sha256 "..."
  license "MIT"

  def install
    bin.install "boba"
  end

  test do
    system "#{bin}/boba version"
  end
end
```

## Manual Setup Steps

### Step 1: Create Homebrew Tap Repository

```bash
# Create repository at: https://github.com/royisme/homebrew-tap
# Clone it locally
git clone https://github.com/royisme/homebrew-tap.git
cd homebrew-tap

# Create initial structure
echo "# Homebrew Tap for BobaMixer" > README.md
echo "MIT License content" > LICENSE
git add .
git commit -m "Initial commit"
git push origin main
```

### Step 2: Verify GitHub Actions Permissions

In your main repository (`BobaMixer`), go to:
1. Settings → Actions → General
2. Scroll down to "Workflow permissions"
3. Ensure "Read and write permissions" is selected
4. Check "Allow GitHub Actions to create and approve pull requests"

### Step 3: Test the Setup

```bash
# Make a small change and create a release
echo "# Test" >> README.md
git add README.md
git commit -m "docs: update README for homebrew test"
git push origin main

# Create test release
make release-patch
```

## Troubleshooting

### Common Issues

**"Repository not found" error**
- Ensure the homebrew tap repository exists at `https://github.com/royisme/homebrew-tap`
- Check that the repository name matches exactly in goreleaser config

**"Permission denied" error**
- Verify GitHub Actions has write permissions in your repository settings
- Check that `permissions: contents: write` is set in the workflow

**Formula not updating**
- Ensure the release tag follows semantic versioning (`v1.2.3`)
- Check that goreleaser ran successfully in the Actions tab
- Verify the tap repository is accessible

### Debug Commands

```bash
# Check if goreleaser config is valid
goreleaser check

# Test release process without actually publishing
goreleaser release --snapshot --clean

# Manually verify homebrew formula
brew install royisme/tap/boba
boba version
```

### Alternative: Disable Homebrew Updates

If you don't need Homebrew integration, you can remove the `homebrew` job from `.github/workflows/release.yml`:

```yaml
# Comment out or remove this entire section
# homebrew:
#   name: Update Homebrew
#   runs-on: ubuntu-latest
#   needs: release
#   ...
```

## Installation for Users

Once set up, users can install BobaMixer via Homebrew:

```bash
# Add the tap
brew tap royisme/tap

# Install BobaMixer
brew install boba

# Update to latest version
brew upgrade boba

# Uninstall
brew uninstall boba
```

## Verification

To verify everything is working:

1. **Check GitHub Actions**: Ensure all jobs pass in your release
2. **Verify Tap Repository**: Formula should be automatically committed
3. **Test Installation**: Try `brew install royisme/tap/boba`
4. **Verify Binary**: Run `boba version` to confirm installation

The homebrew formula will be automatically updated with every release!