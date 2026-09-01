# Daily News GitHub Actions variables

Create these as repository-level GitHub **Variables**, not Actions secrets.
They are deployment identifiers and do not contain credentials. Workflows use
GitHub's short-lived OIDC token to authenticate through Workload Identity
Federation (WIF); they must never use a service-account JSON key.

| Variable | Value format | Used by | Notes |
| --- | --- | --- | --- |
| `GCP_PROJECT_ID` | Google Cloud project ID | deploy and daily-ingestion workflows | The project hosting Cloud Run, Cloud SQL and Secret Manager |
| `GCP_REGION` | Cloud Run region, for example `asia-east1` | deploy and daily-ingestion workflows | Keep service, Job and Artifact Registry in the chosen region where practical |
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | Full provider resource name: `projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/POOL_ID/providers/PROVIDER_ID` | deploy and daily-ingestion workflows | Use the project **number** in this resource name |
| `GCP_SERVICE_ACCOUNT` | Deployer service-account email | deploy and daily-ingestion workflows | A dedicated GitHub deployment identity, not either runtime identity |
| `CLOUD_RUN_JOB_NAME` | Cloud Run Job resource name | deploy and daily-ingestion workflows | The Job that runs the Go daily-news Job binary with `RUN_ID` |

## Workflow requirements

Every workflow that accesses Google Cloud must contain:

```yaml
permissions:
  contents: read
  id-token: write
```

Authenticate with `google-github-actions/auth` using
`GCP_WORKLOAD_IDENTITY_PROVIDER` and `GCP_SERVICE_ACCOUNT`. Keep the provider
and service-account values as variables. Do not add `GOOGLE_CREDENTIALS`, a
base64-encoded key file, `GCP_SA_KEY`, database credentials, LLM keys, or any
other server credential to GitHub Variables or Actions secrets.

## WIF trust boundary

Create a dedicated workload identity pool and GitHub OIDC provider. Map at
least `google.subject=assertion.sub` and map repository attributes needed by
the policy. Apply an attribute condition that restricts the provider to this
repository owner; additionally restrict the repository and, where appropriate,
the protected deployment branch or environment. Grant the GitHub federated
principal `roles/iam.workloadIdentityUser` on the dedicated deployer service
account, then grant that account only the deployment permissions documented in
[gcp-setup.md](gcp-setup.md).

## What you must provide

Before the O3 workflow task, you will need the project ID and number, chosen
region, GitHub owner/repository, deployment branch/environment, WIF pool and
provider IDs, deployer service-account email, and Cloud Run Job name. Enter
only the resulting identifiers above; do not enter secret values.
