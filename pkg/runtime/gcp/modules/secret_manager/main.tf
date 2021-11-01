resource "google_secret_manager_secret" "secret" {
  provider = google-beta
  project = var.project
  secret_id = "${var.secret_id}-${var.environment}"

  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "secret-version" {
  provider = google-beta
  secret = google_secret_manager_secret.secret.id

  secret_data = var.secret_data
}
