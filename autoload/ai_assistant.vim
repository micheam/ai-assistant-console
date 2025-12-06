vim9script

import './uiwidget.vim' as uiwidget

if !has('patch-9.1.1119')
    echoerr "This plugin requires Vim patch 9.1.1119 or higher. Please update your Vim 🙇"
    finish
endif

const default_model = 'claude-opus-4-5'
const default_chat_command = 'aico'

# Option: g:ai_assistant_model
export def Model(): string
    if exists('g:ai_assistant_model')
        return g:ai_assistant_model
    endif
    return default_model
enddef

export def SetModel(model_name: string): void
    g:ai_assistant_model = model_name
enddef

# Option: g:ai_assistant_bin
#
# The command to run the ai-assistant.
# Default: "aico"
# Example: "/path/to/bin/aico"
export def ChatCommand(): string
    if exists('g:ai_assistant_bin')
        return g:ai_assistant_bin
    endif
    return default_chat_command
enddef

export def SetChatCommand(command: string): void
    g:ai_assistant_bin = command
enddef

# ShowModelSelector shows a popup menu to select a ai-assistant model.
#
# TODO: Show Description of the model.
#       May be mappping to 'K' or something.
export def ShowModelSelector(): void
    const cmd = [ ChatCommand(), "models", "--json" ]
    const models = json_decode(system(cmd->join(' ')))
    if len(models) == 0
        echom "No models available."
        return
    endif

    const model_names = models->mapnew((_, model) => model["name"])
    const ui = uiwidget.Select.new(model_names, (selected: number) => {
        const selectedModel = models[selected]
        SetModel(selectedModel["name"])
        echom "Selected model: " .. selectedModel["name"]
        return true
    })
    const currentModelIndex = indexof(models, (_, model) => model["name"] == Model())
    if currentModelIndex != -1
        ui.SetSelectedIndex(currentModelIndex)
    endif
    ui.Render()
enddef

# ShowChatWindow shows a chat window.
# If the chat window is already opened, it will focus on the window.
# Chat window is a prompt-buffer with a chat prompt.
# 
# Example:
#
#  you> Hello, How are you today?
#  I'm fine, thank you. How can I help you?
#
#  you> What is the weather in Tokyo?
#  It's sunny today in Tokyo.
#
export def ShowChatWindow(): void
    const model = Model()
    const cmd = [ ChatCommand(), "send", "--model", model ]

    if bufexists('AI Assistant')
        execute "buffer 'AI Assistant'"
    else
        new 'AI Assistant'
        set buftype=prompt
        var buf = bufnr('')

        b:ai_assistant_session = systemlist([ ChatCommand(), "session", "prepare" ]->join(' '))->join('')
        prompt_setcallback(buf, HandleTextEntered)
        prompt_setprompt(buf, "you> ")
    endif

    startinsert
enddef

def HandleTextEntered(text: string): void
    if text == ''
        return
    endif

    const model = Model()
    const buf = bufnr('')

    const cmd = [
        ChatCommand(), "send",
        "--model", model,
        "--session", b:ai_assistant_session,
        text,
    ]

    # TODO: use --stream
    const response = systemlist(cmd->join(' '))

    append(line('$') - 1, $"\n{model}:")
    for line in response
        append(line('$') - 1, line)
    endfor
enddef



