// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccZoneResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: getProviderConfig(os.Getenv("PANOP_ACCESS_KEY")) + testAccExampleZoneResourceConfig("nonexist.panop.io"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("panop_zone.test", "zone_name", "nonexist.panop.io"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "panop_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccExampleZoneResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`
resource "panop_zone" "test" {
  zone_name = "%s"
}
`, configurableAttribute)
}
