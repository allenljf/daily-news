import 'dart:async';

import 'package:daily_news_mobile/core/auth/auth_providers.dart';
import 'package:daily_news_mobile/core/auth/auth_user.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

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
}
