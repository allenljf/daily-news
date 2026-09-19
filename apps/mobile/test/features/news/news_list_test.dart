import 'package:daily_news_mobile/features/news/application/news_providers.dart';
import 'package:daily_news_mobile/features/categories/application/category_providers.dart';
import 'package:daily_news_mobile/features/categories/data/category.dart';
import 'package:daily_news_mobile/features/categories/data/category_repository.dart';
import 'package:daily_news_mobile/features/categories/presentation/category_settings_sheet.dart';
import 'package:daily_news_mobile/features/news/data/news.dart';
import 'package:daily_news_mobile/features/news/data/news_repository.dart';
import 'package:daily_news_mobile/features/news/presentation/article_web_view.dart';
import 'package:daily_news_mobile/features/news/presentation/news_detail_screen.dart';
import 'package:daily_news_mobile/features/news/presentation/news_list_screen.dart';
import 'package:daily_news_mobile/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

void main() {
  testWidgets('list title uses the selected Category name', (tester) async {
    await _pump(
      tester,
      const NewsListScreen(categoryId: 'category-1'),
      _FakeNewsRepository(),
      categoryRepository: _FakeCategoryRepository(
        categories: [Category(id: 'category-1', name: 'AI Coding Tools')],
      ),
    );

    expect(find.widgetWithText(AppBar, 'AI Coding Tools'), findsOneWidget);
    expect(find.widgetWithText(AppBar, '新聞'), findsNothing);
  });

  testWidgets('empty news list says no data is currently available', (
    tester,
  ) async {
    await _pump(
      tester,
      const NewsListScreen(categoryId: 'category-1'),
      _FakeNewsRepository(firstPageItems: const []),
    );

    expect(find.text('目前沒有資料'), findsOneWidget);
  });

  testWidgets('scroll loads the next 20-item page with its cursor', (
    tester,
  ) async {
    final repository = _FakeNewsRepository();
    await _pump(
      tester,
      const NewsListScreen(categoryId: 'category-1'),
      repository,
    );

    expect(find.text('Article 1'), findsOneWidget);
    expect(repository.pageRequests.single.cursor, isNull);
    expect(repository.pageRequests.single.limit, 20);

    await tester.drag(find.byType(ListView), const Offset(0, -5000));
    await tester.pumpAndSettle();

    expect(repository.pageRequests.map((request) => request.cursor), [
      null,
      'next-page',
    ]);
    expect(find.text('Article 21'), findsOneWidget);
  });

  testWidgets('source tag filter resets pagination', (tester) async {
    final repository = _FakeNewsRepository();
    await _pump(
      tester,
      const NewsListScreen(categoryId: 'category-1'),
      repository,
    );

    await tester.tap(find.widgetWithText(FilterChip, 'GitHub'));
    await tester.pumpAndSettle();

    expect(repository.pageRequests.last.cursor, isNull);
    expect(repository.pageRequests.last.sourceTagId, 'source-github');
  });

  testWidgets(
    'detail renders the original Article in a WebView with its controls',
    (tester) async {
      final repository = _FakeNewsRepository();
      Uri? loadedUrl;
      await _pump(
        tester,
        const NewsDetailScreen(categoryId: 'category-1', newsId: 'article-1'),
        repository,
        webViewBuilder: (context, url) {
          loadedUrl = url;
          return const SizedBox(key: articleWebViewKey);
        },
      );

      expect(loadedUrl, Uri.parse('https://example.test/1'));
      expect(find.byKey(articleWebViewKey), findsOneWidget);
      expect(find.byTooltip('設為永久'), findsOneWidget);
      expect(find.byKey(deleteNewsButtonKey), findsOneWidget);
      expect(find.widgetWithText(AppBar, 'Article 1'), findsOneWidget);
      expect(find.text('新聞詳情'), findsNothing);
      expect(find.text('Summary 1'), findsNothing);
      expect(find.widgetWithText(Chip, 'GitHub'), findsNothing);
    },
  );

  testWidgets('detail toolbar keeps the Article title on one ellipsized line', (
    tester,
  ) async {
    final repository = _FakeNewsRepository(
      detailItem: NewsItem(
        id: 'article-1',
        title:
            'A very long original Article title that should never wrap onto a second line in the toolbar',
        canonicalUrl: Uri.parse('https://example.test/1'),
        insertedAt: DateTime(2026, 8, 30, 8, 1),
        sourceTagId: 'source-github',
        sourceTagLabel: 'GitHub',
        isPermanent: false,
      ),
    );
    await _pump(
      tester,
      const NewsDetailScreen(categoryId: 'category-1', newsId: 'article-1'),
      repository,
      webViewBuilder: (context, url) => const SizedBox.shrink(),
    );

    final title = tester.widget<Text>(
      find.descendant(of: find.byType(AppBar), matching: find.byType(Text)),
    );
    expect(title.maxLines, 1);
    expect(title.overflow, TextOverflow.ellipsis);
  });

  testWidgets('tapping permanence again restores the original expiry', (
    tester,
  ) async {
    final repository = _FakeNewsRepository();
    await _pump(
      tester,
      const NewsDetailScreen(categoryId: 'category-1', newsId: 'article-1'),
      repository,
      webViewBuilder: (context, url) => const SizedBox.shrink(),
    );

    await tester.tap(find.byKey(makePermanentButtonKey));
    await tester.pumpAndSettle();
    expect(repository.permanentFlags, [true]);
    expect(find.byTooltip('恢復原本時效'), findsOneWidget);

    await tester.tap(find.byKey(makePermanentButtonKey));
    await tester.pumpAndSettle();
    expect(repository.permanentFlags, [true, false]);
    expect(find.byTooltip('設為永久'), findsOneWidget);
  });

  testWidgets('detail refuses to load a non-HTTPS Article URL in a WebView', (
    tester,
  ) async {
    final repository = _FakeNewsRepository(
      detailItem: NewsItem(
        id: 'article-1',
        title: 'Article 1',
        canonicalUrl: Uri.parse('http://example.test/1'),
        insertedAt: DateTime(2026, 8, 30, 8, 1),
        sourceTagId: 'source-github',
        sourceTagLabel: 'GitHub',
        isPermanent: false,
      ),
    );
    await _pump(
      tester,
      const NewsDetailScreen(categoryId: 'category-1', newsId: 'article-1'),
      repository,
      webViewBuilder: (context, url) => const SizedBox(key: articleWebViewKey),
    );

    expect(find.byKey(articleWebViewKey), findsNothing);
    expect(find.byKey(articleWebViewBlockedKey), findsOneWidget);
  });

  testWidgets('deleting an Article returns to its Category list', (
    tester,
  ) async {
    final repository = _FakeNewsRepository();
    await _pumpNewsRoute(tester, repository);

    expect(find.text('Article 1'), findsOneWidget);

    await tester.tap(find.byKey(newsListItemKey('article-1')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(deleteNewsButtonKey));
    await tester.pumpAndSettle();

    expect(repository.deletedIds, ['article-1']);
    expect(find.byType(NewsDetailScreen), findsNothing);
    expect(find.text('Article 1'), findsNothing);
    expect(find.text('Article 2'), findsOneWidget);
  });

  testWidgets(
    'Category delete confirmation stops its schedule and returns home',
    (tester) async {
      final categoryRepository = _FakeCategoryRepository(
        categories: [Category(id: 'category-1', name: 'AI Coding Tools')],
      );
      await _pumpCategoryRoute(tester, categoryRepository);

      await tester.tap(find.byIcon(Icons.delete_outline));
      await tester.pumpAndSettle();

      expect(find.text('刪除新聞類別'), findsOneWidget);
      expect(find.text('會停止後續擷取並移除這個類別的關聯。其他類別仍引用的文章不會被刪除。'), findsOneWidget);

      await tester.tap(find.text('取消'));
      await tester.pumpAndSettle();

      expect(categoryRepository.deletedIds, isEmpty);
      expect(find.widgetWithText(AppBar, 'AI Coding Tools'), findsOneWidget);

      await tester.tap(find.byIcon(Icons.delete_outline));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, '刪除'));
      await tester.pumpAndSettle();

      expect(categoryRepository.deletedIds, ['category-1']);
      expect(categoryRepository.loadCount, 2);
      expect(find.text('首頁'), findsOneWidget);
    },
  );

  testWidgets('edit action prefills and saves Category settings', (
    tester,
  ) async {
    final categoryRepository = _FakeCategoryRepository(
      categories: [
        Category(
          id: 'category-1',
          name: 'AI Coding Tools',
          searchKeywords: 'agents',
          sourceSettings: const [
            SourceSetting(
              id: 'source-1',
              label: 'blog',
              websiteInput: 'https://blog.example/rss',
              kind: 'website',
              position: 0,
            ),
          ],
        ),
      ],
    );
    await _pumpCategoryRoute(tester, categoryRepository);

    await tester.tap(find.byKey(editCategoryButtonKey));
    await tester.pumpAndSettle();

    expect(
      tester
          .widget<TextFormField>(find.byKey(categoryNameFieldKey))
          .controller
          ?.text,
      'AI Coding Tools',
    );
    expect(
      tester
          .widget<TextFormField>(find.byKey(sourceSettingFieldKey(0)))
          .controller
          ?.text,
      'https://blog.example/rss',
    );

    await tester.tap(find.byKey(addSourceSettingButtonKey));
    await tester.pumpAndSettle();
    await tester.enterText(
      find.byKey(addSourceDialogFieldKey),
      'https://www.youtube.com',
    );
    await tester.tap(find.widgetWithText(FilledButton, '新增'));
    await tester.pumpAndSettle();

    await tester.ensureVisible(find.byKey(saveCategoryButtonKey));
    await tester.tap(find.byKey(saveCategoryButtonKey));
    await tester.pumpAndSettle();

    expect(categoryRepository.updateCalls, 1);
    expect(categoryRepository.lastUpdatedId, 'category-1');
    expect(categoryRepository.lastDraft?.name, 'AI Coding Tools');
    expect(
      categoryRepository.lastDraft?.sourceSettings.map(
        (source) => source.websiteInput,
      ),
      ['https://blog.example/rss', 'https://www.youtube.com'],
    );
    expect(find.byType(CategorySettingsSheet), findsNothing);
  });
}

