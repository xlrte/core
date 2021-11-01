variable "secret_id" {
  type    = string
}
variable "secret_data"{
  type = string
  sensitive   = true
}
variable "environment"{
  type = string
}

variable "project"{
  type = string
}