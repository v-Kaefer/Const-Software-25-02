variable "public_key_path" {
  description = "Path to SSH public key"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

variable "admin_users" {
  description = "List of admin users to create"
  type = list(object({
    email = string
    name  = string
  }))
  default = [
    {
      email = "admin@example.com"
      name  = "Admin User"
    }
  ]
}

variable "reviewer_users" {
  description = "List of reviewer users to create"
  type = list(object({
    email = string
    name  = string
  }))
  default = [
    {
      email = "reviewer@example.com"
      name  = "Reviewer User"
    }
  ]
}

variable "user_cognito" {
  description = "Main user attributes for Cognito"
  type        = map(string)
  default = {
    email         = "user@example.com"
    name          = "Regular User"
    "custom:role" = "user"
  }
}
