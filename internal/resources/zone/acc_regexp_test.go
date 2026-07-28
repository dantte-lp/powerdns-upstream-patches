//go:build acceptance

package zone_test

import "regexp"

// The diagnostic texts these match are part of the contract with the user, so
// they live in one place rather than inline in each test.
func regexpInvalidMaster() *regexp.Regexp {
	return regexp.MustCompile(`(?s)Invalid master address`)
}

func regexpMastersOnlySlave() *regexp.Regexp {
	return regexp.MustCompile(`(?s)masters is only valid for a Slave zone`)
}
