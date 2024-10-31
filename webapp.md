# Webapp

You can run this as a webapp.

To spin up a local server

```bash
python3 -m http.server 8000
```

Copy fil to GCS

```bash {"id":"01JBGFVC69TDA4VAHVP30Z5HNN","interactive":"true"}
# Copy a file to GCS
gsutil cp -r .build/web/* gs://bsctl/
```

* Define a terraform policy to make the bucket bsctl public

# TODO(jeremy): I need to configure IAC to use GCS for state

```bash {"id":"01JBGFY5RE72NA9NEY3HYSQBGH","interactive":"false"}
# 1. Append the Terraform policy to the iac/main.tf file to make the GCS bucket public
echo 'resource "google_storage_bucket_iam_member" "public_access" {
  bucket = "bsctl"
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}' >> iac/main.tf
```

```bash {"id":"01JBGFYHDGZ6ZH6MNJH03JVGQ1","interactive":"true"}
# 1. Apply the Terraform configuration to make the GCS bucket public
cd iac
terraform init
terraform apply -auto-approve
```

## Deploying on GCS

* I can load index.html by going to https://storage.googleapis.com/bsctl/index.html
* But this gives me a 404 trying to fetch https://storage.googleapis.com/wasm_exec.js
* The URL should be https://storage.googleapis.com/bsctl/wasm_exec.js
* use `handler.Resources=app.CustomProvider` to add a prefix to it

* Reupload the paths to gcs

```bash {"id":"01JBHEM8Z1M345TYDXTPFBN7BY","interactive":"false"}
gsutil cp -r .build/web/* gs://bsctl/
```

Now I'm getting a 404 hitting 
https://storage.googleapis.com/bsctl/web/app.wasm

```bash {"id":"01JBHEPC46XQWS3E4SHFM0FC47","interactive":"false"}
gsutil ls gs://bsctl/
```

* Ya so its missing
* So to build it I need to build the PWA and run it static

```bash {"id":"01JBHEQFQQRX1SFVSG2D5RRF2K","interactive":"false"}
make pwa
BUILD_STATIC=true .build/pwa-server
```

```bash {"id":"01JBHERA9JCXYGNPKDKTH4D6GS","interactive":"false"}
gsutil cp -r .build/web/* gs://bsctl/
```

```bash
gsutil ls gs://bsctl
```

* So its still not there did it get deleted when I ran .build/pwa-server

To investigate whether the `app.wasm` file was deleted or if it never existed, you can recheck the contents of the `web` directory where the build output is stored. Here’s how to do that:
1. **Check the local .build/web directory**

```bash {"id":"01JBHESPRYNENBTFAW6AXHJ6V8","interactive":"false"}
ls -la .build/
```

I see there our build script is putting app.wasm into the web directory

```bash
rm -rf .build/*
```

```bash {"id":"01JBHF1BJ2EQHKBWV781T0HG42","interactive":"false"}
# Rebuild the PWA to regenerate app.wasm
make pwa
BUILD_STATIC=true .build/pwa-server

# After rebuilding, check the contents of the web directory again to confirm that app.wasm is present
ls -la .build/web
```

```bash
make static
```

* Recopy to GCS

```bash {"id":"01JBHF9SK693368DB9PMXH3Y5H","interactive":"false"}
# Copy the newly built app.wasm file to GCS
gsutil cp -r .build/web gs://bsctl/
```

* The page shows me a 404 error in the GoApp but I don't see any errors in the chrome console
* I wonder if the problem is when I setup my route handler

  `app.Route("/", &CommandApp{})` I'm not accounting for the path prefix

```bash {"id":"01JBHFWHE8ZK2XNWJMMA5Q1TJ8","interactive":"true"}
make static
gsutil cp -r .build/web gs://bsctl/
```

* My logging statements are getting hit so it found my entry point