Future<void> _pump(
  WidgetTester tester,
  Widget child,
  NewsRepository repository, {
  CategoryRepository? categoryRepository,
  ArticleWebViewBuilder? webViewBuilder,
}) async {
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        newsRepositoryProvider.overrideWithValue(repository),
        if (categoryRepository != null)
          categoryRepositoryProvider.overrideWithValue(categoryRepository),
        if (webViewBuilder != null)
          articleWebViewBuilderProvider.overrideWithValue(webViewBuilder),
      ],
      child: MaterialApp(
        locale: const Locale('zh'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: child,
      ),
    ),
  );
  await tester.pumpAndSettle();
}

Future<void> _pumpNewsRoute(
  WidgetTester tester,
  NewsRepository repository,
) async {
  final router = GoRouter(
    initialLocation: '/categories/category-1/news',
    routes: [
      GoRoute(
        path: '/categories/:categoryId/news',
        builder: (_, state) =>
            NewsListScreen(categoryId: state.pathParameters['categoryId']!),
        routes: [
          GoRoute(
            path: ':newsId',
            builder: (_, state) => NewsDetailScreen(
              categoryId: state.pathParameters['categoryId']!,
              newsId: state.pathParameters['newsId']!,
            ),
          ),
        ],
      ),
    ],
  );
  addTearDown(router.dispose);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        newsRepositoryProvider.overrideWithValue(repository),
        articleWebViewBuilderProvider.overrideWithValue(
          (context, url) => const SizedBox.shrink(),
        ),
      ],
      child: MaterialApp.router(
        locale: const Locale('zh'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        routerConfig: router,
      ),
    ),
  );
  await tester.pumpAndSettle();
}

