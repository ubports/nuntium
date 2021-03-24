package mms

import "fmt"

type ErrorDecodeShortData struct {
	Length, Expected int
}

func (e ErrorDecodeShortData) Error() string {
	return fmt.Sprintf("expexted offset after decoding out of range [%d] with data length %d ", e.Expected, e.Length)
}

type ErrorDecodeUnknownExpiryToken uint64

func (e ErrorDecodeUnknownExpiryToken) Error() string {
	return fmt.Sprintf("Unknown expiry token: %x", e)
}

type ErrorDecodeInconsistentOffset struct {
	Offset, Expected int
}

func (e ErrorDecodeInconsistentOffset) Error() string {
	return fmt.Sprintf("Decoder offset after read [%d] is other than expected [%d]", e.Offset, e.Expected)
}

const (
	DebugErrorActivateContext = "error-activate-context"
	DebugErrorGetProxy        = "error-get-proxy"
	DebugErrorDownloadStorage = "error-storage"
)

type ForcedDebugError string

func (e ForcedDebugError) Error() string { return fmt.Sprintf("forced debug error: %s", string(e)) }
