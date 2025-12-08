# AICO - AI Assistant Console (Vim Plugin)

Vim plugin for AI Assistant (aico) integration.

## Requirements

- Vim 9.1.1119 or higher
- `aico` CLI tool installed and available in PATH

## Installation

### 1. Install the aico CLI

The Vim plugin requires the `aico` CLI tool to be installed separately.
Please refer to the [Install section in README.md](README.md#install) for installation instructions.

### 2. Install the Vim plugin

If you are using `vim-plug`, add the following configuration to your `.vimrc`:

```vim
Plug 'micheam/ai-assistant-console'
```

Then run the following command in Vim:

```vim
:PlugInstall
```

> **Note**: The plugin installation does not include the `aico` binary. Make sure to complete step 1 first.

## Commands

| Command | Description |
|---------|-------------|
| `:Assistant [prompt]` | Generate text using the current buffer as input |
| `:AssistantModel` | Select an AI model interactively |
| `:AssistantPersona` | Select a persona interactively |

## Usage

### Basic Usage

```vim
" Open prompt buffer to enter your prompt
:Assistant

" Or provide prompt directly
:Assistant explain this code
```

The plugin takes the content of the current buffer as input and displays the AI response in a new scratch buffer.

### Model Selection

```vim
:AssistantModel
```

Opens an interactive popup to select the AI model to use.

### Persona Selection

```vim
:AssistantPersona
```

Opens an interactive popup to select a persona. You can also select "(None - not used)" to disable persona.

## Configuration

```vim
" Set the AI model to use (default: 'claude-opus-4-5')
let g:ai_assistant_model = 'gpt-4o'

" Set the path to aico binary (default: 'aico')
let g:ai_assistant_bin = '/path/to/aico'

" Set the split position: 'above', 'below', 'left', 'right' (default: 'below')
let g:ai_assistant_split_position = 'below'

" Set the split size in lines/columns (default: 15)
let g:ai_assistant_split_size = 20

" Set the persona to use (default: '' - not used)
let g:ai_assistant_persona = 'default'
```

## Documentation

For more details on how to use the AI Assistant Console, please refer to the `doc/ai_assistant_console.txt` file.

## License

This project is licensed under the MIT License.

## Contact

If you have any issues or suggestions, please feel free to contact [me on GitHub](https://github.com/micheam).
