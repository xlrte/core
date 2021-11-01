provider "google-beta" {
  project = var.project
  region  = var.region
}

terraform {
  # This module is now only being tested with Terraform 1.0.x. However, to make upgrading easier, we are setting
  # 0.12.26 as the minimum version, as that version added support for required_providers with source URLs, making it
  # forwards compatible with 1.0.x code.
  required_version = ">= 0.12.26"

  required_providers {
    google-beta = {
      source  = "hashicorp/google-beta"
      version = ">= 3.57.0"
    }
  }
}

# ------------------------------------------------------------------------------
# CREATE A RANDOM SUFFIX AND PREPARE RESOURCE NAMES
# ------------------------------------------------------------------------------

resource "random_id" "name" {
  byte_length = 2
}

locals {
  # If name_override is specified, use that - otherwise use the name_prefix with a random string
  instance_name        = "${var.instance_name}-${var.environment}-${random_id.name.hex}"
  private_ip_name      = "private-ip-${var.environment}" 
}

resource "google_compute_global_address" "private_ip_address" {
  provider      = google-beta
  name          = local.private_ip_name
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = var.network_self_link
}

# Establish VPC network peering connection using the reserved address range
resource "google_service_networking_connection" "private_vpc_connection" {
  provider                = google-beta
  network                 = var.network_self_link
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}

# ------------------------------------------------------------------------------
# CREATE DATABASE INSTANCE WITH PRIVATE IP
# ------------------------------------------------------------------------------

module "postgres" {
  source = "git@github.com:xlrte/terraform-google-sql.git//modules/cloud-sql?ref=v0.6.0"

  project = var.project
  region  = var.region
  name    = local.instance_name
  db_name = "${var.db_name}-${var.environment}"

  engine       = var.postgres_version
  machine_type = var.machine_type
  disk_size = var.disk_size

  # To make it easier to test this example, we are disabling deletion protection so we can destroy the databases
  # during the tests. By default, we recommend setting deletion_protection to true, to ensure database instances are
  # not inadvertently destroyed.
  deletion_protection = var.deletion_protection

  backup_enabled = var.backup_enabled
  backup_start_time = var.backup_start_time
  postgres_point_in_time_recovery_enabled = var.postgres_point_in_time_recovery_enabled
  maintenance_window_day = var.maintenance_window_day
  maintenance_window_hour = var.maintenance_window_hour

  # These together will construct the master_user privileges, i.e.
  # 'master_user_name'@'master_user_host' IDENTIFIED BY 'master_user_password'.
  # These should typically be set as the environment variable TF_VAR_master_user_password, etc.
  # so you don't check these into source control."
  master_user_password = var.master_user_password

  master_user_name = var.master_user_name
  master_user_host = "%"

  # Pass the private network link to the module
  private_network = var.network_self_link

  # Wait for the vpc connection to complete
  dependencies = [google_service_networking_connection.private_vpc_connection.network]

}