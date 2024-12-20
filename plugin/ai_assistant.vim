vim9script

# g:ai_assistant_model
#
# The model of the Assistant, e.g. `gpt-4`, `gpt-4-turbo`, `gpt-4o`, etc.
def AssistantModel(): string
    if exists('g:ai_assistant_model')
        return g:ai_assistant_model
    endif
    return 'chatgpt-4o-latest'
enddef

# buffer local variables
#
# b:context_bufs - the buffer numbers. This is used to get the context of the
#                  chat with AI-Assistant. Index 0 (primary context) is the
#                  buffer number when StartChatWindow is called.
#

def StartChatWindow()
    const src_buf = bufnr('%') # Use current buffer as a primary context source

    new # Create a new buffer
    setlocal buftype=nofile bufhidden=wipe noswapfile
    setlocal filetype=markdown
    b:context_bufs = [src_buf]
    b:chat_window = true # Mark this buffer as a chat window

    # Set the buffer name
    # The buffer name is the name of the current file
    # with the extension replaced by `.chat`
    # e.g. `foo.txt` -> `foo.chat`
    execute $"file {bufname(src_buf)}.chat"

    # Set buffer local commands
    command! -buffer          Send          SendThread()
    command! -buffer          Clear         ClearThread()
    command! -buffer          Stop          StopJob()
    command! -buffer -nargs=1 ContextAdd    AppendContextBuf(str2nr(<q-args>))
    command! -buffer          ContextAddAll {
        SyncContextBufs()
        echo "Context buffers are updated"
        ListContextBufs()
    }
    command! -buffer          ContextList   ListContextBufs()

    # Set buffer local mappings
    # <CR> - send the message to the Assistant
    # <C-c> - stop the running job
    # <C-l> - clear the chat thread
    nnoremap <buffer> <CR> :Send<CR>
    nnoremap <buffer> <C-c> :Stop<CR>
    nnoremap <buffer> <C-l> :Clear<CR>

    ShowWelcomeMessage(bufnr('%'), 1)
    execute 'normal! G'
enddef

def AppendContextBuf(bufnr: number): void
    if !exists('b:context_bufs')
        b:context_bufs = []
    endif
    b:context_bufs->add(bufnr)
enddef

def SyncContextBufs(): void
    # * b:context_bufs[0] is the primary context buffer, so we don't want to clear it.
    # * skip chat window buffer
    const primary_buf = b:context_bufs[0]
    b:context_bufs = [primary_buf]
    for bufnr in filter(range(1, bufnr('$')), 'v:val->buflisted()')
        if bufnr == primary_buf || bufnr->getbufvar('&buftype') ==# 'nofile'
            continue
        endif
        if bufnr->bufexists()
            b:context_bufs->add(bufnr)
        endif
    endfor
enddef

def ListContextBufs()
    if exists('b:context_bufs')
       for bufnr in b:context_bufs
           echo $"{bufnr}: {bufname(bufnr)}"
        endfor
    else
        echo "No context buffers"
    endif
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

    echomsg $"Send Current Buffer as a thread to the Assistant"

    # Build the message to send
    #
    # The format of the message is:
    #
    #   System:
    #
    #   Context of this chat is below:
    #
    #   <for each buffer in b:context_bufs>
    #
    #   <filename of the buffer>
    #
    #   ~~~<filetype of the buffer>
    #   <content of the buffer>
    #   ~~~
    #
    #   </for>
    #
    #   <content of the chat thread>
    #
    var messages = []
    add(messages, "System:")
    add(messages, "Context of this chat is below:")
    add(messages, "")

    for context_buf in b:context_bufs
        add(messages, bufname(context_buf))
        add(messages, "")
        add(messages, "~~~" .. getbufvar(context_buf, '&filetype'))
        extend(messages, getbufline(context_buf, 1, '$'))
        add(messages, "~~~")
        add(messages, "")
    endfor

    extend(messages, getline(1, '$')) # Append existing chat thread

    # Create a temprary file
    # Note: this tempfile will create every time we send a message.
    const tempfile = $"{tempname()}.chat.message"
    writefile(messages, tempfile)
    const cmd = [
        "chat", 
        "--model", AssistantModel(),
        "send",
        "-i",
        tempfile
    ]
    const target_buf = bufnr('%')

    def JobExitCB(job: job, status: number)
        var lines = []
        if status == -1
            lines->add("")
                ->add("> **Warning**")
                ->add("> The job is Cancelled")
        endif
        lines
            ->add("")
            ->add(HrLine())
            ->add("")
            ->add(PromptLine("User"))
            ->add("")
        appendbufline(target_buf, '$', lines)
    enddef

    # Run the external command asynchronously
    var job = job_start(cmd, {
                "out_io": "buffer",
                "out_buf": target_buf,
                "err_io": "buffer",
                "err_buf": target_buf,
                "exit_cb": JobExitCB,
                })

    # Set running job to the current buffer
    # so that we can kill the job when we close the buffer
    b:job = job

    # TODO: set the buffer read-only and disable insert mode
enddef

def ClearThread()
    # Gurd: check if the current buffer is a chat window
    if !exists('b:chat_window')
        echoerr $"The current buffer is not a chat window"
        return
    endif

    deletebufline(bufnr('%'), '$')
    ShowWelcomeMessage(bufnr('%'), 1)
    execute 'normal! G'
enddef

def ShowWelcomeMessage(buf: number, lnum: number = 1)
    setbufline(buf, lnum,  [
        PromptLine("Assistant"),
        "",
        "Please input your message.",
        "Send a message with `:Send` or `<CR>`.",
        "Clear this thread with `:Clear` or `<C-l>`.",
        "Stop the running job with `:Stop` or `<C-c>`.",
        "",
        HrLine(),
        "",
        PromptLine("User"),
        "",
        "",
    ])
enddef

def StopJob()
    if exists('b:job')
        job_stop(b:job)
    endif
enddef

def PromptLine(prompt: string): string
    return $"{prompt}: "
enddef

def HrLine(): string
    # fill the current window with `-`
    return repeat("-", winwidth(0) - 5)
enddef

defcompile

command! -nargs=? Assistant StartChatWindow(<f-args>)
