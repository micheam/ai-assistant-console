# AICO - AI Assistant Console
[![Go](https://github.com/micheam/ai-assistant-console/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/micheam/ai-assistant-console/actions/workflows/go.yml)
[![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/micheam/ai-assistant-console?include_prereleases)](https://github.com/micheam/ai-assistant-console/releases)

AICO is a Unix-friendly CLI for LLM chat and text generation. It provides one interface for multiple AI providers — Anthropic Claude, OpenAI GPT, Groq, and Cerebras — with streaming responses, reusable personas, and file-based context. Pipe stdin, reference files with `@path`, and bring LLMs into your shell workflows. It can also be used from Vim via the [vim-aico](https://github.com/micheam/vim-aico) plugin.

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
Make sure you have _Go version 1.25 or higher_ installed on your system. 
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
   This will create a binary executable in the `dist/` directory.

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

### Groq API Key

To use models hosted on Groq, you need a Groq API key.
You can get an API key from the [Groq Console](https://console.groq.com/keys).

```bash
export AICO_GROQ_API_KEY=<your Groq API key>
```

### Cerebras API Key

To use models hosted on Cerebras, you need a Cerebras API key.
You can get an API key from [Cerebras Cloud](https://cloud.cerebras.ai/).

```bash
export AICO_CEREBRAS_API_KEY=<your Cerebras API key>
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
   session  Manage chat sessions
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug                                                      Enable debug logging (default: false)
   --json                                                       Output in JSON format (default: false)
   --model string, -m string                                    Model to use (e.g., 'gpt-4o' or 'openai:gpt-4o' for explicit provider)
   --session ID                                                 session ID for conversation history
   --last                                                       resume the most recent session (default: false)
   --no-stream                                                  disable streaming output (default: false)
   --persona string, -p string                                  The persona to use (default: "default")
   --system string                                              system prompt
   --source string, -s string                                   source string or @file path - the primary subject of the prompt (e.g., --source @code.go)
   --context string, -c string [ --context string, -c string ]  context string or @file path (e.g., --context 'text' or --context @file.txt)
   --anthropic-api-key string                                   Anthropic API Key [$AICO_ANTHROPIC_API_KEY]
   --openai-api-key string                                      OpenAI API Key [$AICO_OPENAI_API_KEY]
   --groq-api-key string                                        Groq API Key [$AICO_GROQ_API_KEY]
   --cerebras-api-key string                                    Cerebras API Key [$AICO_CEREBRAS_API_KEY]
   --help, -h                                                   show help
   --version, -v                                                print the version
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

### Piping from Stdin

When no `--source` is given, AICO reads the source from stdin, so it fits naturally into shell pipelines:

```bash
$ git diff --staged | aico "Write a commit message for this change"
```

### Chat Sessions

Conversation history is stored as sessions. Use `--last` to continue the most recent conversation, or `--session` to resume a specific one:

```bash
$ aico "What are goroutines?"
$ aico --last "Show me an example"
```

Manage stored sessions with the `session` command:

```bash
$ aico session list
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

AICO can be used from Vim to generate text in Vim buffers.
The Vim plugin lives in a separate repository: [micheam/vim-aico](https://github.com/micheam/vim-aico).
Please see its README for installation and usage.

## Environment Variables

- `AICO_OPENAI_API_KEY`: Your OpenAI API key for accessing GPT models
- `AICO_ANTHROPIC_API_KEY`: Your Anthropic API key for accessing Claude models
- `AICO_GROQ_API_KEY`: Your Groq API key for accessing models hosted on Groq
- `AICO_CEREBRAS_API_KEY`: Your Cerebras API key for accessing models hosted on Cerebras

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
