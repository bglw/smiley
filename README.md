# Smiley: A General Sort of Agent

<img src="https://drive.usercontent.google.com/download?id=1ciE3NUiyCAJeIH0nr8yWVP_spcJL1l-4" width="400" />

This is a minimal TUI that manages multiple SQLite-backed LLM conversations
(contexts) with configurable tool calls. I'm pretty sure you can do all of 
this with Simon Willison's LLM tool already. This is basically example code
for [a context window library](https://www.github.com/superfly/contextwindow).

Demands `OPENAI_API_KEY` set in the environment. Sorry!

When run for the first time, it's going to want to create a `~/.ctxagent`,
which is where it'll stick its database by default (you can tell it where
to stick it). 

Tools are programs, shell scripts, whatever, defined by `~/.ctxagent/tools.toml`,
or whatever file you point `-tools foo.toml` at.

Once running: 

* `C-j` to send text to the LLM. This is annoying but it's what Gemini
  does too.
  
* `C-h` to see a history view of all previous context sessions.

* `C-l` to go back to the LLM conversation.

## Tool Configuration

Claude wrote everything that follows (but not that much of the code! don't
@ me!). I think this is probably true though.

The application supports configurable tool calls that allow the LLM to execute local commands and scripts. Tools are configured via a TOML file located at `~/.ctxagent/tools.toml`.

### Tool Configuration Format

Each tool is defined with:

- **name**: Unique identifier for the tool
- **description**: Multi-paragraph description that teaches the LLM when and how to use the tool
- **command**: Shell command or script to execute when the tool is called
- **parameters**: Optional parameters the tool accepts

### Parameter Configuration

Parameters support:

- **type**: `"string"` (default) or `"number"`
- **description**: Explanation of what the parameter does
- **required**: `true` or `false` (default)

### Example Configuration

```toml
[[tool]]
name = "get_logs"
description = """
Retrieve logs from OpenSearch using KQL queries. Use this when you need to investigate
system issues, errors, or analyze application behavior. The query should be in KQL format.
"""
command = "llmtool run-tool get-logs"

[tool.parameters]
query = { description = "KQL query to filter logs", required = true }
limit = { type = "number", description = "Maximum number of results to return" }
```

### Setup

1. Copy `tools.toml.example` to `~/.ctxagent/tools.toml`
2. Customize the configuration for your specific tools
3. Ensure your tool scripts/commands are executable and in your PATH

The application will automatically load the tool configuration on startup. If no configuration file exists, the application will run without tool support.

## Code shape

Pretty vanilla Bubbletea app. `rootWindow` is the main model; everything
else is a tree of models hanging off it, `rootWindow.top` wraps whatever
the current view is. `controllers` are model-like
without any view code, most of the actual agentry is in the ContextWindow
library. The Bubbletea people say not to use `tea.Cmd` to send messages
between components, but I don't see a cleaner way to do it, so that's what
I do. Suck it, Bubbletea people! (The library is great, thxu).






