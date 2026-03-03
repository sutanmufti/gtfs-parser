package gtfsparser

import "fmt"

type ValidationError struct {
	File    string
	Field   string
	ID      string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("[%s] id=%q field=%q: %s", e.File, e.ID, e.Field, e.Message)
}
