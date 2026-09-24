# AICO - AI Assistant Console
[![Go](https://github.com/micheam/ai-assistant-console/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/micheam/ai-assistant-console/actions/workflows/go.yml)
[![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/micheam/ai-assistant-console?include_prereleases)](https://github.com/micheam/ai-assistant-console/releases)

```
    ▄▄▄  ▄▄  ▄▄▄▄  ▄▄▄
   ██▀██ ██ ██▀▀▀ ██▀██
   ██▀██ ██ ▀████ ▀███▀
   AI-Assistant-Console
```

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
   --source string, -s string                                   the ONE primary subject to act on (see --context)
   --context string, -c string [ --context string, -c string ]  read-only reference material for the prompt; repeatable
   --tool string [ --tool string ]                               enable a client-side tool by name (available: propose_edit); repeatable
   --anthropic-api-key string                                   Anthropic API Key [$AICO_ANTHROPIC_API_KEY]
   --openai-api-key string                                      OpenAI API Key [$AICO_OPENAI_API_KEY]
   --groq-api-key string                                        Groq API Key [$AICO_GROQ_API_KEY]
   --cerebras-api-key string                                    Cerebras API Key [$AICO_CEREBRAS_API_KEY]
   --help, -h                                                   show help
   --version, -v                                                print the version
```

See the [Source vs. Context](#source-vs-context) section below for what `--source`/`--context` accept beyond a plain string (`@file`, `@-`, `label:@...`).

### Basic Text Generation

Generate text by providing a prompt:

```bash
$ aico "Translate into English: こんにちは、世界。"
Hello, world.
```

### Source vs. Context

`--source` (`-s`) and `--context` (`-c`) both feed input to the prompt, but they mean different things to the model:

- **`--source`** is the *primary subject* — the thing your output is about, and the thing that might someday be written back to (a Vim buffer, a file under review). There is exactly one per prompt; passing it twice is an error.
- **`--context`** is *read-only reference material* — supporting evidence used to judge or explain the source, never the thing being acted on. It can be repeated.

Swapping which one you use for the same file changes the meaning of the request:

```bash
# The README is the subject; serve.go is evidence used to judge it.
$ aico "Point out anything in the README that is now stale" \
       --source=@README.md --context=@cmd/serve.go

# serve.go is the subject; the README is evidence used to judge it.
$ aico "Update this code to match the documented behavior" \
       --source=@cmd/serve.go --context=@README.md
```

Both flags accept:

| Form | Meaning |
| --- | --- |
| `text` | inline string, used as-is |
| `@path/to/file` | file contents |
| `@-` | stdin (see below) |
| `label:@path` | file contents, labeled |
| `label:@-` | stdin, labeled |

- A `file="..."` and/or `name="..."` attribute is attached so the model can tell blocks apart (see below).
- `label:` is only recognized in front of an `@`-prefixed value; plain inline text is never split on `:`, so URLs, `go doc` output, and similar text pass through untouched.

```bash
$ aico "Explain this code" --source=@main.go --context=@README.md --context="$(go doc ./cmd/aico)"
```

### Piping from Stdin

`@-` reads from stdin, for either flag:

```bash
$ git diff --staged | aico "Write a commit message for this change" --source=@-
$ echo "team style guide..." | aico "Review this PR" --source=@pr.diff --context="style:@-"
```

When neither `--source` nor `--context` claims stdin with `@-`, AICO falls back to reading piped input as an unlabeled source — this is what lets it drop into a plain shell pipeline without any flags, and is also how an editor integration like [vim-aico](https://github.com/micheam/vim-aico) sends buffer contents by default:

```bash
$ git diff --staged | aico "Write a commit message for this change"
```

Only one `@-` may be used per invocation (across `--source` and all `--context` values combined); a second one is an error, since stdin can only be read once.

### Chat Sessions

Conversation history is stored as sessions. Use `--last` to continue the most recent conversation, or `--session` to resume a specific one:

```bash
$ aico "What are goroutines?"
$ aico --last "Show me an example"
```

`--context` is resolved fresh on every turn, so you can add or swap it freely when resuming a session:

```bash
$ aico --source=@main.go "Review this file"
$ aico --last --context=@CHANGELOG.md "Given the changelog, is this still accurate?"
```

`--source`, on the other hand, identifies the session's ongoing subject: the file/label from the first turn that supplies one is recorded on the session (see `session show`) and isn't overwritten by later turns.

Manage stored sessions with the `session` command:

```bash
$ aico session list
```

### Client-side tools (`--tool propose_edit`)

`--tool propose_edit` lets the model return concrete edits to the `--source` text as structured proposals instead of prose or diff code blocks:

```bash
$ echo 'x = 1' | aico --tool propose_edit "rename x to y"
```

A proposal is never applied automatically — it is only shown to you (or to the calling editor integration; see [vim-aico](https://github.com/micheam/vim-aico), which applies a chosen proposal on explicit confirmation). Enabling a tool is recorded on the session (`session show` includes a `tools` field) so it's sent again on every later turn that resumes the same session — the API requires the same tool definitions whenever a session's history contains a tool call. Because of this, treat `--tool` as something you set at the start of a session; adding it to an existing session mid-conversation is unsupported.

`session show` renders a proposal as `[propose_edit] <description>` followed by `-`/`+` lines for the replaced and replacement text; `session show --json` includes the underlying `tool_use`/`tool_result` content blocks.

### Available Models

To see all available models, use the `models` command:

```bash
$ aico models
```

#### Model aliases

Instead of a versioned model name you can use a short family alias, which resolves to the latest supported model of that family at run time. `aico models` lists each alias next to the model it currently resolves to.

```toml
model = "anthropic:fable"
```

```bash
$ aico models
anthropic:claude-fable-5-1 (fable)
openai:gpt-6-luna (luna)
...
```

- Aliases work anywhere a model name is accepted (`--model`, `model` in `config.toml`), with or without the provider prefix (`fable`, `anthropic:fable`).
- Sessions store the resolved model name, so an existing session keeps its model even after an alias moves on.
- `[models."..."]` settings are looked up by the resolved model name, not by alias.

#### Per-model settings

You can tune `effort` and `max_tokens` per model in `config.toml`. Only models listed there get any settings; everything else runs on the provider's defaults.

```toml
[models."claude-opus-5-5"]
effort = "high"

[models."claude-opus-5"]
effort = "anthropic:xhigh"
max_tokens = 65536

[models."gpt-5.6-sol"]
effort = "high"
```

- Always quote the key: model names may contain `.` or `:`.
- `effort` accepts `low`, `medium` and `high`, which every provider understands. Prefix a provider-specific value with its provider (`anthropic:xhigh`, `openai:minimal`) to pass it through as-is.
- Configuring `effort` for a model that doesn't support it (for example a non-reasoning OpenAI model) results in an API error.

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
