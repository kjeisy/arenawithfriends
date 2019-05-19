package lobby

// Errors
const (
	ErrSessionNotFound         Error = "session not found"
	ErrPlayerAlreadyRegistered Error = "player already registered"
)

// Error describes websocket-related errors
type Error string

func (e Error) Error() string {
	return string(e)
}
