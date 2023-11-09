> [!WARNING]
> This project is still in development and is not ready for use.

# AICO - AI Assistant Console 

![screenshot](screenshot.png)

AICO is an AI-powered text generation tool using OpenAI's GPT-4.

## Install

Since pre-built binaries are not provided, you will need to install Go to build and run AICO.
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
3. Use the `go.mod` file to manage dependencies. You don't need to do anything manually since Go will handle this for you.
4. Build the executable binary by running `make`:
   ```bash
   make
   ```
   This will create a binary executable in the `bin/` directory.

Now, you can use commands as described in the [Usage](#usage) section.

## Usage of `chat` Command

After building the binary, you can run `chat` with the following command:

```bash
$ ./bin/chat -h
NAME:
   chat - Chat with AI

USAGE:
   chat [global options] command [command options] [arguments...]

VERSION:
   0.0.7


COMMANDS:
   config   Show config file path
   tui      Chat with AI in TUI
   send     Send message to AI
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug                    Enable debug mode (default: false) [$AICO_DEBUG]
   --model value, -m value    GPT model to use (default: "gpt-4")
   --persona value, -p value  Persona to use (default: "default")
   --help, -h                 show help
   --version, -v              print the version
```

## Usage as a Vim Plugin

AICO can be used as a Vim plugin to generate text in Vim buffers.
Please see the [Vim plugin documentation](README.vim.md) for more information.

## Environment Variables

- `AICO_DEBUG`: Sets the debug mode of AICO. Default is `false`.

## Development

To contribute to AICO development, clone this repository and make the desired code changes.
Before submitting your changes, ensure the following:

- All tests pass by running `make test`
- The code formatting is consistent and adheres to [Go standards](https://golang.org/doc/effective_go)

## License
The AICO project is released under the [MIT License](LICENSE).

