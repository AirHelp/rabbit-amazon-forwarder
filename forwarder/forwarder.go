package forwarder

// Client interface to forwarding messages
type Client interface {
	Name() string
	Push(message string) error
}
