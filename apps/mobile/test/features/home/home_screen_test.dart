import 'package:daily_news_mobile/app.dart';
import 'package:daily_news_mobile/core/auth/auth_providers.dart';
import 'package:daily_news_mobile/core/auth/auth_user.dart';
import 'package:daily_news_mobile/features/home/application/home_providers.dart';
import 'package:daily_news_mobile/features/home/data/home_refresh_status.dart';
import 'package:daily_news_mobile/features/home/data/home_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('home route shows never updated and enables immediate update', (
    tester,
  ) async {
    await _pumpHome(
      tester,
      const HomeRefreshStatus(
        lastSuccessfulAt: null,
        activeRunStatus: ActiveRunStatus.none,
      ),
    );

    expect(find.text('尚未更新'), findsOneWidget);
    expect(_immediateUpdateButton(tester).onPressed, isNotNull);
  });

  testWidgets('home route formats the successful update time for its locale', (
    tester,
  ) async {
    await _pumpHome(
      tester,
      HomeRefreshStatus(
        lastSuccessfulAt: DateTime(2026, 8, 30, 8, 1),
        activeRunStatus: ActiveRunStatus.none,
      ),
    );

    expect(find.text('2026/8/30 08:01'), findsOneWidget);
  });

  for (final fixture in const [
    (status: ActiveRunStatus.queued, label: '排隊中'),
    (status: ActiveRunStatus.running, label: '更新中'),
  ]) {
    testWidgets(
      'home disables immediate update while a run is ${fixture.status.name}',
      (tester) async {
        await _pumpHome(
          tester,
          HomeRefreshStatus(
            lastSuccessfulAt: null,
            activeRunStatus: fixture.status,
          ),
        );

        expect(find.text(fixture.label), findsOneWidget);
        expect(_immediateUpdateButton(tester).onPressed, isNull);
      },
    );
  }
}

Future<void> _pumpHome(WidgetTester tester, HomeRefreshStatus status) async {
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        authStateProvider.overrideWith(
          (ref) => Stream.value(const AuthUser(id: 'reader-1')),
        ),
        homeRepositoryProvider.overrideWithValue(_FakeHomeRepository(status)),
      ],
      child: const DailyNewsApp(locale: Locale('zh')),
    ),
  );
  await tester.pump();
  await tester.pump();
}

FilledButton _immediateUpdateButton(WidgetTester tester) {
  return tester.widget<FilledButton>(find.widgetWithText(FilledButton, '立即更新'));
}

final class _FakeHomeRepository implements HomeRepository {
  const _FakeHomeRepository(this.status);

  final HomeRefreshStatus status;

  @override
  Future<HomeRefreshStatus> loadRefreshStatus() async => status;
}
