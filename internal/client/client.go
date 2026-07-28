// Package client holds the bundle of PowerDNS API clients that the provider
// hands to every resource and data source.
//
// It is a package of its own rather than a type in internal/provider because
// both the provider and every resource need it, and a resource importing the
// provider package would be an import cycle.
package client

import "github.com/dantte-lp/terraform-provider-powerdns/powerdns"

// Bundle carries the configured API clients.
//
// Recursor is nil when recursor_server_url is not configured; a resource that
// needs it must say so with a diagnostic rather than dereferencing it.
type Bundle struct {
	PDNS     *powerdns.PowerDNSClient
	Recursor *powerdns.RecursorClient
}
