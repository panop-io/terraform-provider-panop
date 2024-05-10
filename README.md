# Terraform Provider Panop

_This provider is used to create/delete Panop ressource like:_
- zones
- assets

## Using the provider
### Resource
Zone
```
resource "panop_zone" "zone1" {
  zone_name = "ducksifiedshop.com"
}
```
Asset
```
resource "panop_asset" "asset1" {
  asset_name = "mail"
  zone_id = panop_zone.zone1.id
}
```
### data source
zone
```
data "panop_zone" "allzones" {
}
```
asset
```
data "panop_asset" "allassets" {
}
```
asset filter
```
data "panop_asset" "allassets" {
  zone_id = 416
}
```

```bash
terraform apply
```

```bash
terraform state list
```
