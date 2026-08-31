import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../api/api_client.dart';
import '../auth/auth_providers.dart';
import 'dio_client.dart';

final apiBaseUrlProvider = Provider<Uri>(
  (ref) => throw StateError('Override apiBaseUrlProvider at app bootstrap.'),
);

// Riverpod owns this transport at the composition root; feature code consumes
// apiClientProvider or a repository provider instead.
// guide-ignore: dio-outside-remote-service
final dioProvider = Provider<Dio>(
  (ref) => createDioClient(
    baseUrl: ref.watch(apiBaseUrlProvider),
    authSession: ref.watch(authRepositoryProvider),
  ),
);

final apiClientProvider = Provider<ApiClient>(
  (ref) => ApiClient(ref.watch(dioProvider)),
);
