resource "google_pubsub_topic" "topic" {
  name = "${var.name}-${var.environment}"
}