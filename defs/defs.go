package defs

/*
Definitions shared among multiple go programs:
- gcr evaluator
- test code
- codeeval.dev backend
*/

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type EvalRequest struct {
	Files    []*File `json:"files"`
	Language string  `json:"lang"`
	// TODO: maybe replace by codeeval.yml
	Command string `json:"command,omitempty"`
	Stdin   string `json:"stdin,omitempty"`
}

type EvalFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type EvalResponse struct {
	// TODO: also add combined output
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	// either compilation or execution error
	Error string      `json:"error,omitempty"`
	Files []*EvalFile `json:"files,omitempty"`
	// how long did it take to execute
	DurationMS float64 `json:"durationms"`
	// true if this is a cached response
	// used by front-end to show a message
	IsCached bool `json:"cached"`
}
