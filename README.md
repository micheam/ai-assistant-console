# AICO - AI Assistant Console 

AICO is an AI-powered text generation tool using OpenAI's GPT-3.

## Install

Since pre-built binaries are not provided, you will need to install Go to build and run AICO.
Make sure you have _Go version 1.20 or higher_ installed on your system. 
You can check the installed version by running `go version`.

If you do not have Go installed or your version is outdated, download and install it from the [Go website](https://golang.org/dl/).

Once you have Go installed, follow these steps to install AICO:

1. Clone the repository:
   ```bash
   git clone https://github.com/micheam/aico.git
   ```
2. Navigate to the project directory:
   ```bash
   cd aico
   ```
3. Use the `go.mod` file to manage dependencies. You don't need to do anything manually since Go will handle this for you.
4. Build the executable binary by running `make`:
   ```bash
   make
   ```
   This will create a binary executable in the `bin/` directory.

Now, you can use AICO as described in the [Usage](#usage) section.

## Usage

After building the binary, you can run AICO with the following command:

```bash
./bin/aico [global options] command [command options] [arguments...]
```

Global options include:

- `--debug`: Enable debug mode (default: false). Can alternatively be set with the `AICO_DEBUG` environment variable.
- `--help, -h`: Show help information
- `--version, -v`: Print the version number

Commands include:

- `version`: Show the current version of AICO
- `chat`: Start an interactive AI chat session

## Environment Variables

- `AICO_DEBUG`: Sets the debug mode of AICO. Default is `false`.

## Development

To contribute to AICO development, clone this repository and make the desired code changes. Before submitting your changes, ensure the following:

- All tests pass by running `make test`
- The code formatting is consistent and adheres to [Go standards](https://golang.org/doc/effective_go)

## License
The AICO project is released under the [MIT License](LICENSE).

