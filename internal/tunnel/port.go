package tunnel

import (
	"fmt"
	"net"
)

// AllocatedPort represents a port that has been atomically allocated
// via net.Listen. The port is held exclusively while the listener is open.
type AllocatedPort struct {
	Port     int
	listener net.Listener
}

// AllocatePort allocates a TCP port atomically using net.Listen.
// It tries ports starting from 'from' and returns the first available one.
// Call Close on the returned AllocatedPort to release the port.
func AllocatePort(from int) (*AllocatedPort, error) {
	p := from
	maxTries := 1000
	for p < from+maxTries {
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
		if err == nil {
			return &AllocatedPort{Port: p, listener: listener}, nil
		}
		p++
	}
	return nil, fmt.Errorf("no free port found after %d tries starting from %d", maxTries, from)
}

// Close releases the allocated port by closing the underlying listener.
// It is safe to call Close multiple times.
func (ap *AllocatedPort) Close() error {
	if ap.listener != nil {
		err := ap.listener.Close()
		ap.listener = nil
		return err
	}
	return nil
}
