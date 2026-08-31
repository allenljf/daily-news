import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app.dart';
import 'core/http/http_providers.dart';

const _configuredApiBaseUrl = String.fromEnvironment(
  'DAILY_NEWS_API_BASE_URL',
  defaultValue: 'https://api.daily-news.invalid/v1/',
);

void main() => runApp(buildConfiguredApp(apiBaseUrl: _configuredApiBaseUrl));

Widget buildConfiguredApp({required String apiBaseUrl}) {
  return ProviderScope(
    overrides: [
      apiBaseUrlProvider.overrideWithValue(parseApiBaseUrl(apiBaseUrl)),
    ],
    child: const DailyNewsApp(),
  );
}

Uri parseApiBaseUrl(String rawValue) {
  final uri = Uri.tryParse(rawValue.trim());
  final hasExpectedPath = uri?.path == '/v1' || uri?.path == '/v1/';
  if (uri == null ||
      !uri.isScheme('https') ||
      uri.host.isEmpty ||
      uri.userInfo.isNotEmpty ||
      uri.hasQuery ||
      uri.hasFragment ||
      !hasExpectedPath) {
    throw ArgumentError(
      'DAILY_NEWS_API_BASE_URL must be an HTTPS origin ending in /v1/ '
      'without credentials, query parameters, or a fragment.',
    );
  }
  return uri.replace(path: '/v1/');
}
