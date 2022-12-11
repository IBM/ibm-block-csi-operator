package errors

import (
	"fmt"

	"github.com/IBM/volume-group-operator/pkg/messages"
)

type MatchingLabelsAndLabelSelectorError struct {
	ErrorMessage string
}

func (e *MatchingLabelsAndLabelSelectorError) Error() string {
	return fmt.Sprintf(messages.MatchingLabelsAndLabelSelectorFailed, e.ErrorMessage)
}
