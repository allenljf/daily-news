import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:daily_news_mobile/core/auth/auth_gateway.dart';
import 'package:daily_news_mobile/core/auth/auth_providers.dart';
import 'package:daily_news_mobile/core/auth/auth_user.dart';
import 'package:daily_news_mobile/core/http/http_providers.dart';
import 'package:daily_news_mobile/main.dart' as bootstrap;
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:daily_news_mobile/app.dart';

void main() {
  testWidgets('shows a protected loading screen while the session is checked', (
    tester,
  ) async {
    final pendingSession = Completer<AuthUser?>();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authStateProvider.overrideWith(
            (ref) => pendingSession.future.asStream(),
          ),
        ],
        child: const DailyNewsApp(),
      ),
    );

    expect(find.text('正在載入…'), findsOneWidget);
    expect(find.byType(CircularProgressIndicator), findsOneWidget);
  });

  testWidgets('production app supplies its normalized API base URL override', (
    tester,
  ) async {
    await tester.pumpWidget(
      bootstrap.buildConfiguredApp(apiBaseUrl: 'https://api.example.test/v1'),
    );

    final appContext = tester.element(find.byType(DailyNewsApp));
    final container = ProviderScope.containerOf(appContext);
    expect(
      container.read(apiBaseUrlProvider),
      Uri.parse('https://api.example.test/v1/'),
    );
  });

  testWidgets(
    'production bootstrap constructs Home with its configured API base URL',
    (tester) async {
      final adapter = _LatestRunAdapter();

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            apiBaseUrlProvider.overrideWithValue(
              bootstrap.parseApiBaseUrl('https://api.example.test/v1'),
            ),
            authGatewayProvider.overrideWithValue(_AuthenticatedGateway()),
            httpClientAdapterProvider.overrideWithValue(adapter),
          ],
          child: const DailyNewsApp(),
        ),
      );
      await tester.pumpAndSettle();

      expect(
        adapter.requests.map((request) => request.uri),
        contains(
          Uri.parse('https://api.example.test/v1/ingestion-runs/latest'),
        ),
      );
      expect(find.text('尚未更新'), findsOneWidget);
    },
  );

  test('production bootstrap rejects a non-HTTPS API base URL', () {
    expect(
      () => bootstrap.parseApiBaseUrl('http://api.example.test/v1/'),
      throwsArgumentError,
    );
  });

  test('production bootstrap rejects credentials in the API base URL', () {
    expect(
      () => bootstrap.parseApiBaseUrl(
        'https://reader:secret@api.example.test/v1/',
      ),
      throwsArgumentError,
    );
  });
}

final class _AuthenticatedGateway implements AuthGateway {
  @override
  Stream<AuthUser?> authStateChanges() =>
      Stream.value(const AuthUser(id: 'reader-1'));

  @override
  Future<String?> getIdToken() async => 'fixture-firebase-token';

  @override
  Future<void> signInWithGoogle() async {}

  @override
  Future<void> signOut() async {}
}

final class _LatestRunAdapter implements HttpClientAdapter {
  final requests = <RequestOptions>[];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    return ResponseBody.fromString(
      jsonEncode(const {'last_successful_at': null, 'active_run': null}),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
