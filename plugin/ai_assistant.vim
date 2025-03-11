vim9script
#===============================================================================
#
# Vim plugin for AI Assistant chat window functionality
#
# This script provides an interactive chat window for AI Assistant integration.
# It includes functionality to manage context buffers, send/receive messages,
# and handle asynchronous jobs.
#
# Maintainer:  Michito Maeda <michito.maeda@gmail.com>
# Last Change: 2022-10-10
#
#===============================================================================
import autoload '../autoload/ai_assistant.vim' as ai_assistant

command! -nargs=0 AssistantModel ai_assistant.ShowModelSelector()
