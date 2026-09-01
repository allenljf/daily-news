# Daily News Go backend

`net/http` API and Cloud Run Job for the single-user Daily News App. The checked-in OpenAPI artifact in `docs/contracts/` remains the Flutter compatibility boundary.

## Verification

```sh
gofmt -w cmd internal
go vet ./...
go test -p 1 ./...
```

## Local runtime

The API requires `DATABASE_URL`, `ALLOWED_USER_EMAIL`, `GCP_PROJECT_ID`, `GCP_REGION`, and `CLOUD_RUN_JOB_NAME`. Firebase and Cloud Run use Application Default Credentials. The Job requires `DATABASE_URL` and `RUN_ID`; a scheduled execution with an unknown ID creates or merges that Taipei-day scheduled Run.

Production secrets are supplied through Cloud Run and Secret Manager; do not commit `.env` values.
