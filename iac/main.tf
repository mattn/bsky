resource "google_storage_bucket_iam_member" "public_access" {
  bucket = "bsctl"
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}
