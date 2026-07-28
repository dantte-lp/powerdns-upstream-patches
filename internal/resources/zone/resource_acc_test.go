//go:build acceptance

package zone_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/dantte-lp/terraform-provider-powerdns/internal/provider"
	pdns "github.com/dantte-lp/terraform-provider-powerdns/powerdns"
)

// Acceptance tests run against the lab (task lab:up). They are not parallel:
// the lab instances are shared, and PowerDNS zone names are a global namespace
// within one server.
//
// The whole suite passes on both backends, so nothing here skips by backend.
// That is a property of powerdns_zone, not of the provider: views and
// networks will need the distinction when their turn comes (ADR 0005).

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"powerdns": func() (tfprotov6.ProviderServer, error) {
			ctx := context.Background()

			upgraded, err := tf5to6server.UpgradeServer(ctx, pdns.Provider().GRPCProvider)
			if err != nil {
				return nil, err
			}

			muxServer, err := tf6muxserver.NewMuxServer(ctx,
				providerserver.NewProtocol6(provider.New("acc")()),
				func() tfprotov6.ProviderServer { return upgraded },
			)
			if err != nil {
				return nil, err
			}
			return muxServer.ProviderServer(), nil
		},
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	for _, key := range []string{"PDNS_SERVER_URL", "PDNS_API_KEY"} {
		if os.Getenv(key) == "" {
			t.Fatalf("%s must be set for acceptance tests; run task lab:up first", key)
		}
	}
}

// runID is unique to one execution of the suite. Fixed names looked fine until
// a step failed at plan: CheckDestroy never ran, the zone survived, and every
// subsequent run failed at create with "Conflict" — a failure about the
// previous run, reported against the current one. AGENTS.md asks for
// tf-acc-<RUN_ID> namespacing precisely for this.
var runID = acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

// testAccZoneName keeps every object created by this suite identifiable and
// attributable to one run.
func testAccZoneName(suffix string) string {
	return fmt.Sprintf("tf-acc-%s-%s.test.", suffix, strings.ToLower(runID))
}

func testAccCheckZoneDestroy(s *terraform.State) error {
	client, err := accClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "powerdns_zone" {
			continue
		}

		exists, err := client.ZoneExists(context.Background(), rs.Primary.ID)
		if err != nil {
			// A destroyed zone may answer with an error rather than a false;
			// treat "not found" as success and anything else as a failure.
			if strings.Contains(strings.ToLower(err.Error()), "not found") {
				continue
			}
			return fmt.Errorf("checking whether zone %s survived destroy: %w", rs.Primary.ID, err)
		}
		if exists {
			return fmt.Errorf("zone %s still exists after destroy", rs.Primary.ID)
		}
	}
	return nil
}

func accClient() (*pdns.PowerDNSClient, error) {
	config := pdns.Config{
		ServerURL: os.Getenv("PDNS_SERVER_URL"),
		APIKey:    os.Getenv("PDNS_API_KEY"),
	}
	client, _, err := config.Clients(context.Background())
	return client, err
}

// TestAccZone_native is the baseline: create, read back, re-plan empty, import
// and verify, destroy clean. It runs on every backend.
func TestAccZone_native(t *testing.T) {
	name := testAccZoneName("native")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "powerdns_zone" "test" {
  name = %q
  kind = "Native"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("powerdns_zone.test", "name", name),
					resource.TestCheckResourceAttr("powerdns_zone.test", "kind", "Native"),
					resource.TestCheckResourceAttr("powerdns_zone.test", "account", "admin"),
					resource.TestCheckResourceAttrSet("powerdns_zone.test", "id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ResourceName:      "powerdns_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccZone_kindCaseInsensitive covers the plan modifier that replaced the
// SDKv2 DiffSuppressFunc. PowerDNS returns "Native" for a configured "native";
// without the modifier this is a permanent diff.
func TestAccZone_kindCaseInsensitive(t *testing.T) {
	name := testAccZoneName("kindcase")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "powerdns_zone" "test" {
  name = %q
  kind = "native"
}
`, name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccZone_slaveWithIPv6Masters is the regression test for upstream #73.
// The configuration is the one from the issue: an IPv6 master alongside an
// IPv4 one. Before the fix this failed at plan with
// "more than one colon in <ip>:<port> string".
func TestAccZone_slaveWithIPv6Masters(t *testing.T) {
	name := testAccZoneName("ipv6masters")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "powerdns_zone" "test" {
  name         = %q
  kind         = "Slave"
  soa_edit_api = "DEFAULT"
  masters = [
    "fd92:81e1:e314:ea7b:0000:1234:5678:60ab",
    "192.168.123.45",
  ]
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("powerdns_zone.test", "masters.#", "2"),
					resource.TestCheckTypeSetElemAttr("powerdns_zone.test", "masters.*",
						"fd92:81e1:e314:ea7b:0000:1234:5678:60ab"),
					resource.TestCheckTypeSetElemAttr("powerdns_zone.test", "masters.*",
						"192.168.123.45"),
				),
			},
		},
	})
}

// TestAccZone_mastersValidatedOnUpdate is the regression test for D-02. The
// inherited resource validated masters on create only, so the documented
// workaround for #73 — create with IPv4, then edit to add IPv6 — let an
// unvalidated value into state. Here the second step supplies a genuinely
// invalid master and must fail at plan.
func TestAccZone_mastersValidatedOnUpdate(t *testing.T) {
	name := testAccZoneName("mastersupd")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "powerdns_zone" "test" {
  name    = %q
  kind    = "Slave"
  masters = ["192.0.2.1"]
}
`, name),
			},
			{
				// PlanOnly: the point is that this never reaches apply. The
				// inherited resource would have accepted it, because its update
				// path validated nothing.
				Config: fmt.Sprintf(`
resource "powerdns_zone" "test" {
  name    = %q
  kind    = "Slave"
  masters = ["192.0.2.1", "ns1.example.com"]
}
`, name),
				ExpectError: regexpInvalidMaster(),
				PlanOnly:    true,
			},
			{
				// The suite's destroy runs against the last step's configuration.
				// Leaving the invalid one last makes the validator block the
				// cleanup, which reads as a test failure about the wrong thing —
				// so the run ends on a valid configuration.
				Config: fmt.Sprintf(`
resource "powerdns_zone" "test" {
  name    = %q
  kind    = "Slave"
  masters = ["192.0.2.1"]
}
`, name),
			},
		},
	})
}

// TestAccZone_mastersRejectedOnNonSlave covers the ValidateConfig rule: the
// failure must arrive at plan, not apply.
func TestAccZone_mastersRejectedOnNonSlave(t *testing.T) {
	name := testAccZoneName("mastersnative")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "powerdns_zone" "test" {
  name    = %q
  kind    = "Native"
  masters = ["192.0.2.1"]
}
`, name),
				ExpectError: regexpMastersOnlySlave(),
				PlanOnly:    true,
			},
		},
	})
}

// TestAccZone_catalog exercises the catalog attribute, which requires
// PowerDNS 4.7 or later. Catalog zones are stored as ordinary zones, so this
// runs on both backends.
func TestAccZone_catalog(t *testing.T) {
	catalog := testAccZoneName("cat")
	member := testAccZoneName("catmember")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "powerdns_zone" "catalog" {
  name = %q
  kind = "Producer"
}

resource "powerdns_zone" "member" {
  name    = %q
  kind    = "Native"
  catalog = powerdns_zone.catalog.name
}
`, catalog, member),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("powerdns_zone.member", "catalog", catalog),
				),
			},
		},
	})
}
