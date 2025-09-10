### This repo

This is a terminal user interface (TUI) using:

* https://github.com/charmbracelet/bubbletea
* https://github.com/charmbracelet/lipgloss
* https://github.com/charmbracelet/bubbles
* https://github.com/superfly/contextwindow

The interface allows users to chat with an LLM and manages multiple 
LLM sessions ("contexts") in a SQLite database, itself managed by 
the "contextwindow" library.

The UI itself is arranged in a tree of components starting from 
"rootWindow". "Controllers" encapsulate logic that don't have display
properties. Actions and state changes in the UI occur by sending
messages; messages are generally sent thru "tea.Cmd", which is just
a "fn () tea.Msg" --- tea.Cmds run in their own goroutines. 

Use "tea.Sequence" to execute several commands serially, and "tea.Batch"
to execute them concurrently.

### Basic stuff
* never write comments
* think hard
* If unsure about a best practice or implementation detail, say so instead of guessing.

### Basic Go Stuff
* use github.com/stretchr/testify/assert to write unit tests (assert.Equal, assert.NoError, assert.Contains, etc)
* don't write tests for trivial things unlikely to fail
* don't write tests for TUI programs
* keep line length to around 80 characters; prefer vertical to horizontal
* one line per comma-delimited struct or array field.
* Use the standard library's net/http package for HTTP API development
* Use structured logging (slog) when logging
* use simple fmt.Errorf() error wrapping, include only information in the format string that the CALLER would not have
* Prefer small helper functions to repeated code
* Use named functions for goroutines unless the goroutine only spans a couple lines
* "defer" unlocks of locks when possible rather than bracketing critical sections.
* don't create "else" arms on error checking conditions; keep code straight-line
* be familiar with RESTful API design principles, best practices, and Go idioms
* use fmt.Fprintf in preference to Write([]byte(stringValue))
* use StringBuilder or something like it in preference to append/join
