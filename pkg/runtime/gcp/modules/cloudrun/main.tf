resource "google_service_account" "service_account" {
  account_id   = "${var.service_name}-${var.environment}"
  display_name = "${var.service_name}-${var.environment}-account"
}

resource "google_cloud_run_service" "default" {
  provider = google-beta
  name     = "${var.service_name}-${var.environment}"
  location = var.region
  project  = var.project

  depends_on = [
    google_service_account.service_account, google_secret_manager_secret_iam_binding.binding
  ]

  template {
    spec {

      service_account_name = google_service_account.service_account.email
      containers {
        image = var.image_id
        resources {
          limits = {
            cpu    = var.cpu
            memory = var.memory
          }
        }
        ports {
          name           = var.http2 ? "h2c": "http1" 
          container_port = 8080
        }
        env {
          name = "XLRTE_ENV"
          value = var.environment
        }
        env {
          name = "GCP_PROJECT_ID"
          value = var.project
        }
        dynamic "env" {
          for_each = var.env
          content {
            name  = env.key
            value = env.value
          }
        }

        dynamic "env" {
          for_each = var.refs
          content {
            name  = env.key
            value = env.value
          }
        }

        dynamic "env" {
          for_each = var.secrets
          content{
            name = env.key
            value_from {
              secret_key_ref {
                name = env.value
                key = "latest"
              }
            }
          }
        }
        
      }
      container_concurrency = var.max_requests
      timeout_seconds       = var.timeout
    }
    metadata {
      annotations = {
        "autoscaling.knative.dev/minScale" = var.min_instances
        "autoscaling.knative.dev/maxScale" = var.max_instances
        "run.googleapis.com/vpc-access-egress" = var.has_serverless_network ? "private-ranges-only": null
        "run.googleapis.com/vpc-access-connector" = var.has_serverless_network ? "${var.serverless_network}": null
      }
    }
  }


  metadata {
    annotations = {
      "run.googleapis.com/launch-stage" = "BETA"
    }
  }

  traffic {
    percent         = var.traffic
    latest_revision = true
  }
}

data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "noauth" {
  count    = var.is_public == true ? 1 : 0
  location = google_cloud_run_service.default.location
  project  = google_cloud_run_service.default.project
  service  = google_cloud_run_service.default.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

resource "google_secret_manager_secret_iam_binding" "binding" {
  for_each = var.secrets
  project = var.project
  secret_id = each.value
  role = "roles/secretmanager.secretAccessor"
  members = [
    "serviceAccount:${google_service_account.service_account.email}",
  ]
}


resource "google_pubsub_topic_iam_binding" "pubsub_binding" {
  for_each = var.publish_topics
  depends_on = [
    google_service_account.service_account,
  ]
  project = var.project
  topic = "${each.value}-${var.environment}"
  role = "roles/pubsub.publisher"
  members = [
    "serviceAccount:${google_service_account.service_account.email}",
  ]
}

resource "google_storage_bucket_iam_binding" "binding" {
  for_each = {
    for index, bucket in var.gcs_buckets:
    index => bucket
  }
  depends_on = [
    google_service_account.service_account,
  ]
  bucket = "${each.value.bucket_name}-${var.environment}"
  role = each.value.role #"roles/storage.admin" # roles/storage.objectAdmin, roles/storage.objectViewer
  members = [
    "serviceAccount:${google_service_account.service_account.email}",
  ]
}
resource "google_pubsub_subscription" "push_subscription" {
  for_each = {
    for index, sub in var.subscription_topics:
    index => sub
  }
  name  = "${each.value.topic_name}_${var.service_name}-${var.environment}"
  topic = "${each.value.topic_name}-${var.environment}"

  ack_deadline_seconds = each.value.ack_deadline_seconds
  message_retention_duration = each.value.message_retention_duration
  retain_acked_messages = each.value.retain_acked_messages
  enable_message_ordering = each.value.enable_message_ordering

  push_config {
    push_endpoint = google_cloud_run_service.default.status[0].url
  }
}