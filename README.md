# Smiley: A General Sort of Agent

<img src="https://drive.usercontent.google.com/download?id=1ciE3NUiyCAJeIH0nr8yWVP_spcJL1l-4" width="400" />

This is a minimal TUI that manages multiple SQLite-backed LLM conversations
(contexts) with configurable tool calls. 

You can probably do all this with [simonw's LLM tool](https://github.com/simonw/llm). 
So why build this?

1. Everyone should build an agent; it's eye-opening (and surprisingly easy).

2. I'm experimenting with context patterns; handling things that blow out 
   context windows (like a large log query). That requires control over 
   the agent loop itself.
   
3. MCP is annoying and this agent doesn't need it.

This is basically example code for [a context window
library](https://www.github.com/superfly/contextwindow).

## Running

Set `OPENAI_API_KEY`, build, and run. Will want to own `$HOME/.ctxagent`, 
where it'll park a `contextwindow.db` SQLite.

When run for the first time, it's going to want to create a `~/.ctxagent`,
which is where it'll stick its database by default (you can tell it where
to stick it). 

## The Interface

Is inscrutable (I know how to drive it so I don't have to care). Basics:

* `C-j` to send text to the LLM. This is annoying but it's what Gemini
  does too.
  
* `C-h` to see a history view of all previous context sessions.

* `C-l` to go back to the LLM conversation.

* `PgUp`, `PgDown`, `End`, and the mouse wheel should all scroll the LLM
  conversation.
  
* `/dump <filename.md>` in the TUI will give you a Markdown dump of the
  conversation.
  
### Flags

* `-system <prompt.md>`: set a system prompt; we'll also read it from
  `~/.ctxagent/system.md` if it's there.
  
* `-tools <tools.toml>`: load a tool configuration; we'll also read it
  from `~/.ctxagent/tools.toml` if it's there.
  
* `-name <name>`: name the conversation.

* `-fork <name>`: fork an existing conversation (copy and resume it).
  
Running the agent with the name of an existing conversation resumes it.

## Tool Configuration

**Do not give this code tools that can make nonreversible changes to your
environment.**

**Beware: any sensitive information you give this access to will be logged
in your `contextwindow.db`.**

```toml
[[tool]]
name = "ping"
description = """
Test network connectivity.
"""
command = "ping -c 10 -i 0.1 {host}"

[tool.parameters]
host = { description = "A hostname or IP address.", required = true }

[[tool]]
name = "todo"
builtin = true

[[tool]]
info_command = "../agent-tools/agent-tools slack info"
```

A tool is normally compromised of:

- **name**: Unique identifier for the tool
- **description**: Multi-paragraph description that teaches the LLM when and how to use the tool
- **command**: Shell command or script to execute when the tool is called
- **parameters**: Optional parameters the tool accepts

Each `parameter` is:

- **type**: `"string"`or `"number"`
- **description**: Explanation of what the parameter does
- **required**: `true` or `false` (default)

Two other ways to define a tool:

* Specify `info_command` and we'll read the TOML for the tool definition from
  the output of that command, which is useful if you want to bundle tools
  and their configurations in a single shell script or binary.
  
* Specify `builtin` and we'll run a builtin command (TK.)

**Annoying note**: Right now, the output of that command needs to be the TOML for 
a *single tool definition* --- don't include `[[tool]]` at the top, and it's 
`[parameters]` and not `[tool.parameters]`. This is dumb but it's the way it is.


