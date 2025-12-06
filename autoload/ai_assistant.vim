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

# RunAssistant runs the ai-assistant with the given prompt.
#
# This function takes the content of the current buffer as input
# and displays the AI response in a new scratch buffer.
#
# Usage:
#   :Assistant 'explain this code'
#   :Assistant 'summarize the following'
export def RunAssistant(prompt: string): void
    if prompt->empty()
        echoerr "Prompt is required"
        return
    endif

    # Get current buffer content
    const bufcontent = getline(1, '$')
    const input_text = bufcontent->join("\n")

    # Build command
    const cmd = [ChatCommand(), '--model=' .. Model(), prompt]

    # Create a new scratch buffer for output
    new
    setlocal buftype=nofile
    setlocal bufhidden=wipe
    setlocal noswapfile
    setlocal filetype=markdown
    const bufname = 'Assistant(' .. localtime() .. '): ' .. prompt[: 30]
    execute 'file ' .. fnameescape(bufname)

    const output_bufnr = bufnr()
    const output_winid = win_getid()

    # Show initial message
    setline(1, '# Waiting for response...')
    redraw

    var output_lines: list<string> = []

    # Start async job
    const job = job_start(cmd, {
        in_io: 'pipe',
        out_io: 'pipe',
        err_io: 'pipe',
        out_cb: (ch, msg) => {
            output_lines->add(msg)
        },
        err_cb: (ch, msg) => {
            output_lines->add('[ERROR] ' .. msg)
        },
        exit_cb: (job, status) => {
            # Update buffer with output
            if bufexists(output_bufnr) && win_id2win(output_winid) != 0
                deletebufline(output_bufnr, 1, '$')
                if output_lines->len() > 0
                    setbufline(output_bufnr, 1, output_lines)
                else
                    setbufline(output_bufnr, 1, '(No response)')
                endif
                win_execute(output_winid, 'cursor(1, 1)')
            endif
        }
    })

    # Send input to the job
    const channel = job_getchannel(job)
    ch_sendraw(channel, input_text)
    ch_close_in(channel)
enddef
