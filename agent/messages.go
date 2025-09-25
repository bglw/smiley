package agent

type Message interface{}

type ToolCallMsg struct {
	Name     string
	Args     string
	Complete bool
	Err      error
	Size     int
	Msg      string
}

type ModelResponseMsg struct {
	Response string
}

type TokenUsageMsg struct {
	Usage float64
}

type WorkingMsg struct {
	Working bool
}

type ErrorMsg struct {
	Err error
	Msg string
}
