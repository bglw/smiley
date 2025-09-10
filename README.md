# TUI Chat Application

A terminal user interface (TUI) for chatting with LLMs, built using Bubble Tea.

## Tool Configuration

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
