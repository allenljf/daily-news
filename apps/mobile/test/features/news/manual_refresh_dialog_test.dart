import 'package:daily_news_mobile/features/news/application/manual_run_providers.dart';
import 'package:daily_news_mobile/features/news/data/manual_run.dart';
import 'package:daily_news_mobile/features/news/data/manual_run_repository.dart';
import 'package:daily_news_mobile/features/news/presentation/manual_refresh_control.dart';
import 'package:daily_news_mobile/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('confirmation explains background work and shows queued state', (
    tester,
  ) async {
    final repository = _FakeManualRunRepository();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [manualRunRepositoryProvider.overrideWithValue(repository)],
        child: MaterialApp(
          locale: const Locale('zh'),
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          supportedLocales: AppLocalizations.supportedLocales,
          home: const Scaffold(body: ManualRefreshControl()),
        ),
      ),
    );

    await tester.tap(find.text('立即更新'));
    await tester.pumpAndSettle();
    expect(find.textContaining('後端背景擷取工作'), findsOneWidget);
    expect(find.textContaining('可以離開 App'), findsOneWidget);

    await tester.tap(find.widgetWithText(FilledButton, '確認更新'));
    await tester.pumpAndSettle();

    expect(repository.requests, 1);
    expect(find.text('排隊中'), findsOneWidget);
    expect(
      tester
          .widget<FilledButton>(find.widgetWithText(FilledButton, '立即更新'))
          .onPressed,
      isNull,
    );
  });
}

final class _FakeManualRunRepository implements ManualRunRepository {
  int requests = 0;

  @override
  Future<ManualRun> requestManualRun() async {
    requests += 1;
    return const ManualRun(id: 'run-1', status: ManualRunStatus.queued);
  }
}
