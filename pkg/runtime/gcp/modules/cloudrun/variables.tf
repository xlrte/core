variable "service_name" {
  type    = string
}
variable "image_id" {
  type    = string
}
variable "region" {
  type    = string
}
variable "project" {
  type    = string
}
variable "environment"{
  type = string
}
variable "traffic" {
  type    = number
}
variable "memory" {
  type    = string
}
variable "cpu" {
  type    = number
}
variable "timeout" {
  type    = number
}
variable "max_requests" {
  type    = number
}
variable "min_instances" {
  type    = number
}
variable "max_instances" {
  type    = number
}
variable "is_public"{
  type = bool
}
variable "http2"{
  type = bool
}
variable "serverless_network"{
  type = string
  default = ""
}

variable "has_serverless_network"{
  type = bool
  default = false
}
variable "env"{
  type = map
}
variable "refs"{
  type = map
}

variable "secrets"{
  type = map
}

variable "subscription_topics"{
  type = list(object({
    topic_name=string,
    ack_deadline_seconds=number,
    message_retention_duration=string,
    retain_acked_messages=bool,
    enable_message_ordering=bool,
  }))
}
variable "publish_topics"{
  type = set(string)
}
variable "gcs_buckets"{
  type = list(object({
    bucket_name=string,
    role=string,
  }))
}
