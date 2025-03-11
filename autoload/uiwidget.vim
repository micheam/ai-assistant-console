vim9script
#==============================================================================
# Description: A Vim9 script for a MultiSelect popup menu widget that allows 
#              users to select multiple items from a list.
# Author: Michito Maeda <michto.maeda@gmail.com>
# License: MIT
# Created: 2025-01-20
# Last Updated: 2025-03-11T10:45:15+09:00
# Version: 0.1.0
# Dependencies: 
#   - Vim patch 9.1.1119 or higher
#==============================================================================

# MultiSelect is a popup menu widget that allows users to select multiple items.
#
# Usage:
#   This script provides a `MultiSelect` class that can be used to create 
#   and manage a popup menu for multi-selection. Example usage:
#
#   ```vim
#   const items = ["OPTION 1", "OPTION 2", "OPTION 3"]
#   const selector = MultiSelect.new(items, (selected_indices): bool => {
#       if selected_indices->len() == 0
#           echom "No item selected"
#           return false
#       endif
#       echom $"Selected items: {selected_indices->map(i => items[i])}"
#       return true
#   })
#   selector.Render()
#   ```
#
# Features:
#   - Popup menu for multi-selection
#   - Keyboard navigation and selection (Enter, Space, Esc)
#   - Customizable handler for selection logic
#
# Todos:
#   - Add support for controlling popup window title
#   - Add support for controlling popup window position
#   - Add support for customizing popup window style
#     (border, borderchars, etc.)
#   - Enable to set the initial selected items.
export class MultiSelect

    var Handler: func(list<number>): bool

    # Internal state
    var items: list<string>
    var winid: number = -1
    var selected_indices: list<number> # elements are 0-based indices

    # Style settings
    public var borderchars = ['─', '│', '─', '│', '╭', '╮', '╯', '╰']
    public var border = [1, 1, 1, 1]

    def new(this.items, this.Handler)
    enddef
    
    # Render renders the widget
    def Render()
        if this.items->len() == 0
            echoerr "No items to render"
            return
        endif

        const popup_contents = this.items->mapnew((idx, elm) => {
            return (this.selected_indices->index(idx) != -1) ? $'* {elm}' : $'  {elm}'
        })

        if this.winid > 0
            # Refresh the popup text
            popup_settext(this.winid, popup_contents)
            return
        endif

        const winid = popup_create(popup_contents, {
            # see: `:h popup_create-arguments` for more items
            minwidth: 50,
            mapping: false,
            filter: this._handleKeyPress,
            drag: true,
            cursorline: true,

            border: this.border,
            borderchars: this.borderchars,
        })
        this.winid = winid
    enddef

    # Close closes the widget
    def Close()
        if this.winid > 0
            popup_close(this.winid)
            this.winid = -1
        endif
    enddef

    def SetSelectedIndices(selected_indices: list<number>)
        this.selected_indices = selected_indices
    enddef

    # Handle key press event, return true to consume the key
    def _handleKeyPress(winid: number, key: string): bool
        const linenum = getcurpos(winid)[1] # 1-based line number
        if key ==# "\<CR>" # Select current item
            this.selected_indices->sort()
            const to_be_closed = this.Handler(this.selected_indices)
            if to_be_closed
                this.Close()
            endif
            return true
        elseif key == "\<Space>" # Select current item
            this._toggle_selection(linenum - 1)
            this.Render()
            return true
        elseif key ==# "\<ESC>" # Close popup
            this.Close()
            return true
        endif

        return popup_filter_menu(winid, key)
    enddef

    # Toggle selection of an option 
    # This will mutate this.selected_indices.
    #
    # Note: `selected` must be a 0-based index of the this.option
    def _toggle_selection(selected: number)
        if selected < 0 || selected >= this.items->len()
            return
        endif
        const found_at = this.selected_indices->index(selected)
        if found_at == -1
            this.selected_indices->add(selected)
            return
        endif
        this.selected_indices->remove(found_at)
    enddef
endclass

# Select is a popup menu widget that allows users to select single item.
#
# Usage:
#   This script provides a `Select` class that can be used to create 
#   and manage a popup menu for single selection. Example usage:
#
#   ```vim
#   const items = ["OPTION 1", "OPTION 2", "OPTION 3"]
#   const selector = Select.new(items, (selected_indices): bool => {
#       if selected_indices->len() == 0
#           echom "No item selected"
#           return false
#       endif
#       echom $"Selected items: {selected_indices->map(i => items[i])}"
#       return true
#   })
#   selector.Render()
#   ```
#
# Features:
#   - Keyboard navigation and selection (Enter, Space, Esc)
#   - Customizable handler for selection logic
#
# Todos:
#   - Add support for controlling popup window title
#   - Add support for controlling popup window position
#   - Add support for customizing popup window style
#     (border, borderchars, etc.)
#   - Enable to set the initial selected items.
export class Select
    var Handler: func(number): bool

    # Internal state
    var items: list<string>
    var winid: number = -1
    var selected_index: number = -1

    # Style settings
    public var borderchars = ['─', '│', '─', '│', '╭', '╮', '╯', '╰']
    public var border = [1, 1, 1, 1]

    def new(this.items, this.Handler)
    enddef

    # Render renders the widget
    # Note: This method should be called only once.
    def Render()
        if this.items->len() == 0
            echoerr "No items to render"
            return
        endif

        const popup_contents = this.items->mapnew((idx, elm) => {
            return (this.selected_index == idx) ? $'* {elm}' : $'  {elm}'
        })

        if this.winid > 0
            # Refresh the popup text
            popup_settext(this.winid, popup_contents)
            return
        endif

        const winid = popup_create(popup_contents, {
            # see: `:h popup_create-arguments` for more items
            minwidth: 50,
            mapping: false,
            filter: this._handleKeyPress,
            drag: true,
            cursorline: true,

            border: this.border,
            borderchars: this.borderchars,
        })
        this.winid = winid
    enddef

    # Close closes the widget
    def Close()
        if this.winid > 0
            popup_close(this.winid)
            this.winid = -1
        endif
    enddef

    def SetSelectedIndex(selected_index: number)
        this.selected_index = selected_index
    enddef

    # _handleKeyPress handles key press event, return true to consume the key
    def _handleKeyPress(winid: number, key: string): bool
        const linenum = getcurpos(winid)[1] # 1-based line number
        if key ==# "\<CR>" # Select current item
            this.Handler(linenum - 1)
            this.Close()
            return true
        elseif key == "\<Space>" # Select current item
            this.selected_index = linenum - 1
            this.Render()
            return true
        elseif key ==# "\<ESC>" # Close popup
            this.Close()
            return true
        endif

        return popup_filter_menu(winid, key)
    enddef
endclass

defcompile
