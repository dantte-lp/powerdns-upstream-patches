package provider

import "github.com/hashicorp/terraform-plugin-framework/path"

// pathServerURL is the attribute path used when reporting a missing server URL.
func pathServerURL() path.Path {
	return path.Root("server_url")
}