Future<void> _pumpCategoryRoute(
  WidgetTester tester,
  CategoryRepository categoryRepository,
) async {
  final router = GoRouter(
    initialLocation: '/categories/category-1',
    routes: [
      GoRoute(
        path: '/',
        builder: (_, _) => const Scaffold(body: Text('首頁')),
      ),
      GoRoute(
        path: '/categories/:categoryId',
        builder: (_, state) =>
            NewsListScreen(categoryId: state.pathParameters['categoryId']!),
      ),
    ],
  );
  addTearDown(router.dispose);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        newsRepositoryProvider.overrideWithValue(_FakeNewsRepository()),
        categoryRepositoryProvider.overrideWithValue(categoryRepository),
      ],
      child: MaterialApp.router(
        locale: const Locale('zh'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        routerConfig: router,
      ),
    ),
  );
  await tester.pumpAndSettle();
}

final class _FakeNewsRepository implements NewsRepository {
  _FakeNewsRepository({this.firstPageItems, this.detailItem});

  final List<NewsItem>? firstPageItems;
  final NewsItem? detailItem;
  final List<NewsPageRequest> pageRequests = [];
  final List<String> permanentIds = [];
  final List<bool> permanentFlags = [];
  final List<String> deletedIds = [];

  @override
  Future<NewsPage> loadNewsPage(NewsPageRequest request) async {
    pageRequests.add(request);
    if (request.cursor == 'next-page') {
      return NewsPage(items: [_item(21)], nextCursor: null);
    }
    final items =
        firstPageItems ??
        [for (var index = 1; index <= 20; index += 1) _item(index)];
    return NewsPage(
      items: [
        for (final item in items)
          if (!deletedIds.contains(item.id)) item,
      ],
      nextCursor: 'next-page',
    );
  }

