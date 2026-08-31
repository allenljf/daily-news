import 'package:daily_news_mobile/features/news/application/news_providers.dart';
import 'package:daily_news_mobile/features/news/data/news.dart';
import 'package:daily_news_mobile/features/news/data/news_repository.dart';
import 'package:daily_news_mobile/features/news/presentation/news_detail_screen.dart';
import 'package:daily_news_mobile/features/news/presentation/news_list_screen.dart';
import 'package:daily_news_mobile/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
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
    'detail permanence and delete actions update through controller',
    (tester) async {
      final repository = _FakeNewsRepository();
      await _pump(
        tester,
        const NewsDetailScreen(categoryId: 'category-1', newsId: 'article-1'),
        repository,
      );

      await tester.tap(find.text('設為永久'));
      await tester.pumpAndSettle();
      expect(repository.permanentIds, ['article-1']);
      expect(find.text('已永久保存'), findsOneWidget);

      await tester.tap(find.text('刪除新聞'));
      await tester.pumpAndSettle();
      expect(repository.deletedIds, ['article-1']);
      expect(find.text('新聞已刪除'), findsOneWidget);
    },
  );
}

Future<void> _pump(
  WidgetTester tester,
  Widget child,
  NewsRepository repository,
) async {
  await tester.pumpWidget(
    ProviderScope(
      overrides: [newsRepositoryProvider.overrideWithValue(repository)],
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

final class _FakeNewsRepository implements NewsRepository {
  final List<NewsPageRequest> pageRequests = [];
  final List<String> permanentIds = [];
  final List<String> deletedIds = [];

  @override
  Future<NewsPage> loadNewsPage(NewsPageRequest request) async {
    pageRequests.add(request);
    if (request.cursor == 'next-page') {
      return NewsPage(items: [_item(21)], nextCursor: null);
    }
    return NewsPage(
      items: [for (var index = 1; index <= 20; index += 1) _item(index)],
      nextCursor: 'next-page',
    );
  }

  @override
  Future<NewsDetail> loadNewsDetail(String categoryId, String newsId) async {
    return NewsDetail.fromItem(_item(1), firstSeenAt: DateTime(2026));
  }

  @override
  Future<void> setPermanent(String newsId, bool permanent) async {
    permanentIds.add(newsId);
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
