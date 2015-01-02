package domain

type AppError struct {
	Error   error
	Message string
	Code    int
}

type LoadData struct {
	ID      string // crane load ID
	Name    string
	Tag     string
	ImageId string // docker image ID
}

type ExecData struct {
	LoadId string
	Cmd    []string
}

type ExecResult struct {
	ExitCode int
	Out      string
}
