vim9script

def StartChatWindow()
    var src_buf = bufnr('%')

    # Create a new buffer
    new
    setlocal buftype=nofile bufhidden=wipe noswapfile
    setlocal filetype=markdown
    b:src_buf = src_buf
    b:chat_window = true # Mark this buffer as a chat window

    # Set the buffer name
    # The buffer name is the name of the current file
    # with the extension replaced by `.chat`
    # e.g. `foo.txt` -> `foo.chat`
    execute $"file {bufname(src_buf)}.chat"

    # Set buffer local commands
    # :Send - send the message to the Assistant
    # :Clear - clear the chat thread
    command! -buffer Send call SendThread()
    command! -buffer Clear call ClearThread()

    ShowWelcomeMessage(1)
    execute 'normal! G'
enddef

def SendThread()

    # Gurd: check if the current buffer is a chat window
    if !exists('b:chat_window')
        echoerr $"The current buffer is not a chat window"
        return
    endif

    # Add a new line to the end of the buffer
    # if the last line is not empty
    if getline('$') !~# '\n$'
        append(line('$'), '')
    endif

    echo $"Send Current Buffer as a thread to the Assistant"

    # Build the message to send
    #
    # The format of the message is:
    #
    #   System:
    #   src content is below:
    #
    #   ~~~<filetype>
    #
    #   <content of the content from b:src_buf>
    #
    #   ~~~
    #
    #   <content of the chat thread>
    #
    var messages = []
    add(messages, "System:")
    add(messages, "src content is below:")
    add(messages, "")
    add(messages, "~~~" .. &filetype)
    extend(messages, getbufline(b:src_buf, 1, '$'))
    add(messages, "~~~")
    add(messages, "")
    extend(messages, getline(1, '$'))

    # Create a temprary file
    # Note: this tempfile will create every time we send a message.
    const tempfile = $"{tempname()}.chat.message"
    writefile(messages, tempfile)
    const cmd = ["chat", "send",
        "-winwidth", winwidth(0) - 5,
        "-i", tempfile]

    def JobExitCB(job: job, status: number)
        # Add a new line to the end of the buffer
        # if the last line is not empty
        if getline('$') !~# '\n$'
            append(line('$'), '')
        endif
        setline(line('$'), [
            "",
            PromptLine("User: ", winwidth(0) - 5),
            "",
        ])
        # move the cursor to the last line
        execute 'normal! G'
    enddef

    # Run the external command asynchronously
    var job = job_start(cmd, {
                "out_io": "buffer",
                "out_buf": bufnr('%'),
                "err_io": "buffer",
                "err_buf": bufnr('%'),
                "exit_cb": JobExitCB,
                })

    # TODO: set the buffer read-only and disable insert mode
enddef

def ClearThread()
    # Gurd: check if the current buffer is a chat window
    if !exists('b:chat_window')
        echoerr $"The current buffer is not a chat window"
        return
    endif
    execute 'normal! ggdG'
    ShowWelcomeMessage(1)
    execute 'normal! G'
enddef

def ShowWelcomeMessage(lnum: number = 1)
    setline(lnum,  [
        PromptLine("Assistant: ", winwidth(0) - 5),
        "",
        "Please input your message.",
        "Send a message with `:Send` command.",
        "Clear this thread with `:Clear` command.",
        "",
        PromptLine("User: ", winwidth(0) - 5),
        "",
        "",
    ])
enddef

def PromptLine(prompt: string, width: number = 80): string
    echom $"prompt: {prompt}, width: {width}"
    var line = prompt
    while len(line) < width
        line = $"{line}-"
    endwhile
    return line
enddef

defcompile

command! -nargs=? Assistant StartChatWindow(<f-args>)