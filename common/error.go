package common

type Error struct {
    What  string
    Cause error
}

func (e *Error) Error() string {
    if e.Cause == nil {
        return e.What
    } else {
        return e.What + ":" + e.Cause.Error()
    }
}