package main

// A general error struct that just contains a string s.
type errorString struct {
	s string
}

func (e errorString) Error() string {
	return e.s
}
