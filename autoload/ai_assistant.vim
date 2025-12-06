vim9script

import './uiwidget.vim' as uiwidget

if !has('patch-9.1.1119')
    echoerr "This plugin requires Vim patch 9.1.1119 or higher. Please update your Vim 🙇"
    finish
endif

const default_model = 'claude-opus-4-5'
const default_chat_command = 'aico'
const default_split_position = 'below'  # 'above', 'below', 'left', 'right'
const default_split_size = 15  # lines for horizontal, columns for vertical
const default_persona = ''  # empty string means "not used"

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

# Option: g:ai_assistant_split_position
#
# The position of the output buffer split.
# Values: 'above', 'below', 'left', 'right'
# Default: 'above'
export def SplitPosition(): string
    if exists('g:ai_assistant_split_position')
        return g:ai_assistant_split_position
    endif
    return default_split_position
enddef

export def SetSplitPosition(position: string): void
    g:ai_assistant_split_position = position
enddef

# Option: g:ai_assistant_split_size
#
# The size of the output buffer split.
# For horizontal splits (above/below): number of lines
# For vertical splits (left/right): number of columns
# Default: 15
export def SplitSize(): number
    if exists('g:ai_assistant_split_size')
        return g:ai_assistant_split_size
    endif
    return default_split_size
enddef

export def SetSplitSize(size: number): void
    g:ai_assistant_split_size = size
enddef

# Option: g:ai_assistant_persona
#
# The persona to use for text generation.
# Empty string means persona is not used.
# Default: '' (not used)
export def Persona(): string
    if exists('g:ai_assistant_persona')
        return g:ai_assistant_persona
    endif
    return default_persona
enddef

export def SetPersona(persona_name: string): void
    g:ai_assistant_persona = persona_name
enddef

# OpenSplitBuffer opens a new buffer with configured position and size.
def OpenSplitBuffer(): void
    const pos = SplitPosition()
    const size = SplitSize()

    var cmd = ''
    if pos == 'above'
        cmd = $':topleft :{size}new'
    elseif pos == 'below'
        cmd = $':botright :{size}new'
    elseif pos == 'left'
        cmd = $':vertical topleft :{size}new'
    elseif pos == 'right'
        cmd = $':vertical botright :{size}new'
    else
        # Fallback to default
        cmd = $':botright :{size}new'
    endif

    execute cmd
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

# ShowPersonaSelector shows a popup menu to select a ai-assistant persona.
export def ShowPersonaSelector(): void
    const cmd = [ ChatCommand(), "--json", "persona", "list" ]
    const personas = json_decode(system(cmd->join(' ')))
    if len(personas) == 0
        echom "No personas available."
        return
    endif

    const persona_names = personas->mapnew((_, p) => p["name"])
    # Add "None (not used)" option at the beginning
    const display_names = ["(None - not used)"] + persona_names
    const ui = uiwidget.Select.new(display_names, (selected: number) => {
        if selected == 0
            # "None" selected - clear persona
            SetPersona('')
            echom "Persona disabled"
        else
            const selectedPersona = personas[selected - 1]
            SetPersona(selectedPersona["name"])
            echom "Selected persona: " .. selectedPersona["name"]
        endif
        return true
    })
    # Set initial selection based on current persona
    const currentPersona = Persona()
    if currentPersona->empty()
        ui.SetSelectedIndex(0)  # "None" option
    else
        const currentIndex = indexof(personas, (_, p) => p["name"] == currentPersona)
        if currentIndex != -1
            ui.SetSelectedIndex(currentIndex + 1)  # +1 for "None" offset
        endif
    endif
    ui.Render()
enddef

# Script-local state for prompt buffer
var pending_input_text: string = ''

# RunAssistant runs the ai-assistant with the given prompt.
#
# This function takes the content of the current buffer as input
# and displays the AI response in a new scratch buffer.
#
# Usage:
#   :Assistant                        " Opens prompt buffer
#   :Assistant 'explain this code'
#   :Assistant summarize the following
export def RunAssistant(prompt: string): void
    # Get current buffer content before doing anything
    const bufcontent = getline(1, '$')
    const input_text = bufcontent->join("\n")

    if prompt->empty()
        # Open prompt buffer for input
        OpenPromptBuffer(input_text)
        return
    endif

    # Execute with the given prompt
    ExecuteAssistant(prompt, input_text)
enddef

# OpenPromptBuffer opens a buffer for entering the prompt.
#
# The buffer uses buftype=acwrite so :w triggers submission.
def OpenPromptBuffer(input_text: string): void
    # Save input text for later use
    pending_input_text = input_text

    # Create prompt buffer
    OpenSplitBuffer()
    setlocal buftype=acwrite
    setlocal bufhidden=wipe
    setlocal noswapfile
    setlocal filetype=markdown
    const bufname = 'Assistant(prompt:' .. localtime() .. ')'
    execute 'file ' .. fnameescape(bufname)

    # Set up autocmd for :w
    const prompt_bufnr = bufnr()
    execute 'autocmd BufWriteCmd <buffer=' .. prompt_bufnr .. '> ++once call ai_assistant#SubmitPrompt()'

    # Add placeholder text
    setline(1, ['# Enter your prompt here', '# :w to submit, :q to cancel', ''])
    cursor(3, 1)
    startinsert
enddef

# SubmitPrompt is called when user writes the prompt buffer.
export def SubmitPrompt(): void
    # Get prompt from current buffer (skip comment lines)
    const lines = getline(1, '$')
    const prompt_lines = lines->copy()->filter((_, line) => line !~# '^#')
    const prompt = prompt_lines->join("\n")->trim()

    if prompt->empty()
        echohl WarningMsg
        echo "Prompt is empty. Write your prompt and :w again, or :q to cancel."
        echohl None
        return
    endif

    # Close prompt buffer
    setlocal nomodified
    bwipeout

    # Execute with saved input
    ExecuteAssistant(prompt, pending_input_text)
    pending_input_text = ''
enddef

# ExecuteAssistant runs the ai-assistant command and displays results.
def ExecuteAssistant(prompt: string, input_text: string): void
    # Build command
    var cmd = [ChatCommand(), '--model=' .. Model()]
    # Add persona if specified
    const persona = Persona()
    if !persona->empty()
        cmd->add('--persona=' .. persona)
    endif
    cmd->add(prompt)

    # Save current window to return to it later
    const original_winid = win_getid()

    # Create a new scratch buffer for output
    OpenSplitBuffer()
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

    # Return to original window
    win_gotoid(original_winid)
    redraw

    var is_first_output = true

    # Start async job
    const job = job_start(cmd, {
        in_io: 'pipe',
        out_io: 'pipe',
        err_io: 'pipe',
        out_cb: (ch, msg) => {
            if !bufexists(output_bufnr)
                return
            endif
            if is_first_output
                # Replace the "Waiting for response..." message
                setbufline(output_bufnr, 1, msg)
                is_first_output = false
            else
                appendbufline(output_bufnr, '$', msg)
            endif
            # Scroll to bottom to show new content
            if win_id2win(output_winid) != 0
                win_execute(output_winid, 'normal! G')
            endif
            redraw
        },
        err_cb: (ch, msg) => {
            if !bufexists(output_bufnr)
                return
            endif
            if is_first_output
                setbufline(output_bufnr, 1, '[ERROR] ' .. msg)
                is_first_output = false
            else
                appendbufline(output_bufnr, '$', '[ERROR] ' .. msg)
            endif
            redraw
        },
        exit_cb: (job, status) => {
            if bufexists(output_bufnr) && win_id2win(output_winid) != 0
                if is_first_output
                    # No output received at all
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
