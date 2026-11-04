package zk

import (
	"bytes"
	"errors"
	"log"
	"net"
	"strings"
	"testing"
	"time"
)

func TestConnectDNSLogging(t *testing.T) {
	for _, info := range []bool{true, false} {
		name := "enabled"
		if !info {
			name = "disabled"
		}
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			conn, events, err := Connect([]string{"127.0.0.1"}, time.Second,
				WithLogger(log.New(&output, "", 0)), WithLogInfo(info),
				WithDialer(func(string, string, time.Duration) (net.Conn, error) {
					return nil, errors.New("network disabled in logging test")
				}))
			if err != nil {
				t.Fatal(err)
			}
			conn.Close()
			for range events {
			}
			for _, message := range []string{
				`Connect: servers=["127.0.0.1"]`,
				`Resolving host "127.0.0.1"`,
				`Host lookup result: host="127.0.0.1" ips=["127.0.0.1"] err=<nil>`,
			} {
				count := strings.Count(output.String(), message)
				want := 0
				if info {
					want = 1
				}
				if count != want {
					t.Errorf("message %q appeared %d times, want %d; logs: %s", message, count, want, output.String())
				}
			}
		})
	}
}

func TestConnectDNSLookupErrorLogging(t *testing.T) {
	var output bytes.Buffer
	// An empty hostname fails resolution locally without querying a DNS server.
	conn, _, err := Connect([]string{":2181"}, time.Second,
		WithLogger(log.New(&output, "", 0)), WithLogInfo(false))
	if err == nil {
		conn.Close()
		t.Fatal("expected an error for an empty hostname")
	}
	if !strings.Contains(output.String(), `Host lookup result: host="" ips=[] err=`) {
		t.Fatalf("missing lookup failure with informational logging disabled: %s", output.String())
	}
	if strings.Contains(output.String(), "Resolving host") || strings.Contains(output.String(), "Connect: servers=") {
		t.Fatalf("informational messages should be disabled: %s", output.String())
	}
}
