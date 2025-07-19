###########################################
# Local values
###########################################

locals {
  team_name_title = title(var.team_name)
  
  team_tags = merge(var.common_tags, {
    Team = var.team_name
  })
} 