# Team configurations using the team module

locals {
  # Try to load teams from teams.json, otherwise use default
  teams_json = try(jsondecode(file("${path.module}/teams.json")), null)
  
  # Only include teams that have assigned GitHub users (non-empty values)
  teams = local.teams_json != null ? [
    for team, user in local.teams_json.teams : team
    if user != ""
  ] : ["blue"]
}

module "team" {
  for_each = toset(local.teams)
  source   = "./modules/team"
  team_id  = each.value
}