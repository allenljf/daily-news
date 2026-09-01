# Daily News GCP setup

This runbook prepares the non-secret GCP resources used by the Daily News API
and daily ingestion Job. Substitute values in the Google Cloud Console or
your local terminal; do not put real project IDs, email addresses, tokens, or
secret values in this repository.

## 1. Create the project resources

1. Choose a GCP project and Cloud Run region. Enable Cloud Run, Artifact
   Registry, Cloud SQL Admin, Secret Manager, IAM Credentials, and Security
   Token Service APIs.
2. Create a regional Artifact Registry Docker repository for the shared API /
   Job image.
3. Create a PostgreSQL Cloud SQL instance and a `daily_news` database. Create
   a least-privilege database user for the application, then store its
   connection information as `DATABASE_URL` in Secret Manager. Store
   `DB_PASSWORD` separately when password authentication is used.
4. Associate or create the Firebase project used by the Flutter app. Enable
   Google Sign-In in Firebase Authentication and register the Android/iOS
   clients. The app's Firebase client configuration may be distributed with
   the mobile app; Firebase Admin credentials may not.
5. Create the Secret Manager entries listed in
   [secret-inventory.md](secret-inventory.md). Do not enable an optional
   adapter until its credential exists and the adapter is authorized to access
   its platform.

## 2. Create three service accounts

Use separate user-managed identities so runtime permissions cannot be used to
deploy and the API cannot receive source-adapter credentials unnecessarily.

| Identity | Used by | Minimum responsibilities |
| --- | --- | --- |
| API runtime service account | Cloud Run Go `net/http` service | Access its own `DATABASE_URL` and `ALLOWED_USER_EMAIL` secrets; connect to Cloud SQL; use ADC for Firebase Admin token verification |
| Ingestion runtime service account | Cloud Run Job | Access `DATABASE_URL` plus only enabled adapter secrets; connect to Cloud SQL; emit Job logs |
| GitHub deployer service account | GitHub Actions via WIF | Deploy/update the Cloud Run service and Job, read the image from Artifact Registry, and execute the Job; it is not a runtime identity |

Bind secret access at the secret level, not project-wide. Grant the Cloud SQL
Client role (`roles/cloudsql.client`) only to the API and Job runtime service
accounts if using the Cloud SQL connector or socket. Grant Secret Manager
Secret Accessor (`roles/secretmanager.secretAccessor`) to each runtime identity
only for the secret rows it consumes. Cloud Run's recommended runtime model is
Application Default Credentials, so do not set `GOOGLE_APPLICATION_CREDENTIALS`
to an exported key file in Cloud Run.

## 3. Grant deployment and execution permissions

Grant the GitHub deployer service account narrowly:

| Resource / action | Minimum role |
| --- | --- |
| Create/update the named Cloud Run service and Job | Cloud Run Developer (`roles/run.developer`) scoped to the resource where possible |
| Use each Cloud Run runtime service account during deployment | Service Account User (`roles/iam.serviceAccountUser`) on that runtime account |
| Read the shared image | Artifact Registry Reader (`roles/artifactregistry.reader`) on its repository |
| Trigger the named daily ingestion Job | Cloud Run Invoker (`roles/run.invoker`) on that Job |
| Impersonate the deployer through GitHub WIF | Workload Identity User (`roles/iam.workloadIdentityUser`) on the deployer account, granted to the restricted federated principal |

The exact Cloud Run deployment role may need the corresponding create/update
permissions if resources do not exist yet. Do not grant Owner, Editor, or
project-wide Secret Accessor merely to simplify setup. Give a human operator
the administrative roles required to initially create the WIF pool/provider
and service accounts, then remove elevated setup access when no longer needed.

## 4. Configure WIF for GitHub Actions

1. Create a workload identity pool and GitHub OIDC provider in the selected
   project.
2. Map `google.subject` from `assertion.sub`; map GitHub repository attributes
   used in the policy.
3. Add an attribute condition that permits only the intended GitHub owner and
   repository. Also restrict the deployment workflow to its protected branch
   or GitHub Environment when applicable.
4. Grant the restricted federated principal
   `roles/iam.workloadIdentityUser` on the GitHub deployer service account.
5. Populate only the identifiers in
   [github-variables.md](github-variables.md). A workflow requests OIDC with
   `id-token: write` and uses the provider to impersonate the deployer.

## 5. Deploy-time configuration boundary

The Go replacement phase will build one image used by both Cloud Run resources.
The service starts the Go API binary; the Job runs the Go daily-news Job binary
with `RUN_ID` as its explicit input. The Cloud Run resource names, runtime
service accounts, Secret Manager names, and WIF boundary stay unchanged.
Attach the correct runtime service account and only its referenced secrets to
each resource at deployment time. Do not bake `.env`, Firebase Admin keys,
database URLs, or adapter credentials into the image.

## Operator checklist

You must complete these external steps before a real staging deployment:

- Choose the GCP project, region, Firebase project and GitHub repository
  boundary.
- Create Cloud SQL, Artifact Registry, Secret Manager entries, and the three
  service accounts above.
- Supply the real `ALLOWED_USER_EMAIL` and `DATABASE_URL`; decide which
  optional adapter credentials are needed.
- Configure WIF and repository Variables according to
  [github-variables.md](github-variables.md).

The Go replacement phase will update image build and command wiring while
retaining these documented names. The O4 smoke test waits for Go parity, then
requires your allowlisted account and the staging resources; it never asks you
to put their secrets in this repository.

## References

- [Cloud Run service identities and IAM roles](https://cloud.google.com/run/docs/reference/iam/roles)
- [Cloud Run Job execution permissions](https://cloud.google.com/run/docs/execute/jobs)
- [Cloud Run Secret Manager integration](https://cloud.google.com/run/docs/configuring/services/secrets)
- [Google Cloud WIF for deployment pipelines](https://cloud.google.com/iam/docs/workload-identity-federation-with-deployment-pipelines)
- [Firebase Admin SDK on Google Cloud](https://firebase.google.com/docs/admin/setup)
