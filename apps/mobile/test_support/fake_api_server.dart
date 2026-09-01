import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:daily_news_mobile/core/auth/auth_gateway.dart';
import 'package:daily_news_mobile/core/auth/auth_user.dart';
import 'package:dio/dio.dart';

final class FakeDailyNewsApiServer {
  FakeDailyNewsApiServer() {
    authGateway = _FakeAuthGateway(() => signedIn = true);
    adapter = _FakeApiAdapter(this);
  }

  late final AuthGateway authGateway;
  late final HttpClientAdapter adapter;
  bool signedIn = false;
  int manualRunRequests = 0;
  bool articleAvailable = false;
  bool articleIsPermanent = false;
  bool articleDeleted = false;
  Map<String, Object?>? category;

  void close() {
    (authGateway as _FakeAuthGateway).close();
    adapter.close(force: true);
  }
}

final class _FakeAuthGateway implements AuthGateway {
  _FakeAuthGateway(this._onSignIn) {
    _controller = StreamController<AuthUser?>.broadcast(
      onListen: () => _controller.add(_user),
    );
  }

  final void Function() _onSignIn;
  late final StreamController<AuthUser?> _controller;
  AuthUser? _user;

  @override
  Stream<AuthUser?> authStateChanges() => _controller.stream;

  @override
  Future<String?> getIdToken() async =>
      _user == null ? null : 'fake-firebase-token';

  @override
  Future<void> signInWithGoogle() async {
    _user = const AuthUser(id: 'reader-1');
    _onSignIn();
    _controller.add(_user);
  }

  @override
  Future<void> signOut() async {
    _user = null;
    _controller.add(null);
  }

  void close() => _controller.close();
}

final class _FakeApiAdapter implements HttpClientAdapter {
  _FakeApiAdapter(this.server);
  final FakeDailyNewsApiServer server;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final path = options.uri.path;
    final method = options.method;

    if (options.headers['Authorization'] != 'Bearer fake-firebase-token') {
      return _json({
        'type': 'about:blank',
        'title': 'Unauthorized',
        'status': 401,
      }, statusCode: 401);
    }

    if (method == 'GET' && path == '/v1/ingestion-runs/latest') {
      return _json({'last_successful_at': null, 'active_run': null});
    }
    if (method == 'GET' && path == '/v1/categories') {
      return _json(server.category == null ? [] : [server.category]);
    }
    if (method == 'POST' && path == '/v1/categories') {
      final request = Map<String, Object?>.from(options.data as Map);
      server.category = {
        'id': 'category-1',
        'name': request['name'],
        'search_keywords': request['search_keywords'],
        'special_requirements': request['special_requirements'],
        'source_settings': const <Object?>[],
      };
      return _json(server.category!, statusCode: 201);
    }
    if (method == 'POST' && path == '/v1/ingestion-runs') {
      server.manualRunRequests += 1;
      server.articleAvailable = true;
      return _json({
        'id': 'run-1',
        'trigger': 'manual',
        'status': 'queued',
        'started_at': '2026-09-01T00:00:00Z',
        'finished_at': null,
      }, statusCode: 202);
    }
    if (method == 'GET' && path == '/v1/categories/category-1/news') {
      final items = server.articleAvailable && !server.articleDeleted
          ? [_article()]
          : const <Object?>[];
      return _json({'items': items, 'next_cursor': null});
    }
    if (method == 'GET' && path == '/v1/categories/category-1/news/article-1') {
      return _json({..._article(), 'first_seen_at': '2026-09-01T00:00:00Z'});
    }
    if (method == 'PATCH' && path == '/v1/news/article-1') {
      server.articleIsPermanent = true;
      return _json(_article());
    }
    if (method == 'DELETE' && path == '/v1/news/article-1') {
      server.articleDeleted = true;
      return ResponseBody.fromString('', 204);
    }
    return _json({
      'type': 'about:blank',
      'title': 'Not found',
      'status': 404,
      'detail': '$method $path is not configured',
    }, statusCode: 404);
  }

  Map<String, Object?> _article() => {
    'id': 'article-1',
    'title': '新的科技新聞',
    'summary': '測試摘要',
    'canonical_url': 'https://example.test/news/1',
    'published_at': '2026-09-01T00:00:00Z',
    'inserted_at': '2026-09-01T00:01:00Z',
    'expires_at': server.articleIsPermanent ? null : '2026-10-01T00:01:00Z',
    'source_tag_id': 'source-1',
    'source_tag_label': 'Example',
  };

  static ResponseBody _json(Object? value, {int statusCode = 200}) =>
      ResponseBody.fromString(
        jsonEncode(value),
        statusCode,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );

  @override
  void close({bool force = false}) {}
}
