package zone

import (
	"errors"
	"testing"
)

// TestParseMaster covers the four accepted forms and the ways a value can be
// wrong. The IPv6 cases are the reason this function exists: the inherited
// implementation split on ":" and rejected anything with more than one colon,
// so every bare IPv6 address failed (upstream issue #73).
func TestParseMaster(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr error
	}{
		// Accepted.
		{name: "bare ipv4", value: "192.0.2.1"},
		{name: "bare ipv6", value: "2001:db8::1"},
		{name: "bare ipv6 full", value: "fd92:81e1:e314:ea7b:0000:1234:5678:60ab"},
		{name: "bare ipv6 loopback", value: "::1"},
		{name: "ipv4 with port", value: "192.0.2.1:53"},
		{name: "ipv6 with port bracketed", value: "[2001:db8::1]:53"},
		{name: "ipv4 highest port", value: "192.0.2.1:65535"},
		{name: "ipv4 lowest port", value: "192.0.2.1:1"},

		// Rejected.
		{name: "empty", value: "", wantErr: ErrMasterEmpty},
		{name: "hostname", value: "ns1.example.com", wantErr: ErrMasterNotAnIP},
		{name: "hostname with port", value: "ns1.example.com:53", wantErr: ErrMasterNotAnIP},
		{name: "ipv6 with port unbracketed", value: "2001:db8::1:53", wantErr: nil},
		{name: "port zero", value: "192.0.2.1:0", wantErr: ErrMasterPortRange},
		{name: "port too high", value: "192.0.2.1:65536", wantErr: ErrMasterPortRange},
		{name: "port not a number", value: "192.0.2.1:dns", wantErr: ErrMasterPortNotANum},
		{name: "garbage", value: "not an address", wantErr: ErrMasterNotAnIP},
		{name: "ipv4 out of range", value: "999.0.2.1", wantErr: ErrMasterNotAnIP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := parseMaster(tt.value)

			switch {
			case tt.wantErr == nil && err != nil:
				t.Fatalf("parseMaster(%q) = %v, want nil", tt.value, err)
			case tt.wantErr != nil && err == nil:
				t.Fatalf("parseMaster(%q) = nil, want %v", tt.value, tt.wantErr)
			case tt.wantErr != nil && !errors.Is(err, tt.wantErr):
				t.Fatalf("parseMaster(%q) = %v, want %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

// TestParseMaster_InheritedDefect is the regression guard for the exact input
// from upstream issue #73. It is separate from the table above so that a future
// edit to the table cannot quietly drop the case the fix exists for.
func TestParseMaster_InheritedDefect(t *testing.T) {
	t.Parallel()

	// The configuration in the issue: an IPv6 master alongside an IPv4 one.
	reported := []string{
		"fd92:81e1:e314:ea7b:0000:1234:5678:60ab",
		"192.168.123.45",
	}

	for _, master := range reported {
		if err := parseMaster(master); err != nil {
			t.Errorf("parseMaster(%q) = %v; upstream #73 must stay fixed", master, err)
		}
	}
}
