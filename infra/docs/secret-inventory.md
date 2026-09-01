# Daily News secret inventory

This inventory names every runtime setting without recording its value. Store
secrets in Google Secret Manager and inject only the secrets required by the
Cloud Run service or Cloud Run Job. Do not commit values to this repository,
GitHub Variables, Flutter `--dart-define` arguments, container images, logs,
or issue/PR text.

| Setting | Required when | Store / inject into | Purpose | Must not be placed in |
| --- | --- | --- | --- | --- |
| `ALLOWED_USER_EMAIL` | Always | Secret Manager → API service environment variable | The one Firebase-authenticated email the API permits | Flutter app, GitHub Variables, image, logs |
| `DATABASE_URL` | Always | Secret Manager → API service and ingestion Job environment variables | PostgreSQL connection URL consumed through Go `database/sql` + pgx adapter | Git, image, GitHub Variables, logs |
| `DB_PASSWORD` | When Cloud SQL uses password authentication | Secret Manager; use it only to construct or rotate `DATABASE_URL` | Cloud SQL database-password credential | Git, image, GitHub Variables, Flutter, logs |
| `GEMINI_API_KEY` | Only when Gemini Developer API is selected instead of Vertex AI IAM | Secret Manager → ingestion Job environment variable | Gemini grounding requests | Git, image, GitHub Variables, Flutter, logs |
| `GITHUB_NEWS_TOKEN` | Only for higher GitHub API quota or private resources | Secret Manager → ingestion Job environment variable | GitHub adapter authorization | Git, image, GitHub Variables, Flutter, logs |
| `YOUTUBE_API_KEY` | Only when the YouTube adapter is enabled | Secret Manager → ingestion Job environment variable | YouTube Data API requests | Git, image, GitHub Variables, Flutter, logs |
| `META_ACCESS_TOKEN` | Only when an authorized Meta adapter is enabled | Secret Manager → ingestion Job environment variable | Meta platform API authorization | Git, image, GitHub Variables, Flutter, logs |

Firebase client configuration is public client configuration, not a server
secret. Firebase Admin on Cloud Run uses Application Default Credentials; do
not create or upload a Firebase Admin service-account JSON key for this app.

## Secret Manager procedure

1. Enable Secret Manager and create one secret for each required row.
2. Add the real value directly in the Google Cloud Console or `gcloud` from a
   trusted terminal; never paste it into this repository or a workflow file.
3. Grant `roles/secretmanager.secretAccessor` on each individual secret to
   only the Cloud Run identity that needs it. The service normally needs
   `ALLOWED_USER_EMAIL` and `DATABASE_URL`; the Job also needs
   `DATABASE_URL` and only the enabled adapter credentials.
4. Configure Cloud Run to reference a specific secret version as an
   environment variable. Rotate by adding a new version and redeploying the
   affected service or Job.

`DB_PASSWORD` is intentionally listed separately for rotation and inventory
purposes. Runtime code consumes `DATABASE_URL`; do not add a second password
parsing path unless a later backend task explicitly needs one.

## What you must provide

Before staging deployment, you will need to create the actual secret values
and choose which optional adapters to enable. No value is needed while these
documents are being added.
