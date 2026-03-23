package queues

import "tbunny/internal/view"

// labelAndValue is an alias for the shared view.LabelAndValue type.
type labelAndValue = view.LabelAndValue

// createLabelAndValue delegates to view.NewLabelAndValue.
func createLabelAndValue(labelText string) *labelAndValue {
	return view.NewLabelAndValue(labelText)
}
