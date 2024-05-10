resource "panop_zone" "zone1" {
  zone_name = "test1"
  tenant_id = 1
  zone_type = "dns"
}

resource "panop_zone" "zone2" {
  zone_name = "test2"
  tenant_id = 1
  zone_type = "dns"
}

resource "panop_zone" "zone3" {
  zone_name = "test3"
  tenant_id = 1
  zone_type = "dns"
}

resource "panop_zone" "zone4" {
}