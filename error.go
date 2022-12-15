package nevent

import (
	"fmt"
)

type AskError struct {
	AnswerError string
}

func (e *AskError) Error() string {
	return fmt.Sprintf("nevent ask err:%s", e.AnswerError)
}

func (e *AskError) GetAnswerError() string {
	return e.AnswerError
}
