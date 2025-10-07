package constants

type FieldStatusString string

const (
	AvailableStatus FieldStatusString = "available"
	BookedStatus    FieldStatusString = "booked"
)

func (f FieldStatusString) String() string {
	return string(f)
}
