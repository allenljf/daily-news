import 'package:dio/dio.dart';

import '../auth/auth_session.dart';

Dio createDioClient({required Uri baseUrl, required AuthSession authSession}) {
  if (!baseUrl.isScheme('https') || baseUrl.host.isEmpty) {
    throw ArgumentError.value(baseUrl, 'baseUrl', 'must be an HTTPS URL');
  }

  final dio = Dio(
    BaseOptions(
      baseUrl: baseUrl.toString(),
      connectTimeout: const Duration(seconds: 15),
      sendTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 30),
      headers: const {Headers.acceptHeader: Headers.jsonContentType},
    ),
  );
  dio.interceptors.add(_FirebaseAuthInterceptor(authSession));
  return dio;
}

final class _FirebaseAuthInterceptor extends Interceptor {
  _FirebaseAuthInterceptor(this._authSession);

  final AuthSession _authSession;

  @override
  void onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final token = await _authSession.getIdToken();
    if (token != null && token.isNotEmpty) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  void onError(DioException error, ErrorInterceptorHandler handler) async {
    if (error.response?.statusCode == 401) {
      try {
        await _authSession.signOut();
      } on Object {
        // Preserve the original authentication rejection for the caller.
      }
    }
    handler.next(error);
  }
}
