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
* When I run it locally it works
* So I know from the logging statements it is hitting "main.go"

* If I look at app.js that gets loaded I see

"const goappEnv = {"GOAPP_INTERNAL_URLS":"null","GOAPP_ROOT_PREFIX":"/bsctl","GOAPP_STATIC_RESOURCES_URL":"/bsctl","GOAPP_VERSION":"9b4d7c591a4770be3729e56a8001c1b49f3612a2"};

* Maybe my layout is incorrect
 https://go-app.dev/github-deploy

* I don't have app.wasm inside the web subdirectory

```bash {"id":"01JBHHHYTEHTM4CB6HK9X0VWM2","interactive":"false"}
gsutil ls -r gs://bsctl/
```

```bash
gsutil cp .build/web/app.wasm gs://bsctl/web/app.wasm
```

The command to remove the old `app.wasm` file from the GCS bucket was executed successfully. Now the next step is to check again if the newly copied `app.wasm` file is present in the GCS bucket.
Here's the command to verify that:

```bash {"id":"01JBHHN07M14Y1HS552CPNXYEK","interactive":"false"}
gsutil rm -rf gs://bsctl/
```

```markdown
gsutil ls -r gs://bsctl/
```

So bucket is empty now

```bash {"id":"01JBHHPKRX9ZAPC2CYRJZF91T7","interactive":"true"}
gsutil cp -r .build/web gs://bsctl/
```

* Looks like that deleted the bucket

It seems that the error indicates that the destination bucket `gs://bsctl` does not exist. To resolve this issue, let's first check if the bucket exists, and if it doesn’t, we can create the bucket. Here are the commands you should execute next:
1. Check if the bucket exists:

```bash {"id":"01JBHHSD5GX3TNM3Q65GXJYBYX","interactive":"true"}
gsutil mb -p foyle-public gs://bsctl
```

```bash {"id":"01JBHHTAJADXEJXKAHWHQYH9F4","interactive":"false"}
# 1. Check if the bucket exists
gsutil ls gs://bsctl/
```

```bash {"id":"01JBHHTNGHSN1GN6PR6Z9HC5XM","interactive":"false"}
gsutil cp -r .build/web gs://bsctl/
```

```bash {"id":"01JBHHTWBDGKG5MR9MTRG985T9","interactive":"false"}
gsutil ls -r gs://bsctl/
```

So the layout isn't what I want

```bash {"id":"01JBHHW7PM7D5QZSRKD046PGAZ","interactive":"true"}
gsutil mv "gs://bsctl/web/*" gs://bsctl/
```

```bash {"id":"01JBHHZDHSGFQRD4NK561KYNXJ","interactive":"false"}
gsutil ls -r gs://bsctl/
```

```bash
gsutil mv gs://bsctl/app.wasm gs://bsctl/web/app.wasm
```

```bash {"id":"01JBHJ08471HMTP29HTHXTCC4G","interactive":"false"}
gsutil ls -r gs://bsctl/
```

```bash {"id":"01JBHHZR00RMZ5VPCGJHD7636B","interactive":"false"}
# Since the files are now confirmed to be in the bucket, the next step can be to set appropriate permissions for the files if necessary.
gsutil -m acl ch -R -u AllUsers:Reader gs://bsctl/
```

```bash
gsutil ls -r gs://bsctl/
```

```bash
cd iac
terraform apply
```

* I think I know what the problem is

* The URL is "https://storage.googleapis.com/bsctl/index.html"
* So GoApp ends up treating "index.html" as the route path and we don't have a route handler for it
* If we add a router for index.html does that fix it?

```bash {"id":"01JBHJCV8WBEMNX0V023F561AX","interactive":"true"}
rm -rf .build/*
make static
gsutil cp ".build/static/*" gs://bsctl
gsutil cp ".build/static/web/*" gs://bsctl/web/
```

```bash
gsutil ls -r gs://bsctl/
```

```bash {"id":"01JBHJKJ1FZZG6K6HCHC56397Q","interactive":"false"}
gsutil ls -r gs://bsctl
```

* Success that did it!

```bash {"id":"01JBHJT044ZXNQFPHEDC8DCCAR","interactive":"false"}
# Confirm the deployment of your application by checking the App Engine services
gcloud app services list

# You can also view the logs to ensure there are no errors during deployment
gcloud app logs tail -s default
```