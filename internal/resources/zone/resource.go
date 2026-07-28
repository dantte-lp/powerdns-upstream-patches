// Package zone implements powerdns_zone on terraform-plugin-framework.
//
// Ported from the inherited SDKv2 resource. Behaviour is preserved except for
// two defects fixed in the move: IPv6 addresses in masters are now parsed
// correctly, and masters is validated on every path rather than only on create
// (upstream issue #73; capability map D-01 and D-02).
package zone

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/dantte-lp/terraform-provider-powerdns/internal/client"
	"github.com/dantte-lp/terraform-provider-powerdns/powerdns"
)

const (
	kindSlave       = "Slave"
	defaultAccount  = "admin"
	maxAccountChars = 40
)

var (
	_ resource.Resource                   = (*zoneResource)(nil)
	_ resource.ResourceWithConfigure      = (*zoneResource)(nil)
	_ resource.ResourceWithImportState    = (*zoneResource)(nil)
	_ resource.ResourceWithValidateConfig = (*zoneResource)(nil)
)

type zoneResource struct {
	clients *client.Bundle
}

// NewResource returns the powerdns_zone resource.
func NewResource() resource.Resource {
	return &zoneResource{}
}

type zoneModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Kind       types.String `tfsdk:"kind"`
	Account    types.String `tfsdk:"account"`
	Catalog    types.String `tfsdk:"catalog"`
	Masters    types.Set    `tfsdk:"masters"`
	SoaEditAPI types.String `tfsdk:"soa_edit_api"`
}

func (r *zoneResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (r *zoneResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a PowerDNS zone.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Zone identifier, which PowerDNS reports as the canonical zone name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Fully qualified zone name, including the trailing dot.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{ZoneNameValidator()},
			},
			"kind": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Zone kind: `Native`, `Master`, `Slave`, `Producer` or " +
					"`Consumer`. Compared case-insensitively, matching PowerDNS.",
				PlanModifiers: []planmodifier.String{
					caseInsensitive(),
				},
			},
			"account": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(defaultAccount),
				MarkdownDescription: "Account name recorded against the zone.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(maxAccountChars),
				},
			},
			"catalog": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Catalog zone this zone belongs to. Requires PowerDNS 4.7 or later.",
				Validators:          []validator.String{ZoneNameValidator()},
			},
			"masters": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				MarkdownDescription: "Master servers for a `Slave` zone. Each element is an IP " +
					"address, optionally with a port: `192.0.2.1`, `2001:db8::1`, " +
					"`192.0.2.1:53`, `[2001:db8::1]:53`.",
				Validators: []validator.Set{MastersValidator()},
			},
			"soa_edit_api": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "SOA-EDIT-API metadata value applied to the zone.",
			},
		},
	}
}

// ValidateConfig enforces the cross-attribute rule the SDKv2 resource checked
// inside Create: masters is only meaningful for a Slave zone. Doing it here
// moves the failure from apply to plan.
func (r *zoneResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data zoneModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Masters.IsNull() || len(data.Masters.Elements()) == 0 {
		return
	}
	// An unknown kind cannot be checked now; Create will catch it.
	if data.Kind.IsUnknown() || data.Kind.IsNull() {
		return
	}

	if !strings.EqualFold(data.Kind.ValueString(), kindSlave) {
		resp.Diagnostics.AddAttributeError(
			path.Root("masters"),
			"masters is only valid for a Slave zone",
			fmt.Sprintf(
				"kind is %q. PowerDNS only replicates from masters for a Slave zone, so "+
					"masters would be silently ignored. Set kind = \"Slave\", or remove masters.",
				data.Kind.ValueString(),
			),
		)
	}
}

func (r *zoneResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*client.Bundle)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data",
			fmt.Sprintf("expected *client.Bundle, got %T", req.ProviderData),
		)
		return
	}
	r.clients = clients
}

func (r *zoneResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan zoneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	masters, diags := mastersFromSet(ctx, plan.Masters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneInfo := powerdns.ZoneInfo{
		Name:        plan.Name.ValueString(),
		Kind:        plan.Kind.ValueString(),
		Catalog:     plan.Catalog.ValueString(),
		Account:     plan.Account.ValueString(),
		Nameservers: []string{},
		SoaEditAPI:  plan.SoaEditAPI.ValueString(),
		Masters:     masters,
	}

	tflog.Debug(ctx, "creating PowerDNS zone", map[string]any{"name": zoneInfo.Name})

	created, err := r.clients.PDNS.CreateZone(ctx, zoneInfo)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create the PowerDNS zone", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *zoneResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state zoneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneInfo, err := r.clients.PDNS.GetZone(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the PowerDNS zone", err.Error())
		return
	}

	// PowerDNS answers a missing zone with an empty object rather than an
	// error, so an empty name is how deletion outside Terraform shows up.
	if zoneInfo.Name == "" {
		tflog.Warn(ctx, "zone no longer exists; removing from state",
			map[string]any{"id": state.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(zoneInfo.Name)
	state.Kind = types.StringValue(zoneInfo.Kind)
	state.Account = types.StringValue(zoneInfo.Account)
	state.SoaEditAPI = types.StringValue(zoneInfo.SoaEditAPI)
	state.Catalog = optionalString(zoneInfo.Catalog)

	// Masters are only meaningful for a Slave zone; for any other kind
	// PowerDNS returns an empty list, and writing that into state would show as
	// a permanent diff against a configuration that omits the attribute.
	if strings.EqualFold(zoneInfo.Kind, kindSlave) {
		masters, diags := types.SetValueFrom(ctx, types.StringType, zoneInfo.Masters)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Masters = masters
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *zoneResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan zoneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state zoneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	masters, diags := mastersFromSet(ctx, plan.Masters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	update := powerdns.ZoneInfoUpd{
		Name:       plan.Name.ValueString(),
		Kind:       plan.Kind.ValueString(),
		Catalog:    plan.Catalog.ValueString(),
		Account:    plan.Account.ValueString(),
		SoaEditAPI: plan.SoaEditAPI.ValueString(),
		Masters:    masters,
	}

	if err := r.clients.PDNS.UpdateZone(ctx, plan.ID.ValueString(), update); err != nil {
		resp.Diagnostics.AddError("Unable to update the PowerDNS zone", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *zoneResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state zoneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.clients.PDNS.DeleteZone(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete the PowerDNS zone", err.Error())
		return
	}
}

func (r *zoneResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mastersFromSet converts the schema set to the slice the client expects.
// A null or empty set yields nil rather than an empty slice, because the API
// payload omits masters entirely when there are none.
func mastersFromSet(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}

	var masters []string
	diags := set.ElementsAs(ctx, &masters, false)
	if len(masters) == 0 {
		return nil, diags
	}
	return masters, diags
}

// optionalString maps the API's empty string to a null attribute, so a zone
// with no catalog does not read back as "" against a configuration that omits
// the argument.
func optionalString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}
