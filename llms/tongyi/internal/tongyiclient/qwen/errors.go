package qwenclient

type WrapMessageError struct {
	Message string
	Cause   error
}

func (e *WrapMessageError) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return e.Message + ": " + e.Cause.Error()
}
