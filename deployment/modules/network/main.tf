
resource "google_compute_region_network_endpoint_group" "cloudrun_neg" {
  name                  = "${var.network_name}-cloudrun-neg"
  network_endpoint_type = "SERVERLESS"
  project               = var.project
  region                = var.region
  cloud_run {
    service = var.cloud_run_service_name
  }
}


module "lb_http" {
  source            = "GoogleCloudPlatform/lb-http/google//modules/serverless_negs"
  version           = "~> 4.4"
  project           = var.project
  name              = "${var.network_name}-lb"

  ssl                             = true
  managed_ssl_certificate_domains = [var.domain]
  https_redirect                  = true
  backends = {
    default = {
      description                     = null
      enable_cdn                      = false
      custom_request_headers          = null
      custom_response_headers         = null
      security_policy                 = null

      log_config = {
        enable = true
        sample_rate = 1.0
      }

      groups = [
        {
          # Your serverless service should have a NEG created that's referenced here.
          group = google_compute_region_network_endpoint_group.cloudrun_neg.id
        }
      ]

      iap_config = {
        enable               = false
        oauth2_client_id     = null
        oauth2_client_secret = null
      }
    }
  }
}


resource "google_dns_record_set" "resource-recordset" {
  provider = google-beta
  managed_zone = var.managed_zone
  name         = var.dns_name
  project      = var.project
  type         = "A"
  rrdatas      = [module.lb_http.external_ip]
  ttl          = 86400
}
