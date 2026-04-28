package main

import (
	"encoding/json"
	"net"
	"testing"

	"launchpal/internal/privhelper"
)

// pairedInternal is a copy of the paired helper in the privhelper tests,
// exposed here so admin_mode_test can drive a privhelper.Client via
// net.Pipe without crossing package boundaries.
func pairedInternal(t *testing.T, respond func(req privhelper.Request) privhelper.Response) (*privhelper.Client, func()) {
	t.Helper()
	clientEnd, serverEnd := net.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = serverEnd.Close() }()
		decoder := json.NewDecoder(serverEnd)
		encoder := json.NewEncoder(serverEnd)
		for {
			var req privhelper.Request
			if err := decoder.Decode(&req); err != nil {
				return
			}
			if err := encoder.Encode(respond(req)); err != nil {
				return
			}
		}
	}()
	client := privhelper.NewClient(privhelper.ClientOptions{Conn: clientEnd})
	cleanup := func() {
		_ = client.Close()
		_ = serverEnd.Close()
		<-done
	}
	return client, cleanup
}
