### To login:

* gcloud auth application-default login

### GCP Services

* servicenetworking.googleapis.com
* vpcaccess.googleapis.com
* containerregistry.googleapis.com
* dns.googleapis.com
* pubsub.googleapis.com
* run.googleapis.com
* secretmanager.googleapis.com
* sql-component.googleapis.com
* storage.googleapis.com

### Push a docker image
  gcloud auth configure-docker
  docker build . -f hello-example.Dockerfile -t gcr.io/[project]/hello-app:v1
  docker push gcr.io/[project]/hello-app:v1