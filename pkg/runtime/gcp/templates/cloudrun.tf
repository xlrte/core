module "cloudrun-{{.ServiceName}}" {
  source = "../modules/cloudrun"
  service_name = "{{.ServiceName}}"
  project = var.project
  region = var.region
  environment = var.environment
  {{ if .HasServerlessNetwork}}
  serverless_network = {{.ServerlessNetworkLink}}
  has_serverless_network = {{.HasServerlessNetwork}}
  {{ end }}
  image_id = "{{.ImageID}}"
  traffic = {{.Traffic}}
  memory = "{{.RuntimeConfig.Memory}}"
  cpu = {{.RuntimeConfig.CPU}}
  timeout = {{.RuntimeConfig.Timeout}}
  max_requests = {{.RuntimeConfig.MaxRequests}}
  min_instances = {{.RuntimeConfig.Scaling.MinInstances}}
  max_instances =  {{.RuntimeConfig.Scaling.MaxInstances}}
  is_public = {{.IsPublic}}
  http2 = {{.Http2}}
  env = { {{ range $key, $value := .Env.Vars }}
    {{ $key }} = "{{ $value }}"
  {{ end }}}
  refs = { {{ range $key, $value := .Env.Refs }}
    {{ $key }} = {{ $value }}
  {{ end }}}
  secrets = { {{ range $key, $value := .Env.Secrets }}
    {{ $key }} = {{ $value }}
  {{ end }}}

  publish_topics = [{{ range $key, $value := .PublishTopics }}"{{ $value }}",{{ end }}]

  subscription_topics = [{{ range $key, $value := .SubscribeTopics }}
    {
      topic_name = "{{$value.TopicName}}"
      ack_deadline_seconds = {{$value.AckDeadline}}
      message_retention_duration = "{{$value.Retention}}"
      retain_acked_messages = {{$value.RetainAckedMessages}}
      enable_message_ordering = {{$value.EnableMessageOrdering}}
    },{{ end }}]

  gcs_buckets = [{{ range $key, $value := .CloudStorage }}
    {
      bucket_name = "{{$value.Bucket}}"
      role = "{{$value.Role}}"
    },{{ end }}]

  depends_on = [{{ range $key, $value := .DependsOn }}{{ $value }},{{ end }}]

}

output "cloud_run_endpoint-{{.ServiceName}}" {
  value = module.cloudrun-{{.ServiceName}}.cloud_run_endpoint
}
