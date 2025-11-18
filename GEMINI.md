# BobaMixer Project Overview for Gemini

This document provides a comprehensive overview of the BobaMixer project, detailing its purpose, technologies, architecture, and development practices, derived from its `README.md` file. This information is intended to serve as instructional context for future interactions with Gemini.

## Project Overview

BobaMixer is an intelligent router and cost optimizer designed for AI workflows. It acts as a control plane for managing various AI service providers, local CLI tools, and their bindings. Its core functionalities include auto-injecting credentials, running local AI CLI tools, and providing an optional local HTTP proxy to consolidate requests. Advanced features encompass intelligent routing based on task characteristics, multi-level budget management with alerts, and precise usage analytics for cost tracking. The project is primarily written in Go (version 1.25 or newer) and leverages SQLite for local storage, and the Bubble Tea framework for its terminal user interface (TUI). It emphasizes a modular design, robust error handling, and adherence to Go best practices.

## Building and Running

To get started with BobaMixer, follow these steps:

### Installation

*   **Using Go:**
    ```bash
    go install github.com/royisme/bobamixer/cmd/boba@latest
    ```
*   **Using Homebrew:**
    ```bash
    brew tap royisme/tap
    brew install bobamixer
    ```

### First Time Setup (Interactive Onboarding)

*   Run `boba` in your terminal. The onboarding wizard will automatically guide you through configuration, API key input, and verification.

### Alternative CLI Setup (for power users)

1.  **Initialize config directory:**
    ```bash
    boba init
    ```
2.  **Configure API Key:** (e.g., for Anthropic)
    ```bash
    boba secrets set claude-anthropic-official
    ```
3.  **Bind tool to Provider:** (e.g., bind `claude` tool to `claude-anthropic-official` provider)
    ```bash
    boba bind claude claude-anthropic-official
    ```
4.  **Verify configuration:**
    ```bash
    boba doctor
    ```
5.  **Run a tool:** (e.g., run the `claude` tool to check its version)
    ```bash
    boba run claude --version
    ```

### Using Environment Variables

*   Set API keys as environment variables (e.g., `export ANTHROPIC_API_KEY="sk-ant..."`). BobaMixer prioritizes environment variables.

### Build from Source (for Developers)

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/royisme/BobaMixer.git
    ```
2.  **Navigate into the directory:**
    ```bash
    cd BobaMixer
    ```
3.  **Install dependencies:**
    ```bash
    go mod download
    ```
4.  **Build:**
    ```bash
    make build
    ```
5.  **Run tests:**
    ```bash
    make test
    ```
6.  **Lint check:**
    ```bash
    make lint
    ```

## Development Conventions

The BobaMixer project adheres to strict Go language standards and best practices:

*   **Code Quality:**
    *   All exported types and functions must have documentation comments.
    *   `golangci-lint` is used for static analysis, with a target of 0 issues.
    *   Follow the [Effective Go](https://go.dev/doc/effective_go) guide.
    *   Run `make test && make lint` before committing changes.
*   **Error Handling:** Complete error wrapping and graceful degradation are implemented.
*   **Concurrency Safety:** `sync.RWMutex` is used to protect shared state, ensuring thread-safe operations.
*   **Security:** All exceptions are marked with `#nosec` after an audit.
