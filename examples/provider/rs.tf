resource "panop_zone" "zone1" {
  zone_name = "fakeducksifiedshop.com"
  zone_type = "dns"
}

resource "panop_asset" "asset2" {
  for_each   = var.assets
  asset_name = each.key
  asset_type = "dns"
  zone_id    = panop_zone.zone1.id
}

variable "assets" {
  default = ["www"]
  type    = set(string)
}

