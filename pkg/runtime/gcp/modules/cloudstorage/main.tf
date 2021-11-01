resource "google_storage_bucket" "bucket" {
  name          = "${var.bucket_name}-${var.environment}"
  location      = var.location
  storage_class = var.storage_class
  versioning {
      enabled = var.versioning_enabled
  }
}


resource "google_storage_bucket_access_control" "public_rule" {
  count    = var.public == true ? 1 : 0
  bucket = google_storage_bucket.bucket.name
  role   = "READER"
  entity = "allUsers"
}