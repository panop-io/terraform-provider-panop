resource "panop_zone" "zone1" {
  zone_name = "fakeducksifiedshop2.com"
}

resource "panop_zone" "zone2" {
  zone_name = "fakeducksifiedshop.com"
}


resource "panop_asset" "asset1" {
  asset_name = "mail"
  zone_id    = panop_zone.zone1.id
}

