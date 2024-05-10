// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccZoneDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: getProviderConfig(os.Getenv("PANOP_ACCESS_KEY")) + testAccZoneDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.panop_zone.test", "zones.0.zone_name", "fakeducksifiedshop.com"),
				),
			},
		},
	})
}

const testAccZoneDataSourceConfig = `
data "panop_zone" "test" {
}
`
