package tunnel

import (
	"net"
	"sync"
	"testing"
)

func TestAllocatePort(t *testing.T) {
	ap, err := AllocatePort(15000)
	if err != nil {
		t.Fatalf("AllocatePort(15000) error: %v", err)
	}
	defer ap.Close()

	if ap.Port < 15000 {
		t.Errorf("AllocatePort() returned port %d, want >= 15000", ap.Port)
	}

	// Port should be busy while listener is open
	if !PortOpen(ap.Port) {
		t.Errorf("Allocated port %d should be open (listening)", ap.Port)
	}
}

func TestAllocatePortRespectsBusyPort(t *testing.T) {
	// Listen on a port first
	l, err := net.Listen("tcp", "127.0.0.1:15001")
	if err != nil {
		t.Fatalf("net.Listen error: %v", err)
	}
	defer l.Close()

	// AllocatePort should skip 15001 and give us 15002
	ap, err := AllocatePort(15001)
	if err != nil {
		t.Fatalf("AllocatePort(15001) error: %v", err)
	}
	defer ap.Close()

	if ap.Port == 15001 {
		t.Errorf("AllocatePort() returned busy port 15001, expected 15002 or higher")
	}
}

func TestAllocatePortReleasesOnClose(t *testing.T) {
	ap, err := AllocatePort(15010)
	if err != nil {
		t.Fatalf("AllocatePort(15010) error: %v", err)
	}

	port := ap.Port

	// Close the listener
	if err := ap.Close(); err != nil {
		t.Fatalf("ap.Close() error: %v", err)
	}

	// Port should now be free
	if PortOpen(port) {
		t.Errorf("Port %d should be free after Close()", port)
	}
}

func TestAllocatePortCloseIdempotent(t *testing.T) {
	ap, err := AllocatePort(15020)
	if err != nil {
		t.Fatalf("AllocatePort(15020) error: %v", err)
	}

	// Close twice should not panic or error
	if err := ap.Close(); err != nil {
		t.Errorf("first Close() error: %v", err)
	}
	if err := ap.Close(); err != nil {
		t.Errorf("second Close() error: %v", err)
	}
}

func TestConcurrentAllocationNoDuplicatePorts(t *testing.T) {
	const goroutines = 20
	startPort := 15100

	// Collect all allocated ports in a slice, close all after allocation completes.
	// This avoids false duplicates caused by early-release-then-reallocate.
	var mu sync.Mutex
	type apHolder struct {
		ap   *AllocatedPort
		port int
	}
	var holders []apHolder
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ap, err := AllocatePort(startPort)
			if err != nil {
				t.Errorf("AllocatePort error: %v", err)
				return
			}

			mu.Lock()
			holders = append(holders, apHolder{ap: ap, port: ap.Port})
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Now check for duplicates while all ports are still held
	allocated := map[int]bool{}
	for _, h := range holders {
		if allocated[h.port] {
			t.Errorf("Duplicate port allocated: %d", h.port)
		}
		allocated[h.port] = true
	}

	// Clean up all listeners
	for _, h := range holders {
		h.ap.Close()
	}

	// Verify we allocated as many unique ports as goroutines
	if len(allocated) != goroutines {
		t.Errorf("Allocated %d unique ports, want %d", len(allocated), goroutines)
	}
}

func TestAllocatePortLargeOffset(t *testing.T) {
	// Allocate ports starting from a high offset to avoid conflicts
	ap, err := AllocatePort(30000)
	if err != nil {
		t.Fatalf("AllocatePort(30000) error: %v", err)
	}
	defer ap.Close()

	if ap.Port < 30000 || ap.Port > 31000 {
		t.Errorf("AllocatePort(30000) returned unexpected port %d", ap.Port)
	}
}
