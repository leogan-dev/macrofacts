package foods

type NotFoundError struct {
	Msg string
}

func (e NotFoundError) Error() string { return e.Msg }

func IsNotFound(err error) bool {
	_, ok := err.(NotFoundError)
	return ok
}
