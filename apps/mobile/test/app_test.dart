import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:daily_news_mobile/app.dart';

void main() {
  testWidgets('shows a protected loading screen while the session is checked', (
    tester,
  ) async {
    await tester.pumpWidget(const ProviderScope(child: DailyNewsApp()));

    expect(find.text('正在載入…'), findsOneWidget);
    expect(find.byType(CircularProgressIndicator), findsOneWidget);
  });
}
