## What `xlrte` does, in user story form
As a developer,

I want to be able to say _"I want to deploy a service that uses a database, a block storage bucket & publishes messages to a topic while listening to another"_,

without having to figure out IAM permissions, network setup, initialization order of resources and the dozens of other infrastructure configuration issues that arise.

#### Focus on architecture, not infrastructure
`xlrte` enables just this. The configuration below is all that is needed.

```yaml
name: my-shiny-app
runtime: cloudrun
spec:
  base_name: my-shiny-app
  http:
    public: true
    http2: false
depends_on:
  cloudsql: 
  - name: my-pg-db
    type: postgres
  pubsub:
    consume:
    - name: upload_events
    produce:
    - name: resize_events
  cloudstorage:
  - name: media-uploads
    public: true
    access: readwrite
```

To find out more:
* [Read the documentation](https://xlrte.dev/docs/getting-started/setup-gcp)
* checkout our [example project](https://github.com/xlrte/example-app-gcp)