> [!WARNING]
> This project is still in development and is not ready for use.

# AICO - AI Assistant Console
[![Go](https://github.com/micheam/ai-assistant-console/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/micheam/ai-assistant-console/actions/workflows/go.yml)
[![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/micheam/ai-assistant-console?include_prereleases)](https://github.com/micheam/ai-assistant-console/releases)

![screenshot](screenshot.png)

AICO is an AI-powered console application that supports multiple AI providers including OpenAI's GPT models and Anthropic's Claude models. It provides both interactive REPL and batch processing modes, along with advanced features like chat session management and Vim plugin integration.

## Install

### Option 1: Download Pre-built Binaries (macOS/Linux only)

Pre-built binaries are available for macOS and Linux from the [GitHub Releases page](https://github.com/micheam/ai-assistant-console/releases). 

> **Note**: Windows binaries are not provided as we don't have a Windows testing environment. Windows users should build from source.

1. Go to the [releases page](https://github.com/micheam/ai-assistant-console/releases)
2. Download the appropriate binary for your platform
3. Extract and place the binary in your PATH

### Option 2: Build from Source

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
export OPENAI_API_KEY=<your OpenAI API key>
```

### Anthropic API Key

To use Anthropic Claude models, you need an Anthropic API key.
You can get an API key from [Anthropic Console](https://console.anthropic.com/).

```bash
export ANTHROPIC_API_KEY=<your Anthropic API key>
```

## Usage

After installation, you can use the `chat` command to interact with AI assistants.

```
NAME:
   chat - AI Assistant Console

USAGE:
   chat [global options] [command [command options]]

COMMANDS:
   models   Show the available models
   repl     Start a chat session in a REPL
   send     Send a message to the AI assistant
   session  Manage chat sessions
   config   Manage the configuration for the AI assistant
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### REPL Mode (Interactive Chat)

To start an interactive chat session, use the `repl` subcommand:

```bash
$ chat repl
```

In REPL mode, you can have interactive conversations with AI assistants.

### Available Models

To see all available models, use the `models` command:

```bash
$ chat models
```

### Batch Mode

To send a single message to AI, use the `send` subcommand:

```bash
$ chat send "Translate into English: こんにちは、世界。"
Hello, world.
```

### Session Management

AICO now supports chat session management. You can:

- List sessions: `chat session list`
- Show a session: `chat session show <session-id>`
- Delete a session: `chat session delete <session-id>`
- Prepare an empty session: `chat session prepare`

Sessions can be exported in multiple formats including markdown, JSON, and plain text.

## Usage as a Vim Plugin

AICO can be used as a Vim plugin to generate text in Vim buffers.
Please see the [Vim plugin documentation](README.vim.md) for more information.

## Environment Variables

- `OPENAI_API_KEY`: Your OpenAI API key for accessing GPT models
- `ANTHROPIC_API_KEY`: Your Anthropic API key for accessing Claude models

## Development

To contribute to AICO development, clone this repository and make the desired code changes.
Before submitting your changes, ensure the following:

- All tests pass by running `make test`
- The code formatting is consistent and adheres to [Go standards](https://golang.org/doc/effective_go)

## License
The AICO project is released under the [MIT License](LICENSE).


[the OpenAI API Keys page]: https://platform.openai.com/api-keys
