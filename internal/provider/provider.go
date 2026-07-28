// Package provider implements the PowerDNS provider on
// terraform-plugin-framework. It is served alongside the inherited SDKv2
// provider through a mux server until the migration completes
// (docs/adr/0003-plugin-framework-migration.md).
package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/dantte-lp/terraform-provider-powerdns/powerdns"
)

// Defaults mirrored from the SDKv2 provider's EnvDefaultFunc values. The
// framework has no schema-level default for provider arguments, so they are
// applied in Configure instead — see resolve* below.
const (
	defaultCacheMemSize = "100"
	defaultCacheTTL     = 30
)

// Ensure the interface contract at compile time rather than at plugin start.
var _ provider.Provider = (*powerdnsProvider)(nil)

type powerdnsProvider struct {
	version string
}

// New returns the framework provider factory.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &powerdnsProvider{version: version}
	}
}

// Clients is handed to every resource and data source through Configure.
type Clients struct {
	PDNS     *powerdns.PowerDNSClient
	Recursor *powerdns.RecursorClient
}

// providerModel mirrors the schema below.
type providerModel struct {
	APIKey            types.String `tfsdk:"api_key"`
	ClientCertFile    types.String `tfsdk:"client_cert_file"`
	ClientCertKeyFile types.String `tfsdk:"client_cert_key_file"`
	ServerURL         types.String `tfsdk:"server_url"`
	InsecureHTTPS     types.Bool   `tfsdk:"insecure_https"`
	CACertificate     types.String `tfsdk:"ca_certificate"`
	CacheRequests     types.Bool   `tfsdk:"cache_requests"`
	CacheMemSize      types.String `tfsdk:"cache_mem_size"`
	CacheTTL          types.Int64  `tfsdk:"cache_ttl"`
	RecursorServerURL types.String `tfsdk:"recursor_server_url"`
}

func (p *powerdnsProvider) Metadata(
	_ context.Context,
	_ provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "powerdns"
	resp.Version = p.version
}

// Schema must stay byte-for-byte equivalent to the SDKv2 provider's schema for
// as long as both are muxed: terraform-plugin-mux rejects servers whose
// provider schemas differ, descriptions included. TestProviderSchemasMatch
// asserts this, so a drifting description fails the unit gate rather than the
// plugin handshake.
func (p *powerdnsProvider) Schema(
	_ context.Context,
	_ provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "REST API authentication API key. Can also be set via PDNS_API_KEY.",
			},
			"client_cert_file": schema.StringAttribute{
				Optional:    true,
				Description: "REST API authentication client certificate file path (.crt). Can also be set via PDNS_CLIENT_CERT_FILE.",
			},
			"client_cert_key_file": schema.StringAttribute{
				Optional:    true,
				Description: "REST API authentication client certificate key file path (.key). Can also be set via PDNS_CLIENT_CERT_KEY_FILE.",
			},
			"server_url": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL of the PowerDNS server (e.g., https://pdns.example.com). Can also be set via PDNS_SERVER_URL.",
			},
			"insecure_https": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable verification of the PowerDNS server's TLS certificate. Also via PDNS_INSECURE_HTTPS.",
			},
			"ca_certificate": schema.StringAttribute{
				Optional:    true,
				Description: "Content or path of a Root CA to verify the server certificate. Also via PDNS_CACERT.",
			},
			"cache_requests": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable caching of REST API requests. Also via PDNS_CACHE_REQUESTS.",
			},
			"cache_mem_size": schema.StringAttribute{
				Optional:    true,
				Description: "Cache memory size in MB. Also via PDNS_CACHE_MEM_SIZE.",
			},
			"cache_ttl": schema.Int64Attribute{
				Optional:    true,
				Description: "Cache TTL in seconds. Also via PDNS_CACHE_TTL.",
			},
			"recursor_server_url": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL of the PowerDNS recursor server. Also via PDNS_RECURSOR_SERVER_URL.",
			},
		},
	}
}

func (p *powerdnsProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var data providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := powerdns.Config{
		APIKey:            resolveString(data.APIKey, "PDNS_API_KEY", ""),
		ClientCertFile:    resolveString(data.ClientCertFile, "PDNS_CLIENT_CERT_FILE", ""),
		ClientCertKeyFile: resolveString(data.ClientCertKeyFile, "PDNS_CLIENT_CERT_KEY_FILE", ""),
		ServerURL:         resolveString(data.ServerURL, "PDNS_SERVER_URL", ""),
		RecursorServerURL: resolveString(data.RecursorServerURL, "PDNS_RECURSOR_SERVER_URL", ""),
		InsecureHTTPS:     resolveBool(data.InsecureHTTPS, "PDNS_INSECURE_HTTPS", false),
		CACertificate:     resolveString(data.CACertificate, "PDNS_CACERT", ""),
		CacheEnable:       resolveBool(data.CacheRequests, "PDNS_CACHE_REQUESTS", false),
		CacheMemorySize:   resolveString(data.CacheMemSize, "PDNS_CACHE_MEM_SIZE", defaultCacheMemSize),
		CacheTTL:          resolveInt(data.CacheTTL, "PDNS_CACHE_TTL", defaultCacheTTL),
	}

	if config.ServerURL == "" {
		resp.Diagnostics.AddAttributeError(
			pathServerURL(),
			"Missing PowerDNS server URL",
			"The provider needs the base URL of the PowerDNS Authoritative API. "+
				"Set the server_url argument, or the PDNS_SERVER_URL environment variable.",
		)
		return
	}

	tflog.Debug(ctx, "initialising the PowerDNS clients", map[string]any{
		"server_url":       config.ServerURL,
		"recursor_present": config.RecursorServerURL != "",
	})

	pdnsClient, recursorClient, err := config.Clients(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create the PowerDNS client",
			err.Error(),
		)
		return
	}

	clients := &Clients{PDNS: pdnsClient, Recursor: recursorClient}
	resp.DataSourceData = clients
	resp.ResourceData = clients
}

// Resources is empty until the first resource is ported. A type name may be
// served by exactly one half of the mux, so a resource appears here in the same
// commit that removes it from the SDKv2 provider.
func (p *powerdnsProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

func (p *powerdnsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// resolveString applies the SDKv2 EnvDefaultFunc precedence: an explicit
// configuration value wins, then the environment, then the default.
func resolveString(v types.String, envVar, fallback string) string {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueString()
	}
	if env, ok := os.LookupEnv(envVar); ok && env != "" {
		return env
	}
	return fallback
}

func resolveBool(v types.Bool, envVar string, fallback bool) bool {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueBool()
	}
	if env, ok := os.LookupEnv(envVar); ok && env != "" {
		// An unparseable value falls back rather than failing: this mirrors
		// SDKv2's EnvDefaultFunc, which ignores what it cannot read.
		if parsed, err := strconv.ParseBool(env); err == nil {
			return parsed
		}
	}
	return fallback
}

func resolveInt(v types.Int64, envVar string, fallback int) int {
	if !v.IsNull() && !v.IsUnknown() {
		return int(v.ValueInt64())
	}
	if env, ok := os.LookupEnv(envVar); ok && env != "" {
		if parsed, err := strconv.Atoi(env); err == nil {
			return parsed
		}
	}
	return fallback
}
