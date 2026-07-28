package zone

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/dantte-lp/terraform-provider-powerdns/powerdns"
)

// zoneNameValidator wraps the inherited ValidateZoneName so the framework
// resource enforces exactly the rule the SDKv2 one did — including PowerDNS
// zone variants such as "example.com..internal". Reimplementing it would risk
// the two diverging while both are served through the mux.
type zoneNameValidator struct{}

// ZoneNameValidator returns the validator applied to zone-name attributes.
func ZoneNameValidator() validator.String {
	return zoneNameValidator{}
}

func (v zoneNameValidator) Description(_ context.Context) string {
	return "must be a fully qualified domain name with a trailing dot, " +
		`or a PowerDNS zone variant such as "example.com..internal"`
}

func (v zoneNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v zoneNameValidator) ValidateString(
	_ context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// An empty catalog is how "no catalog" is expressed; it is not a name.
	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	_, errs := powerdns.ValidateZoneName(value, req.Path.String())
	for _, err := range errs {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid zone name", err.Error())
	}
}
