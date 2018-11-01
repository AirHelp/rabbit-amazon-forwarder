package forwarder

const (
	// EmptyMessageError empty error message
	EmptyMessageError = "message is empty"
)

// Client interface to forwarding messages
type Client interface {
	Name() string
	Push(message string) error
	Stop() error
}
