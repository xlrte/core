# ---------------------------------------------------------------------------------------------------------------------
# REQUIRED PARAMETERS
# These variables are expected to be passed in by the operator
# ---------------------------------------------------------------------------------------------------------------------

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

variable "network_self_link"{
  type = string
}

variable "master_user_name" {
  description = "The username part for the default user credentials, i.e. 'master_user_name'@'master_user_host' IDENTIFIED BY 'master_user_password'. This should typically be set as the environment variable TF_VAR_master_user_name so you don't check it into source control."
  type        = string
  sensitive   = true
}

variable "master_user_password" {
  description = "The password part for the default user credentials, i.e. 'master_user_name'@'master_user_host' IDENTIFIED BY 'master_user_password'. This should typically be set as the environment variable TF_VAR_master_user_password so you don't check it into source control."
  type        = string
  sensitive   = true
}

variable "deletion_protection"{
  type = bool
}

variable "machine_type" {
  description = "The machine type to use, see https://cloud.google.com/sql/pricing for more details"
  type        = string
}

variable "db_name" {
  description = "Name for the db"
  type        = string
}

variable "instance_name" {
  description = "Name for the db"
  type        = string
}

variable "disk_size"{
  type = number
}

# ---------------------------------------------------------------------------------------------------------------------
# OPTIONAL PARAMETERS
# Generally, these values won't need to be changed.
# ---------------------------------------------------------------------------------------------------------------------

variable "postgres_version" {
  description = "The engine version of the database, e.g. `POSTGRES_13`. See https://cloud.google.com/sql/docs/db-versions for supported versions."
  type        = string
  default     = "POSTGRES_13"
}

variable "backup_enabled" {
  description = "Set to false if you want to disable backup."
  type        = bool
}

variable "backup_start_time" {
  description = "HH:MM format (e.g. 04:00) time indicating when backup configuration starts. NOTE: Start time is randomly assigned if backup is enabled and 'backup_start_time' is not set"
  type        = string
}

variable "postgres_point_in_time_recovery_enabled" {
  description = "Will restart database if enabled after instance creation - only applicable to PostgreSQL"
  type        = bool
}

variable "maintenance_window_day" {
  description = "Day of week (1-7), starting on Monday, on which system maintenance can occur. Performance may be degraded or there may even be a downtime during maintenance windows."
  type        = number
}

variable "maintenance_window_hour" {
  description = "Hour of day (0-23) on which system maintenance can occur. Ignored if 'maintenance_window_day' not set. Performance may be degraded or there may even be a downtime during maintenance windows."
  type        = number
}