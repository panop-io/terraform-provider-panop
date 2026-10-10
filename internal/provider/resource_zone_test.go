// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccZoneResource(t *testing.T) {
	zoneName := "nonexist.panop.io"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: getProviderConfig(os.Getenv("PANOP_ACCESS_KEY")) + testAccZoneResourceConfig(zoneName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panop_zone.test",
						tfjsonpath.New("zone_name"),
						knownvalue.StringExact(zoneName),
					),
					statecheck.ExpectKnownValue(
						"panop_zone.test",
						tfjsonpath.New("zone_type"),
						knownvalue.StringExact("dns"),
					),
					statecheck.ExpectKnownValue(
						"panop_zone.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"panop_zone.test",
						tfjsonpath.New("token"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "panop_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
				// token comes from the create response. Import runs Read against
				// the zone list, which is not guaranteed to return the same value.
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccZoneResourceConfig(zoneName string) string {
	return fmt.Sprintf(`
resource "panop_zone" "test" {
  zone_name = %[1]q
}
`, zoneName)
}
