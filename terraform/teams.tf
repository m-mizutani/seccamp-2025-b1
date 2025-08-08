# Team configurations using the team module

locals {
  teams = ["blue"]
}

module "team" {
  for_each = toset(local.teams)
  source   = "./modules/team"
  team_id  = each.value
}