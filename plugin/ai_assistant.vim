vim9script
#===============================================================================
#
# Vim plugin for AI Assistant (aico) integration.
#
# This plugin provides commands to interact with the aico CLI tool:
#   :AssistantModel    - Select an AI model interactively
#   :AssistantPersona  - Select a persona interactively
#   :Assistant [prompt] - Generate text using the current buffer as input
#
# Maintainer:  Michito Maeda <michito.maeda@gmail.com>
# Last Change: 2025-12-06
#
#===============================================================================
import autoload '../autoload/ai_assistant.vim' as ai_assistant

command! -nargs=0 AssistantModel ai_assistant.ShowModelSelector()
command! -nargs=0 AssistantPersona ai_assistant.ShowPersonaSelector()
command! -nargs=* Assistant ai_assistant.RunAssistant(<q-args>)
