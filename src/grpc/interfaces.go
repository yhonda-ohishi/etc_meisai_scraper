package grpc

import "net"

// NetListener is an interface for network listening (mockable)
type NetListener interface {
	Listen(network, address string) (net.Listener, error)
}

// DefaultNetListener implements NetListener using standard net package
type DefaultNetListener struct{}

// Listen creates a network listener
func (d *DefaultNetListener) Listen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}