# Installation

This guide covers all methods for installing BobaMixer on different platforms.

## System Requirements

- **Operating System**: macOS, Linux, or Windows (via WSL)
- **Go Version**: 1.22+ (if building from source)
- **SQLite**: 3.x (usually pre-installed)
- **Disk Space**: ~50MB minimum (plus database growth)
- **Git**: Optional, for git hooks integration

## Installation Methods

### Method 1: Using Go Install (Recommended)

If you have Go installed, this is the simplest method:

```bash
go install github.com/royisme/bobamixer/cmd/boba@latest
```

This installs the `boba` binary to `$GOPATH/bin`. Make sure this directory is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Add this line to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to make it permanent.

### Method 2: Using Homebrew (macOS/Linux)

For macOS and Linux users, Homebrew provides an easy installation:

```bash
# Add the BobaMixer tap
brew tap royisme/tap

# Install BobaMixer
brew install bobamixer
```

Update to the latest version:

```bash
brew upgrade bobamixer
```

### Method 3: Download Pre-built Binary

Download the appropriate binary for your platform from the [Releases page](https://github.com/royisme/BobaMixer/releases).

**For Linux (amd64):**
```bash
# Download
wget https://github.com/royisme/BobaMixer/releases/latest/download/bobamixer_Linux_x86_64.tar.gz

# Extract
tar -xzf bobamixer_Linux_x86_64.tar.gz

# Move to PATH
sudo mv boba /usr/local/bin/

# Verify
boba version
```

**For macOS (Apple Silicon):**
```bash
# Download
wget https://github.com/royisme/BobaMixer/releases/latest/download/bobamixer_Darwin_arm64.tar.gz

# Extract
tar -xzf bobamixer_Darwin_arm64.tar.gz

# Move to PATH
sudo mv boba /usr/local/bin/

# Verify
boba version
```

**For macOS (Intel):**
```bash
# Download
wget https://github.com/royisme/BobaMixer/releases/latest/download/bobamixer_Darwin_x86_64.tar.gz

# Extract
tar -xzf bobamixer_Darwin_x86_64.tar.gz

# Move to PATH
sudo mv boba /usr/local/bin/

# Verify
boba version
```

### Method 4: Build from Source

For developers or if you want the latest development version:

```bash
# Clone the repository
git clone https://github.com/royisme/BobaMixer.git
cd BobaMixer

# Build
make build

# Install to GOPATH/bin
make install

# Or manually move the binary
sudo mv dist/boba /usr/local/bin/
```

**Development Setup:**

If you plan to contribute, set up the development environment:

```bash
# Install dependencies and git hooks
make dev

# Run tests
make test

# Run linter
make lint
```

See the [Contributing Guide](https://github.com/royisme/BobaMixer/blob/main/CONTRIBUTING.md) for more details.

## Post-Installation Setup

### 1. Create Configuration Directory

After installation, create the BobaMixer home directory:

```bash
mkdir -p ~/.boba/logs
chmod 700 ~/.boba
```

### 2. Initialize Configuration

Run the doctor command to generate example configurations:

```bash
boba doctor
```

This creates:
- `~/.boba/profiles.yaml` - Profile definitions
- `~/.boba/routes.yaml` - Routing rules
- `~/.boba/pricing.yaml` - Model pricing information
- `~/.boba/secrets.yaml` - API keys storage
- `~/.boba/usage.db` - SQLite database

### 3. Secure Secrets File

Ensure your secrets file has proper permissions:

```bash
chmod 600 ~/.boba/secrets.yaml
```

BobaMixer will refuse to run if `secrets.yaml` has incorrect permissions.

### 4. Verify Installation

Check that everything is working:

```bash
# Check version
boba version

# Run health check
boba doctor

# List default profiles
boba ls --profiles
```

## Platform-Specific Notes

### macOS

On macOS, you may see a security warning when first running BobaMixer. To resolve:

1. Try to run `boba`
2. Open **System Preferences** → **Security & Privacy**
3. Click **Allow Anyway** next to the BobaMixer warning
4. Run `boba` again

Alternatively, remove the quarantine attribute:

```bash
xattr -d com.apple.quarantine /usr/local/bin/boba
```

### Linux

On some Linux distributions, you may need to install SQLite:

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install sqlite3 libsqlite3-dev
```

**Fedora/RHEL:**
```bash
sudo dnf install sqlite sqlite-devel
```

**Arch Linux:**
```bash
sudo pacman -S sqlite
```

### Windows (WSL)

BobaMixer runs on Windows via WSL (Windows Subsystem for Linux):

1. Install WSL 2:
   ```powershell
   wsl --install
   ```

2. Install Ubuntu or your preferred distribution

3. Follow the Linux installation instructions within WSL

## Docker (Experimental)

You can also run BobaMixer in a Docker container:

```bash
# Build the image
docker build -t bobamixer .

# Run with mounted config
docker run -v ~/.boba:/root/.boba bobamixer stats --today
```

**Note**: Docker support is experimental and may have limitations with interactive features.

## Upgrading

### Using Homebrew

```bash
brew upgrade bobamixer
```

### Using Go Install

```bash
go install github.com/royisme/bobamixer/cmd/boba@latest
```

### Manual Upgrade

Download the latest binary and replace the existing one:

```bash
# Backup current version
cp /usr/local/bin/boba /usr/local/bin/boba.backup

# Download and install new version
# (follow download binary instructions above)

# Verify
boba version
```

### Migration Notes

When upgrading, BobaMixer will automatically migrate your database schema if needed. Always backup your data before upgrading:

```bash
# Backup configuration and database
cp -r ~/.boba ~/.boba.backup.$(date +%Y%m%d)

# Upgrade BobaMixer
# ... follow upgrade instructions ...

# Verify
boba doctor
```

## Uninstalling

### Using Homebrew

```bash
brew uninstall bobamixer
brew untap royisme/tap
```

### Manual Uninstall

```bash
# Remove binary
sudo rm /usr/local/bin/boba

# Optionally remove configuration (this deletes all data!)
rm -rf ~/.boba
```

To keep your data for future reinstallation, don't delete `~/.boba`.

## Troubleshooting Installation

### Command Not Found

If you get "command not found" after installation:

1. Check that the binary is in your PATH:
   ```bash
   which boba
   ```

2. If empty, add to PATH:
   ```bash
   export PATH=$PATH:/usr/local/bin
   ```

3. Make it permanent by adding to your shell profile

### Permission Denied

If you get "permission denied":

```bash
# Make executable
chmod +x /usr/local/bin/boba

# Or reinstall with sudo
sudo cp boba /usr/local/bin/
```

### SQLite Issues

If you see database errors:

```bash
# Check SQLite version
sqlite3 --version

# Should be 3.x or higher
# If not, install/upgrade SQLite
```

### Git Hooks Not Working

If git hooks fail to install:

```bash
# Check git version
git --version

# Install hooks manually
cd your-project
cp ~/.boba/git-hooks/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit
```

## Next Steps

After installation, proceed to:

- **[Getting Started](/guide/getting-started)** - Configure your first profile
- **[Configuration Guide](/guide/configuration)** - Learn about all configuration options
- **[CLI Reference](/reference/cli)** - Explore all available commands

## Getting Help

If you encounter installation issues:

1. Check the [Troubleshooting Guide](/advanced/troubleshooting)
2. Search [GitHub Issues](https://github.com/royisme/BobaMixer/issues)
3. Open a new issue with:
   - Your OS and version
   - Installation method used
   - Error messages
   - Output of `boba doctor`
