import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/news.dart';
import 'news_providers.dart';

typedef NewsDetailArgs = ({String categoryId, String newsId});

final class NewsDetailUiState {
  const NewsDetailUiState({required this.detail, this.deleted = false});
  final NewsDetail detail;
  final bool deleted;
  NewsDetailUiState copyWith({NewsDetail? detail, bool? deleted}) =>
      NewsDetailUiState(
        detail: detail ?? this.detail,
        deleted: deleted ?? this.deleted,
      );
}

final newsDetailControllerProvider =
    AsyncNotifierProvider.family<
      NewsDetailController,
      NewsDetailUiState,
      NewsDetailArgs
    >(NewsDetailController.new);

final class NewsDetailController extends AsyncNotifier<NewsDetailUiState> {
  NewsDetailController(this.args);
  final NewsDetailArgs args;

  @override
  Future<NewsDetailUiState> build() async => NewsDetailUiState(
    detail: await ref
        .watch(newsRepositoryProvider)
        .loadNewsDetail(args.categoryId, args.newsId),
  );

  Future<void> setPermanent() async {
    final current = state.asData?.value;
    if (current == null) return;
    await ref.read(newsRepositoryProvider).setPermanent(args.newsId, true);
    state = AsyncData(
      current.copyWith(
        detail: current.detail.copyWith(
          item: current.detail.item.copyWith(isPermanent: true),
        ),
      ),
    );
  }

  Future<void> delete() async {
    final current = state.asData?.value;
    if (current == null) return;
    await ref.read(newsRepositoryProvider).deleteNews(args.newsId);
    state = AsyncData(current.copyWith(deleted: true));
  }
}
