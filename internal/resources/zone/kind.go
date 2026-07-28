package zone

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// caseInsensitiveModifier keeps state when the only difference from the
// configuration is letter case.
//
// PowerDNS accepts "slave" and returns "Slave", so a configuration written in
// lower case would otherwise show a permanent diff. The SDKv2 resource handled
// this with DiffSuppressFunc; the framework equivalent is a plan modifier.
type caseInsensitiveModifier struct{}

func caseInsensitive() planmodifier.String {
	return caseInsensitiveModifier{}
}

func (m caseInsensitiveModifier) Description(_ context.Context) string {
	return "suppresses a diff when the planned and stored values differ only in case"
}

func (m caseInsensitiveModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m caseInsensitiveModifier) PlanModifyString(
	_ context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	if strings.EqualFold(req.StateValue.ValueString(), req.PlanValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}
