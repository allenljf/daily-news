import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../api/api_client.dart';
import '../auth/auth_providers.dart';
import 'dio_client.dart';

final apiBaseUrlProvider = Provider<Uri>(
  (ref) => throw StateError('Override apiBaseUrlProvider at app bootstrap.'),
);

// Tests replace only the socket boundary while retaining the production
// interceptor, API client, service, repository, and controller wiring.
// guide-ignore: dio-outside-remote-service
final httpClientAdapterProvider = Provider<HttpClientAdapter?>((ref) => null);

// Riverpod owns this transport at the composition root; feature code consumes
// apiClientProvider or a repository provider instead.
// guide-ignore: dio-outside-remote-service
final dioProvider = Provider<Dio>((ref) {
  final dio = createDioClient(
    baseUrl: ref.watch(apiBaseUrlProvider),
    authSession: ref.watch(authRepositoryProvider),
  );
  final adapter = ref.watch(httpClientAdapterProvider);
  if (adapter != null) {
    dio.httpClientAdapter = adapter;
  }
  return dio;
});

final apiClientProvider = Provider<ApiClient>(
  (ref) => ApiClient(ref.watch(dioProvider)),
);
