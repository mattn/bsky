# Releasing

Deploying the app on Google Cloud Storage

```bash {"id":"01JBJA4MQ93SFDSW8C8EXNAHP1","interactive":"false"}
make static
gsutil cp ".build/static/*" gs://bsctl
gsutil cp ".build/static/web/*" gs://bsctl/web/
gsutil cp "web/*.svg" gs://bsctl/web/
```