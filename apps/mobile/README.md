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

## Android setup

Android uses the application id `com.allenljf.dailynews`, matching the iOS
bundle id. `Firebase.initializeApp()` reads `android/app/google-services.json`
through the `com.google.gms.google-services` Gradle plugin, and
`google_sign_in` reads the web client id from the generated
`default_web_client_id` resource to obtain a Firebase ID token.

Because `Firebase.initializeApp()` needs that file, the Android build fails
until it is present. To create it:

1. In the Firebase Console open project `daily-news-93f7b`, add an Android app
   with package name `com.allenljf.dailynews`.
2. Register the signing certificate SHA-1. For a local debug build:

   ```bash
   keytool -exportcert -alias androiddebugkey \
     -keystore ~/.android/debug.keystore -storepass android 2>/dev/null \
     | openssl x509 -inform DER -noout -fingerprint -sha1
   ```

3. Download `google-services.json` and place it at
   `apps/mobile/android/app/google-services.json`.
4. Run on an Android device or emulator:

   ```bash
   flutter run \
     --dart-define=DAILY_NEWS_API_BASE_URL=https://api.example.com/v1/
   ```

The client config is public client configuration, not a server secret. Only the
web client id produced by enabling the Google provider is needed for sign-in.
