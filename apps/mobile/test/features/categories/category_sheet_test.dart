import 'package:daily_news_mobile/app.dart';
import 'package:daily_news_mobile/core/auth/auth_providers.dart';
import 'package:daily_news_mobile/core/auth/auth_user.dart';
import 'package:daily_news_mobile/features/categories/application/category_providers.dart';
import 'package:daily_news_mobile/features/categories/data/category.dart';
import 'package:daily_news_mobile/features/categories/data/category_repository.dart';
import 'package:daily_news_mobile/features/categories/presentation/category_grid.dart';
import 'package:daily_news_mobile/features/categories/presentation/category_settings_sheet.dart';
import 'package:daily_news_mobile/features/home/application/home_providers.dart';
import 'package:daily_news_mobile/features/home/data/home_refresh_status.dart';
import 'package:daily_news_mobile/features/home/data/home_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('an empty Category area shows only the add tile', (tester) async {
    await _pumpHome(tester, _FakeCategoryRepository());

    expect(find.byKey(addCategoryTileKey), findsOneWidget);
    expect(find.byType(CategoryTile), findsNothing);
    expect(find.byIcon(Icons.add), findsOneWidget);
  });

  testWidgets('Category tiles are followed by the add tile', (tester) async {
    await _pumpHome(
      tester,
      _FakeCategoryRepository(
        categories: [
          Category(id: 'category-1', name: 'AI'),
          Category(id: 'category-2', name: '台灣'),
        ],
      ),
    );

    expect(find.byType(CategoryTile), findsNWidgets(2));
    expect(find.text('AI'), findsOneWidget);
    expect(find.text('台灣'), findsOneWidget);
    expect(find.byKey(addCategoryTileKey), findsOneWidget);
  });

  testWidgets('add tile opens the complete Category settings sheet', (
    tester,
  ) async {
    await _pumpHome(tester, _FakeCategoryRepository());

    await tester.tap(find.byKey(addCategoryTileKey));
    await tester.pumpAndSettle();

    expect(find.text('新聞類別'), findsOneWidget);
    expect(find.text('搜尋關鍵字'), findsOneWidget);
    expect(find.text('搜尋網站'), findsOneWidget);
    expect(find.text('未指定網站'), findsOneWidget);
    expect(find.text('其他特殊需求'), findsOneWidget);
    expect(find.text('儲存設定'), findsOneWidget);
  });

  testWidgets('add-source dialog appends an editable Source Setting', (
    tester,
  ) async {
    await _pumpHome(tester, _FakeCategoryRepository());
    await _openSettingsSheet(tester);

    await tester.tap(find.byKey(addSourceSettingButtonKey));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(addSourceDialogFieldKey), 'github.com');
    await tester.tap(find.widgetWithText(FilledButton, '新增'));
    await tester.pumpAndSettle();

    expect(find.byKey(sourceSettingFieldKey(0)), findsOneWidget);
    expect(find.byKey(sourceSettingFieldKey(1)), findsOneWidget);
    expect(find.text('github.com'), findsOneWidget);
  });

  testWidgets('save omits blank sources, closes, and refreshes Categories', (
    tester,
  ) async {
    final repository = _FakeCategoryRepository();
    await _pumpHome(tester, repository);
    await _openSettingsSheet(tester);

    await tester.enterText(find.byKey(categoryNameFieldKey), 'Developer News');
    await tester.enterText(find.byKey(sourceSettingFieldKey(0)), '   ');
    await tester.tap(find.byKey(addSourceSettingButtonKey));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(addSourceDialogFieldKey), 'github.com');
    await tester.tap(find.widgetWithText(FilledButton, '新增'));
    await tester.pumpAndSettle();

    await tester.ensureVisible(find.byKey(saveCategoryButtonKey));
    await tester.tap(find.byKey(saveCategoryButtonKey));
    await tester.pumpAndSettle();

    expect(repository.createCalls, 1);
    expect(repository.lastDraft?.sourceSettings, const [
      SourceSettingDraft(
        label: 'github.com',
        websiteInput: 'github.com',
        kind: 'website',
      ),
    ]);
    expect(find.byType(CategorySettingsSheet), findsNothing);
    expect(find.text('Developer News'), findsOneWidget);
  });
}

Future<void> _pumpHome(
  WidgetTester tester,
  _FakeCategoryRepository categoryRepository,
) async {
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        authStateProvider.overrideWith(
          (ref) => Stream.value(const AuthUser(id: 'reader-1')),
        ),
        homeRepositoryProvider.overrideWithValue(_FakeHomeRepository()),
        categoryRepositoryProvider.overrideWithValue(categoryRepository),
      ],
      child: const DailyNewsApp(locale: Locale('zh')),
    ),
  );
  await tester.pumpAndSettle();
}

Future<void> _openSettingsSheet(WidgetTester tester) async {
  await tester.tap(find.byKey(addCategoryTileKey));
  await tester.pumpAndSettle();
}

final class _FakeHomeRepository implements HomeRepository {
  @override
  Future<HomeRefreshStatus> loadRefreshStatus() async {
    return const HomeRefreshStatus(
      lastSuccessfulAt: null,
      activeRunStatus: ActiveRunStatus.none,
    );
  }
}

final class _FakeCategoryRepository implements CategoryRepository {
  _FakeCategoryRepository({List<Category> categories = const []})
    : _categories = [...categories];

  final List<Category> _categories;
  int createCalls = 0;
  CategoryDraft? lastDraft;

  @override
  Future<List<Category>> loadCategories() async =>
      List.unmodifiable(_categories);

  @override
  Future<Category> createCategory(CategoryDraft draft) async {
    createCalls += 1;
    lastDraft = draft;
    final category = Category(
      id: 'category-${_categories.length + 1}',
      name: draft.name,
      searchKeywords: draft.searchKeywords,
      specialRequirements: draft.specialRequirements,
      sourceSettings: [
        for (var index = 0; index < draft.sourceSettings.length; index += 1)
          SourceSetting(
            id: 'source-${index + 1}',
            label: draft.sourceSettings[index].label,
            websiteInput: draft.sourceSettings[index].websiteInput,
            kind: draft.sourceSettings[index].kind,
            position: index,
          ),
      ],
    );
    _categories.add(category);
    return category;
  }
}
