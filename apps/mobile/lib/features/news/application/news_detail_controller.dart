import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/news.dart';
import 'news_list_controller.dart';
import 'news_providers.dart';

typedef NewsDetailArgs = ({String categoryId, String newsId});

final newsDetailControllerProvider =
    AsyncNotifierProvider.family<
      NewsDetailController,
      NewsDetail,
      NewsDetailArgs
    >(NewsDetailController.new);

final class NewsDetailController extends AsyncNotifier<NewsDetail> {
  NewsDetailController(this.args);
  final NewsDetailArgs args;

  @override
  Future<NewsDetail> build() => ref
      .watch(newsRepositoryProvider)
      .loadNewsDetail(args.categoryId, args.newsId);

  Future<void> togglePermanent() async {
    final current = state.asData?.value;
    if (current == null) return;
    final item = current.item;
    final expiresAt = await ref
        .read(newsRepositoryProvider)
        .setPermanent(args.newsId, !item.isPermanent);
    state = AsyncData(
      current.copyWith(
        item: item.copyWith(
          isPermanent: expiresAt == null,
          expiresAt: expiresAt,
        ),
      ),
    );
  }

  Future<void> delete() async {
    await ref.read(newsRepositoryProvider).deleteNews(args.newsId);
    ref.invalidate(newsListControllerProvider(args.categoryId));
  }
}
