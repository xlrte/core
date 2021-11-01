module "cloudstorage-{{.BucketName}}" {
  source = "../modules/cloudstorage"
  bucket_name = "{{.BucketName}}"
  storage_class = "{{.StorageClass}}"
  location = "{{.Location}}"
  versioning_enabled = {{.VersioningEnabled}}
  environment = var.environment
  public = {{.IsPublic}}
}
