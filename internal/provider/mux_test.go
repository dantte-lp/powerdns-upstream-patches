package provider_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/dantte-lp/terraform-provider-powerdns/internal/provider"
	"github.com/dantte-lp/terraform-provider-powerdns/powerdns"
)

// TestMuxServer_ProviderSchemasMatch is the load-bearing test of the migration
// mechanism. terraform-plugin-mux refuses to serve providers whose provider
// schemas differ — including attribute descriptions — so the framework schema
// in internal/provider must track the SDKv2 one exactly for as long as both are
// muxed.
//
// Without this test the mismatch surfaces at plugin handshake, as a Terraform
// error during `terraform plan` in a consumer's project. With it, a drifting
// description fails `task test`.
func TestMuxServer_ProviderSchemasMatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	upgraded, err := tf5to6server.UpgradeServer(ctx, powerdns.Provider().GRPCProvider)
	if err != nil {
		t.Fatalf("upgrading the SDKv2 provider to protocol 6: %v", err)
	}

	providers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(provider.New("test")()),
		func() tfprotov6.ProviderServer { return upgraded },
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		t.Fatalf("creating the mux server: %v", err)
	}

	// NewMuxServer defers the schema comparison until the schemas are actually
	// fetched, so constructing the server is not on its own proof of anything.
	// GetProviderSchema is where a mismatch is reported.
	resp, err := muxServer.ProviderServer().GetProviderSchema(
		ctx, &tfprotov6.GetProviderSchemaRequest{},
	)
	if err != nil {
		t.Fatalf("GetProviderSchema: %v", err)
	}

	for _, diag := range resp.Diagnostics {
		if diag.Severity == tfprotov6.DiagnosticSeverityError {
			t.Errorf("mux server reported a schema error: %s — %s",
				diag.Summary, diag.Detail)
		}
	}
}

// TestMuxServer_NoDuplicateTypeNames guards the one-way rule from ADR 0003: a
// resource or data source is served by exactly one half of the mux. Porting a
// resource without removing it from the SDKv2 provider produces a duplicate,
// which the mux server rejects — this test names the offending type instead of
// leaving the reader with a handshake failure.
func TestMuxServer_NoDuplicateTypeNames(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	sdkProvider := powerdns.Provider()
	frameworkResources := provider.New("test")()

	// The framework half advertises its types through GetProviderSchema; ask
	// the server rather than reaching into the provider implementation.
	server := providerserver.NewProtocol6(frameworkResources)()
	resp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("GetProviderSchema on the framework provider: %v", err)
	}

	for name := range resp.ResourceSchemas {
		if _, duplicated := sdkProvider.ResourcesMap[name]; duplicated {
			t.Errorf("resource %q is served by both halves of the mux; "+
				"remove it from powerdns.Provider() in the same commit that "+
				"adds it to internal/provider (ADR 0003)", name)
		}
	}

	for name := range resp.DataSourceSchemas {
		if _, duplicated := sdkProvider.DataSourcesMap[name]; duplicated {
			t.Errorf("data source %q is served by both halves of the mux; "+
				"remove it from powerdns.Provider() in the same commit that "+
				"adds it to internal/provider (ADR 0003)", name)
		}
	}
}
