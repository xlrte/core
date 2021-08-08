variable "service_name" {
  type = string
  default = "{{.ServiceName}}"
}

variable "image_id" {
  type = string
  default = "{{.ImageID}}"
}

variable "region" {
  type    = string
  default = "{{.Region}}"
}

variable "project" {
  type    = string
  default = "{{.Project}}"
}

variable "traffic" {
  type    = number
  default = {{.Traffic}}
}