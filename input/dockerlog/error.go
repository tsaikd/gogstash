package inputdockerlog

type ErrorContainerLogLoopRunning struct {
	ID string
}

func (t *ErrorContainerLogLoopRunning) Error() string {
	return "container log loop running: " + t.ID
}
