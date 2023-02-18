package purego

// Error represents an error value returned from purego
type Error struct {
	s string
}

func (e Error) Error() string {
	return e.s
}
