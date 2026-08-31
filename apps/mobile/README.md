# Daily News mobile

Provide the public backend base URL at build or run time with a Dart define:

```bash
flutter run \
  --dart-define=DAILY_NEWS_API_BASE_URL=https://api.example.com/v1/
```

The value must use HTTPS, end in `/v1/`, and contain no credentials, query
parameters, or fragment. It is a public client endpoint, not a secret. When no
value is supplied, the app uses the reserved non-routable
`https://api.daily-news.invalid/v1/` placeholder and displays its localized
network error state instead of failing during provider construction.
