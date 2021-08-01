variable "service_name" {
  type = string
  default = "cloudrun-srv"
}

variable "image_id" {
  type = string
  default = "gcr.io/chaordic/hello-app:v2"
}

variable "region" {
  type    = string
  default = "europe-west6"
}

variable "project" {
  type    = string
  default = "chaordic"
}

variable "traffic" {
  type    = number
  default = 100
}

variable "domain" {
  type = string
  default = "cde.app"
}

variable "dns_name" {
  type = string
  default = "cde.app."
}

variable "managed_zone"{
  type = string
  default = "cdeapp"
}

variable network_name{
  type = string
  default = "cde-network"
}