  @override
  Future<NewsDetail> loadNewsDetail(String categoryId, String newsId) async {
    return NewsDetail.fromItem(
      detailItem ?? _item(1),
      firstSeenAt: DateTime(2026),
    );
  }

  @override
  Future<DateTime?> setPermanent(String newsId, bool permanent) async {
    permanentIds.add(newsId);
    permanentFlags.add(permanent);
    return permanent ? null : DateTime(2026, 10, 1);
  }

  @override
  Future<void> deleteNews(String newsId) async {
    deletedIds.add(newsId);
  }

  NewsItem _item(int index) => NewsItem(
    id: 'article-$index',
    title: 'Article $index',
    summary: 'Summary $index',
    canonicalUrl: Uri.parse('https://example.test/$index'),
    insertedAt: DateTime(2026, 8, 30, 8, index),
    sourceTagId: index.isEven ? 'source-web' : 'source-github',
    sourceTagLabel: index.isEven ? 'Web' : 'GitHub',
    isPermanent: false,
  );
}

final class _FakeCategoryRepository implements CategoryRepository {
  _FakeCategoryRepository({required this.categories});

  final List<Category> categories;
  final List<String> deletedIds = [];
  var loadCount = 0;
  var updateCalls = 0;
  String? lastUpdatedId;
  CategoryDraft? lastDraft;

  @override
  Future<List<Category>> loadCategories() async {
    loadCount++;
    return categories;
  }

  @override
  Future<Category> createCategory(CategoryDraft draft) =>
      throw UnimplementedError();

  @override
  Future<Category> updateCategory(
    String categoryId,
    CategoryDraft draft,
  ) async {
    updateCalls++;
    lastUpdatedId = categoryId;
    lastDraft = draft;
    final updated = Category(
      id: categoryId,
      name: draft.name,
      searchKeywords: draft.searchKeywords,
      specialRequirements: draft.specialRequirements,
      contentLanguage: draft.contentLanguage,
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
    final index = categories.indexWhere(
      (category) => category.id == categoryId,
    );
    if (index >= 0) {
      categories[index] = updated;
    }
    return updated;
  }

  @override
  Future<void> deleteCategory(String categoryId) async {
    deletedIds.add(categoryId);
    categories.removeWhere((category) => category.id == categoryId);
  }
}
