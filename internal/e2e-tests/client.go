package e2etests

type Client interface {
	Close()
	Do(command string) (interface{}, error)
	Dof(command string, args ...interface{}) (interface{}, error)
	MustDo(command string) interface{}
}
