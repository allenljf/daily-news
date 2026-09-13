import 'package:daily_news_mobile/app.dart';
import 'package:daily_news_mobile/core/auth/auth_providers.dart';
import 'package:daily_news_mobile/core/auth/auth_user.dart';
import 'package:daily_news_mobile/features/categories/application/category_providers.dart';
import 'package:daily_news_mobile/features/categories/data/category.dart';
import 'package:daily_news_mobile/features/categories/data/category_repository.dart';
import 'package:daily_news_mobile/features/home/application/home_providers.dart';
import 'package:daily_news_mobile/features/home/data/home_refresh_status.dart';
import 'package:daily_news_mobile/features/home/data/home_repository.dart';
import 'package:daily_news_mobile/features/news/application/manual_run_providers.dart';
import 'package:daily_news_mobile/features/news/data/manual_run.dart';
import 'package:daily_news_mobile/features/news/data/manual_run_repository.dart';
import 'package:daily_news_mobile/features/home/presentation/home_screen.dart';
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

  testWidgets('home keeps update details and action in one title-free row', (
    tester,
  ) async {
    await _pumpHome(
      tester,
      const HomeRefreshStatus(
        lastSuccessfulAt: null,
        activeRunStatus: ActiveRunStatus.none,
      ),
    );

    expect(find.text('每日新聞'), findsNothing);
    final updateRow = find.ancestor(
      of: find.text('最近更新時間'),
      matching: find.byType(Row),
    );
    expect(updateRow, findsOneWidget);
    expect(
      find.descendant(of: updateRow, matching: find.text('立即更新')),
      findsOneWidget,
    );
  });

  testWidgets('home content is kept below the system safe area', (
    tester,
  ) async {
    await _pumpHome(
      tester,
      const HomeRefreshStatus(
        lastSuccessfulAt: null,
        activeRunStatus: ActiveRunStatus.none,
      ),
    );

    expect(find.byType(SafeArea), findsOneWidget);
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

  for (final status in const [
    ActiveRunStatus.queued,
    ActiveRunStatus.running,
  ]) {
    testWidgets(
      'refresh page stays enabled while a run is ${status.name} and reloads home and categories without requesting a run',
      (tester) async {
        final homeRepository = _FakeHomeRepository(
          HomeRefreshStatus(lastSuccessfulAt: null, activeRunStatus: status),
        );
        final categoryRepository = _FakeCategoryRepository();
        final manualRunRepository = _FakeManualRunRepository();

        await _pumpHome(
          tester,
          homeRepository.status,
          homeRepository: homeRepository,
          categoryRepository: categoryRepository,
          manualRunRepository: manualRunRepository,
        );

        expect(_immediateUpdateButton(tester).onPressed, isNull);
        expect(
          tester
              .widget<OutlinedButton>(find.byKey(homeDataRefreshButtonKey))
              .onPressed,
          isNotNull,
        );

        await tester.tap(find.byKey(homeDataRefreshButtonKey));
        await tester.pump();
        await tester.pump();

        expect(homeRepository.loadCount, 2);
        expect(categoryRepository.loadCount, 2);
        expect(manualRunRepository.requestCount, 0);
      },
    );
  }
}

Future<void> _pumpHome(
  WidgetTester tester,
  HomeRefreshStatus status, {
  _FakeHomeRepository? homeRepository,
  _FakeCategoryRepository? categoryRepository,
  _FakeManualRunRepository? manualRunRepository,
}) async {
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        authStateProvider.overrideWith(
          (ref) => Stream.value(const AuthUser(id: 'reader-1')),
        ),
        homeRepositoryProvider.overrideWithValue(
          homeRepository ?? _FakeHomeRepository(status),
        ),
        categoryRepositoryProvider.overrideWithValue(
          categoryRepository ?? _FakeCategoryRepository(),
        ),
        manualRunRepositoryProvider.overrideWithValue(
          manualRunRepository ?? _FakeManualRunRepository(),
        ),
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
  _FakeHomeRepository(this.status);

  final HomeRefreshStatus status;
  var loadCount = 0;

  @override
  Future<HomeRefreshStatus> loadRefreshStatus() async {
    loadCount++;
    return status;
  }
}

final class _FakeCategoryRepository implements CategoryRepository {
  var loadCount = 0;

  @override
  Future<Category> createCategory(CategoryDraft draft) {
    throw UnimplementedError();
  }

  @override
  Future<List<Category>> loadCategories() async {
    loadCount++;
    return const [];
  }
}

final class _FakeManualRunRepository implements ManualRunRepository {
  var requestCount = 0;

  @override
  Future<ManualRun> requestManualRun() async {
    requestCount++;
    return const ManualRun(id: 'run-1', status: ManualRunStatus.queued);
  }
}
