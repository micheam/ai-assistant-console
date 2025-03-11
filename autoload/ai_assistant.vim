vim9script

import './uiwidget.vim' as uiwidget

if !has('patch-9.1.1119')
    echoerr "This plugin requires Vim patch 9.1.1119 or higher. Please update your Vim 🙇"
    finish
endif

# Option: g:ai_assistant_model
export def Model(): string
    if exists('g:ai_assistant_model')
        return g:ai_assistant_model
    endif
    return 'claude-3-5-sonnet'
enddef

export def SetModel(model_name: string): void
    g:ai_assistant_model = model_name
enddef

# Option: g:ai_assistant_chat_command
#
# The command to run the ai-assistant.
# Default: "chat"
# Example: "~/bin/chat"
export def ChatCommand(): string
    if exists('g:ai_assistant_chat_command')
        return g:ai_assistant_chat_command
    endif
    return "chat"
enddef

export def SetChatCommand(command: string): void
    g:ai_assistant_chat_command = command
enddef

# ShowModelSelector shows a popup menu to select a ai-assistant model.
export def ShowModelSelector(): void
    const cmd = [ ChatCommand(), "models", "--json" ]

    var models: list<dict<any>> = []
    for line in systemlist(cmd->join(' '))
        const model = json_decode(line)
        models->add(model)
    endfor

    # TODO: Show Description of the model.
    #       May be mappping to 'K' or something.

    const model_names = models->mapnew((_, model) => model["name"])
    const ui = uiwidget.Select.new(model_names, (selected_index: number) => {
        const selectedModel = models[selected_index]
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



