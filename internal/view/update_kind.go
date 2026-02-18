package view

const (
	FullUpdate UpdateKind = iota
	PartialUpdate
)

type UpdateKind int

func (k UpdateKind) String() string {
	if k == FullUpdate {
		return "full"
	}

	return "partial"
}
