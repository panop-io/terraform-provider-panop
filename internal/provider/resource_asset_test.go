// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAssetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: getProviderConfig(os.Getenv("PANOP_ACCESS_KEY")) + testAccAssetResourceConfig("www", "dns", 337),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("panop_asset.test", "asset_name", "www"),
					resource.TestCheckResourceAttr("panop_asset.test", "asset_type", "dns"),
					resource.TestCheckResourceAttr("panop_asset.test", "zone_id", "337"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "panop_asset.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAssetResourceConfig(assetName, assetType string, zoneId int64) string {
	return fmt.Sprintf(`
resource "panop_asset" "test" {
  asset_name = "%s"
  asset_type = "%s"
  zone_id = %d
}
`, assetName, assetType, zoneId)
}
