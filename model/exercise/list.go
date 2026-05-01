package exercise

type ListFilter struct {
	NameRegex *string
	Category  *string
	UserID    string
	Offset    int64
	Limit     int64
}
