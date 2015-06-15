package inputdockerstats

type ErrorContainerLoopRunning struct {
	ID string
}

func (t *ErrorContainerLoopRunning) Error() string {
	return "container log loop running: " + t.ID
}
