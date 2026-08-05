terraform {
  required_providers {
    deeplink = {
      source = "yinebebt/deeplink"
    }
  }
}

variable "endpoint" {
  type        = string
  description = "Base URL of your deeplink server"
  default     = "https://link.example.com"
}

variable "api_key" {
  type        = string
  sensitive   = true
  description = "API key for mutating deeplink endpoints"
}

provider "deeplink" {
  endpoint = var.endpoint
  api_key  = var.api_key
}

resource "deeplink_link" "docs" {
  type        = "redirect"
  url         = "https://example.com/docs"
  title       = "Product docs"
  description = "Canonical docs short link managed by Terraform"
}

output "short_link" {
  value = deeplink_link.docs.short_link
}

output "short_id" {
  value = deeplink_link.docs.short_id
}
