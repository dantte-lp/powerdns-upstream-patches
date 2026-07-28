package zone

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Errors returned by parseMaster. Sentinels rather than formatted strings so a
// caller can distinguish the cases without matching on text.
var (
	ErrMasterEmpty       = errors.New("master address is empty")
	ErrMasterNotAnIP     = errors.New("master address is not a valid IP")
	ErrMasterPortRange   = errors.New("master port is outside 1-65535")
	ErrMasterPortNotANum = errors.New("master port is not a number")
)

// parseMaster accepts the four forms PowerDNS accepts in a zone's masters list:
//
//	192.0.2.1              bare IPv4
//	2001:db8::1            bare IPv6
//	192.0.2.1:53           IPv4 with a port
//	[2001:db8::1]:53       IPv6 with a port, bracketed
//
// The inherited implementation split on ":" and rejected anything with more
// than one colon, which made every bare IPv6 address unusable
// ("more than one colon in <ip>:<port> string", upstream issue #73). The order
// below is what fixes it: a bare address is recognised first, because
// net.SplitHostPort cannot tell "2001:db8::1" from a host-port pair, and only
// then is the value treated as host and port.
func parseMaster(value string) error {
	if value == "" {
		return ErrMasterEmpty
	}

	// A bare IP of either family, including an IPv6 address full of colons.
	if net.ParseIP(value) != nil {
		return nil
	}

	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrMasterNotAnIP, value)
	}

	if net.ParseIP(host) == nil {
		return fmt.Errorf("%w: %q", ErrMasterNotAnIP, host)
	}

	number, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrMasterPortNotANum, port)
	}
	if number < 1 || number > 65535 {
		return fmt.Errorf("%w: %d", ErrMasterPortRange, number)
	}

	return nil
}

// mastersValidator checks every element of the masters set.
//
// It lives in the schema rather than in Create, which is the other half of the
// inherited defect: the SDKv2 resource validated masters on create and not at
// all on update, so a value the provider rejected at create time could still
// reach state by way of an edit. A schema validator runs on both paths, and on
// plan rather than apply.
type mastersValidator struct{}

// MastersValidator returns the validator applied to the masters attribute.
func MastersValidator() validator.Set {
	return mastersValidator{}
}

func (v mastersValidator) Description(_ context.Context) string {
	return "each element must be an IP address, optionally with a port " +
		"(192.0.2.1, 2001:db8::1, 192.0.2.1:53, [2001:db8::1]:53)"
}

func (v mastersValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v mastersValidator) ValidateSet(
	ctx context.Context,
	req validator.SetRequest,
	resp *validator.SetResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var masters []string
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &masters, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, master := range masters {
		if err := parseMaster(master); err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid master address",
				fmt.Sprintf(
					"%q is not usable as a master: %s.\n\n"+
						"Accepted forms are a bare IPv4 or IPv6 address, or either with a "+
						"port — 192.0.2.1, 2001:db8::1, 192.0.2.1:53, [2001:db8::1]:53.",
					master, err,
				),
			)
		}
	}
}
