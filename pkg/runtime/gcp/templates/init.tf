variable "project"{
  type = string
  default = "{{.Project}}"
}

variable "region"{
  type = string
  default = "{{.Region}}"
}

variable "billing_account"{
  type = string
  default = "{{.BillingAccount}}"
}

variable "services"{
  type = set(string)
  default = [
    "servicenetworking.googleapis.com",
    "vpcaccess.googleapis.com",
    "containerregistry.googleapis.com",
    "dns.googleapis.com",
    "pubsub.googleapis.com",
    "run.googleapis.com",
    "secretmanager.googleapis.com",
    "sql-component.googleapis.com",
    "storage.googleapis.com",
  ]
}

provider "google" {
  project     = var.project
  region      = var.region
}

resource "google_project" "init_project" {
  name       = var.project
  project_id = var.project
  billing_account = var.billing_account
}

resource "google_project_service" "project" {
  for_each = var.services
  project = google_project.init_project.project_id
  service = each.value

  timeouts {
    create = "30m"
    update = "40m"
  }

  disable_dependent_services = true
}

resource "google_storage_bucket" "state_bucket" {
  name          = "xlrte-state-{{.Project}}"
  storage_class = "STANDARD"
  location = "US"
  versioning {
      enabled = true
  }
  depends_on = [
    google_project_service.project
  ]
}

