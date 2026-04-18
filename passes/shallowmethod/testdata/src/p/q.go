package p

// Interface in a peer file — shallowmethod must still recognize
// OtherAdapter as an implementation.
type Closer interface {
	Close(reason string) error
}

var globalCloser Closer

type OtherAdapter struct{}

func (o *OtherAdapter) Close(reason string) error {
	return globalCloser.Close(reason)
}
