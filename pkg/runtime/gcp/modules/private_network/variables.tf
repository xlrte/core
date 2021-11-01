
variable "project" {
  description = "The project ID to host the database in."
  type        = string
}

variable "region" {
  description = "The region to host the database in."
  type        = string
}

variable "environment"{
  type = string
}

variable "min_instances"{
  type = number
}

variable "max_instances"{
  type = number
}

variable "instance_type"{
  type = string
}