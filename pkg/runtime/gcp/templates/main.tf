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
