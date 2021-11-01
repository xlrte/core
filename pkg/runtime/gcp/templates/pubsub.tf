module "pubsub-{{.TopicName}}" {
  source = "../modules/pubsub"
  name = "{{.TopicName}}"
  environment = var.environment
}
