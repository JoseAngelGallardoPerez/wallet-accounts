package balance

// Error defines rate error
type Error string

// Error returns error message
func (e Error) Error() string {
	return string(e)
}
