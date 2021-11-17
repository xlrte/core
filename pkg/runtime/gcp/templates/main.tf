
terraform {
  backend "gcs" {
    bucket  = "{{.StateStore}}"
    prefix  = "terraform/{{.Environment}}"
  }
}

provider "google" {
  project     = "{{.Project}}"
  region      = "{{.Region}}"
}

variable "project"{
  type = string
  default = "{{.Project}}"
}

variable "region"{
  type = string
  default = "{{.Region}}"
}

variable "environment"{
  type = string
  default = "{{.Environment}}"
}
