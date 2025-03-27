// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAssetDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: getProviderConfig(os.Getenv("PANOP_ACCESS_KEY")) + testAccAssetDataAssetSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.panop_asset.test", "assets.0.asset_name", "api"),
				),
			},
		},
	})
}

const testAccAssetDataAssetSourceConfig = `
data "panop_asset" "test" {
}
`
