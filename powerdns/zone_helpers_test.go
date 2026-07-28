package powerdns

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Zone existence helpers shared by the SDKv2 acceptance tests that remain here.
//
// They outlived resource_powerdns_zone_test.go, which was removed when
// powerdns_zone moved to the framework half of the mux: the data source and
// powerdns_reverse_zone still use them, and both are still SDKv2.

func testAccCheckPDNSZoneDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "powerdns_zone" {
			continue
		}

		client := testAccProvider.Meta().(*ProviderClients).PDNS
		exists, err := client.ZoneExists(context.Background(), rs.Primary.Attributes["name"])
		if err != nil {
			return fmt.Errorf("Error checking if zone still exists: %#v", rs.Primary.ID)
		}
		if exists {
			return fmt.Errorf("Zone still exists: %#v", rs.Primary.ID)
		}

	}
	return nil
}

func testAccCheckPDNSZoneExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		client := testAccProvider.Meta().(*ProviderClients).PDNS
		exists, err := client.ZoneExists(context.Background(), rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Zone does not exist: %#v", rs.Primary.ID)
		}
		return nil
	}
}
