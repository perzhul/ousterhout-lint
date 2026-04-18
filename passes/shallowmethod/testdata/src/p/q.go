package p

// Interface declared in a separate file of the same package. shallowmethod
// should still recognize OtherAdapter as an implementation and suppress.
type Closer interface {
	Close(reason string) error
}

var globalCloser Closer

type OtherAdapter struct{}

// OK: implements Closer (declared in p.go peer file).
func (o *OtherAdapter) Close(reason string) error {
	return globalCloser.Close(reason)
}
