import 'dart:convert';
import 'dart:typed_data';

import 'package:daily_news_mobile/core/api/api_client.dart';
import 'package:daily_news_mobile/core/api/api_failure.dart';
import 'package:daily_news_mobile/core/api/cursor_page_dto.dart';
import 'package:daily_news_mobile/core/auth/auth_session.dart';
import 'package:daily_news_mobile/core/http/dio_client.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('authenticated Dio client', () {
    test('adds the current Firebase ID token as a Bearer header', () async {
      final authSession = _FakeAuthSession('fixture-firebase-credential');
      final adapter = _RecordingAdapter.json(statusCode: 200, body: const {});
      final dio = createDioClient(
        baseUrl: Uri.parse('https://api.example.test/v1/'),
        authSession: authSession,
      )..httpClientAdapter = adapter;

      await dio.get<Map<String, Object?>>('categories');

      expect(
        adapter.lastRequest?.headers['Authorization'],
        'Bearer fixture-firebase-credential',
      );
    });

    test('does not send the token to a different absolute origin', () async {
      final authSession = _FakeAuthSession('fixture-firebase-credential');
      final adapter = _RecordingAdapter.json(statusCode: 200, body: const {});
      final dio = createDioClient(
        baseUrl: Uri.parse('https://api.example.test/v1/'),
        authSession: authSession,
      )..httpClientAdapter = adapter;

      await dio.get<Map<String, Object?>>(
        'https://untrusted.example.test/collect',
      );

      expect(adapter.lastRequest?.headers, isNot(contains('Authorization')));
    });

    test('signs out when the API rejects the session with 401', () async {
      final authSession = _FakeAuthSession('fixture-expired-credential');
      final adapter = _RecordingAdapter.json(
        statusCode: 401,
        body: const {
          'detail': {'title': 'Authentication required', 'status': 401},
        },
        contentType: Headers.jsonContentType,
      );
      final dio = createDioClient(
        baseUrl: Uri.parse('https://api.example.test/v1/'),
        authSession: authSession,
      )..httpClientAdapter = adapter;

      await expectLater(
        dio.get<Object?>('categories'),
        throwsA(isA<DioException>()),
      );

      expect(authSession.signOutCount, 1);
    });
  });

  test('translates an API Problem Details response into ApiFailure', () async {
    final authSession = _FakeAuthSession('fixture-valid-credential');
    final dio =
        createDioClient(
            baseUrl: Uri.parse('https://api.example.test/v1/'),
            authSession: authSession,
          )
          ..httpClientAdapter = _RecordingAdapter.json(
            statusCode: 404,
            body: const {
              'detail': {'title': 'Article not found', 'status': 404},
            },
            contentType: 'application/problem+json',
          );
    final client = ApiClient(dio);

    await expectLater(
      client.getJson('news/missing'),
      throwsA(
        isA<ApiFailure>()
            .having((failure) => failure.statusCode, 'statusCode', 404)
            .having((failure) => failure.title, 'title', 'Article not found'),
      ),
    );
  });

  test(
    'keeps a non-Problem HTTP response distinct from network failure',
    () async {
      final authSession = _FakeAuthSession('fixture-valid-credential');
      final dio =
          createDioClient(
              baseUrl: Uri.parse('https://api.example.test/v1/'),
              authSession: authSession,
            )
            ..httpClientAdapter = _RecordingAdapter.json(
              statusCode: 503,
              body: const {'message': 'Temporarily unavailable'},
            );
      final client = ApiClient(dio);

      await expectLater(
        client.getJson('news'),
        throwsA(
          isA<ApiFailure>()
              .having((failure) => failure.kind, 'kind', ApiFailureKind.http)
              .having((failure) => failure.statusCode, 'statusCode', 503),
        ),
      );
    },
  );

  test('decodes a cursor page DTO with its next cursor', () {
    final page = CursorPageDto<_ItemDto>.fromJson(const {
      'items': [
        {'id': 'article-1', 'title': 'A useful headline'},
      ],
      'next_cursor': 'opaque-next-page',
    }, (json) => _ItemDto.fromJson(json! as Map<String, Object?>));

    expect(page.items, const [
      _ItemDto(id: 'article-1', title: 'A useful headline'),
    ]);
    expect(page.nextCursor, 'opaque-next-page');
  });
}

final class _FakeAuthSession implements AuthSession {
  _FakeAuthSession(this.token);

  final String? token;
  int signOutCount = 0;

  @override
  Future<String?> getIdToken() async => token;

  @override
  Future<void> signOut() async {
    signOutCount += 1;
  }
}

final class _RecordingAdapter implements HttpClientAdapter {
  _RecordingAdapter.json({
    required this.statusCode,
    required this.body,
    this.contentType = Headers.jsonContentType,
  });

  final int statusCode;
  final Object body;
  final String contentType;
  RequestOptions? lastRequest;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    lastRequest = options;
    return ResponseBody.fromString(
      jsonEncode(body),
      statusCode,
      headers: {
        Headers.contentTypeHeader: [contentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

final class _ItemDto {
  const _ItemDto({required this.id, required this.title});

  factory _ItemDto.fromJson(Map<String, Object?> json) =>
      _ItemDto(id: json['id']! as String, title: json['title']! as String);

  final String id;
  final String title;

  @override
  bool operator ==(Object other) =>
      other is _ItemDto && other.id == id && other.title == title;

  @override
  int get hashCode => Object.hash(id, title);
}
