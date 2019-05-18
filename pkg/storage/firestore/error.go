package firestore

// Error messages
const (
	ErrPlatformNotSupported = "this platform doesn't support firestore"
)

// Error contains firestore-related errors
type Error string

func (e Error) Error() string {
	return string(e)
}
