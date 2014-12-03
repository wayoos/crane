package domain

type AppError struct {
	Error   error
	Message string
	Code    int
}

type LoadData struct {
	ID   string
	Name string
	Tag  string
}

type ExecData struct {
	LoadId string
	Cmd    []string
}

type ExecResult struct {
	ExitCode int
	Out      string
}
