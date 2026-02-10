> [!WARNING]
> This project is still in development and is not ready for use.

# AICO - AI Assistant Console
[![Go](https://github.com/micheam/ai-assistant-console/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/micheam/ai-assistant-console/actions/workflows/go.yml)
[![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/micheam/ai-assistant-console?include_prereleases)](https://github.com/micheam/ai-assistant-console/releases)

AICO is an AI-powered console application that supports multiple AI providers including OpenAI's GPT models and Anthropic's Claude models. It provides simple text generation with Vim plugin integration.

## Install

### Option 1: Quick Install with Installation Script (macOS/Linux only)

The easiest way to install AICO is using our installation script, which automatically downloads and installs the latest release:

**One-line installation:**
```bash
curl -fsSL https://raw.githubusercontent.com/micheam/ai-assistant-console/main/install.sh | bash
```

**Two-step installation (recommended for security):**

For security-conscious users, we recommend reviewing the script before execution:

```bash
# Download the installation script
curl -fsSL https://raw.githubusercontent.com/micheam/ai-assistant-console/main/install.sh -o install.sh

# Review the script contents
less install.sh

# Execute the script
bash install.sh
```

The installation script will:
- Detect your platform (OS and architecture)
- Download the latest release from GitHub
- Verify the SHA256 checksum
- Install the binary to `$HOME/.local/bin/aico`

**PATH Configuration:**

If `$HOME/.local/bin` is not in your PATH, add the following line to your `~/.bashrc` or `~/.zshrc`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

After adding this line, reload your shell configuration:
```bash
source ~/.bashrc  # or source ~/.zshrc
```

### Option 2: Download Pre-built Binaries (macOS/Linux only)

Pre-built binaries are available for macOS and Linux from the [GitHub Releases page](https://github.com/micheam/ai-assistant-console/releases).

> **Note**: Windows binaries are not provided as we don't have a Windows testing environment. Windows users should build from source.

1. Go to the [releases page](https://github.com/micheam/ai-assistant-console/releases)
2. Download the appropriate binary for your platform
3. Extract and place the binary in your PATH

### Option 3: Build from Source

To build from source, you'll need to install Go.
Make sure you have _Go version 1.20 or higher_ installed on your system. 
You can check the installed version by running `go version`.

If you do not have Go installed or your version is outdated, download and install it from the [Go website](https://golang.org/dl/).

Once you have Go installed, follow these steps to install AICO:

1. Clone the repository:
   ```bash
   git clone https://github.com/micheam/ai-assistant-console.git
   ```
2. Navigate to the project directory:
   ```bash
   cd ai-assistant-console
   ```
3. Build the executable binary by running `make`:
   ```bash
   make
   ```
   This will create a binary executable in the `bin/` directory.

Now, you can use commands as described in the [Usage](#usage) section.

## API Keys Setup

AICO supports multiple AI providers. You'll need to set up API keys for the providers you want to use:

### OpenAI API Key

To use OpenAI models (GPT-4, GPT-4o, etc.), you need an OpenAI API key.
You can get an API key from [the OpenAI API Keys page].

```bash
export AICO_OPENAI_API_KEY=<your OpenAI API key>
```

### Anthropic API Key

To use Anthropic Claude models, you need an Anthropic API key.
You can get an API key from [Anthropic Console](https://console.anthropic.com/).

```bash
export AICO_ANTHROPIC_API_KEY=<your Anthropic API key>
```

## Usage

After installation, you can use the `aico` command to generate text with AI.

```
NAME:
   aico - AI Assistant Console

USAGE:
   aico [global options] [command [command options]]

COMMANDS:
   env      show environment information
   config   Manage the configuration for the AI assistant
   models   manage AI models
   persona  manage personas
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug                  Enable debug logging
   --json                   Output in JSON format
   --model, -m              The model to use
   --no-stream              Disable streaming output
   --persona, -p            The persona to use (default: "default")
   --system                 System prompt
   --source, -s             Source string or @file path
   --context, -c            Context string or @file path
   --help, -h               show help
   --version, -v            print the version
```

### Basic Text Generation

Generate text by providing a prompt:

```bash
$ aico "Translate into English: こんにちは、世界。"
Hello, world.
```

### Using Source and Context

You can provide source content and additional context:

```bash
$ aico "Explain this code" --source=@main.go --context=@README.md --context="$(go list ./...)"
```

### Available Models

To see all available models, use the `models` command:

```bash
$ aico models
```

### Persona Management

Manage personas with the `persona` command:

```bash
$ aico persona list
```

## Usage as a Vim Plugin

AICO can be used as a Vim plugin to generate text in Vim buffers.
Please see the [Vim plugin documentation](README.vim.md) for more information.

## Environment Variables

- `AICO_OPENAI_API_KEY`: Your OpenAI API key for accessing GPT models
- `AICO_ANTHROPIC_API_KEY`: Your Anthropic API key for accessing Claude models

## Development

To contribute to AICO development, clone this repository and make the desired code changes.
Before submitting your changes, ensure the following:

- All tests pass by running `make test`
- The code formatting is consistent and adheres to [Go standards](https://golang.org/doc/effective_go)

### Testing the Installation Script

To test the installation script in a clean container environment:

```bash
./test/integration/run-integration-test.sh
```

This will:
1. Auto-detect your container runtime (Apple Container or Docker)
2. Build a container image with Ubuntu 22.04
3. Run the installation script in the container
4. Verify the installation and basic functionality

**Supported Container Runtimes:**
- [Apple Container](https://github.com/apple/container) (macOS with Apple Silicon, recommended)
- [Docker](https://www.docker.com/) (all platforms)

See [test/integration/README.md](test/integration/README.md) for more details.

## License
The AICO project is released under the [MIT License](LICENSE).


[the OpenAI API Keys page]: https://platform.openai.com/api-keys
