variable "admin_users" {
  description = "List of admin users to create in Cognito"
  type = list(object({
    email = string
    name  = string
  }))
  default = []
}

variable "reviewer_users" {
  description = "List of reviewer users to create in Cognito"
  type = list(object({
    email = string
    name  = string
  }))
  default = []
}

variable "user_cognito" {
  description = "Main user configuration for Cognito"
  type        = map(string)
  default = {
    email = "user@example.com"
    name  = "Default User"
  }
}
