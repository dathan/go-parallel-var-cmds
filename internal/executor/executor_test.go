package executor

import "testing"

func TestParseHosts(t *testing.T) {
    input := "host1, host2;host3\nhost4\n\nhost5"
    expected := []string{"host1", "host2", "host3", "host4", "host5"}
    hosts := ParseHosts(input)
    if len(hosts) != len(expected) {
        t.Fatalf("expected %d hosts, got %d", len(expected), len(hosts))
    }
    for i := range hosts {
        if hosts[i] != expected[i] {
            t.Errorf("expected host %d to be %s, got %s", i, expected[i], hosts[i])
        }
    }
}
