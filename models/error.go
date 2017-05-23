package models

type err_type string

const (
	INTERNAL     = err_type("internal")
	UNAUTHORISED = err_type("unauthorised")
	NOT_FOUND    = err_type("not-found")
)

type ModelError struct {
	Err  error
	Type err_type
}

func WrapErr(err error, errorType err_type) *ModelError {
	// we don't want to wrap an error
	// that is nil
	if err == nil {
		return nil
	}

	return &ModelError{
		Err:  err,
		Type: errorType,
	}
}

func (err *ModelError) String() string {
	return err.Err.Error()
}

func (err *ModelError) Error() string {
	return err.Err.Error()
}
