variable "project" {
  default = "dev-5-373713"
}

variable "credentials_file" {
  default = "dev-5-373713-9504544ba89c.json"
}

variable "region" {
  default = "us-central1"
}

variable "zone" {
  default = "us-central1-c"
}

variable "load_balancing_scheme" {
  description = "Load balancing scheme type (EXTERNAL for classic external load balancer, EXTERNAL_MANAGED for Envoy-based load balancer, INTERNAL for classic internal load balancer, and INTERNAL_SELF_MANAGED for internal load balancer)"
  type        = string
  default     = "INTERNAL"
}
