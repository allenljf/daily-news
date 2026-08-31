import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/news.dart';
import 'news_providers.dart';

final class SourceTagUi {
  const SourceTagUi({required this.id, required this.label});
  final String id;
  final String label;
}

final class NewsListUiState {
  NewsListUiState({
    required List<NewsItem> items,
    required List<SourceTagUi> sourceTags,
    this.nextCursor,
    this.selectedSourceTagId,
    this.isLoadingMore = false,
  }) : items = List.unmodifiable(items),
       sourceTags = List.unmodifiable(sourceTags);

  final List<NewsItem> items;
  final List<SourceTagUi> sourceTags;
  final String? nextCursor;
  final String? selectedSourceTagId;
  final bool isLoadingMore;

  NewsListUiState copyWith({
    List<NewsItem>? items,
    List<SourceTagUi>? sourceTags,
    String? nextCursor,
    String? selectedSourceTagId,
    bool clearNextCursor = false,
    bool clearSelection = false,
    bool? isLoadingMore,
  }) => NewsListUiState(
    items: items ?? this.items,
    sourceTags: sourceTags ?? this.sourceTags,
    nextCursor: clearNextCursor ? null : nextCursor ?? this.nextCursor,
    selectedSourceTagId: clearSelection
        ? null
        : selectedSourceTagId ?? this.selectedSourceTagId,
    isLoadingMore: isLoadingMore ?? this.isLoadingMore,
  );
}

final newsListControllerProvider =
    AsyncNotifierProvider.family<NewsListController, NewsListUiState, String>(
      NewsListController.new,
    );

final class NewsListController extends AsyncNotifier<NewsListUiState> {
  NewsListController(this.categoryId);
  final String categoryId;

  @override
  Future<NewsListUiState> build() => _load();

  Future<NewsListUiState> _load({String? sourceTagId}) async {
    final page = await ref
        .watch(newsRepositoryProvider)
        .loadNewsPage(
          NewsPageRequest(categoryId: categoryId, sourceTagId: sourceTagId),
        );
    return NewsListUiState(
      items: page.items,
      nextCursor: page.nextCursor,
      selectedSourceTagId: sourceTagId,
      sourceTags: _tags(page.items),
    );
  }

  Future<void> selectSourceTag(String? sourceTagId) async {
    final previousTags = state.asData?.value.sourceTags ?? const [];
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final loaded = await _load(sourceTagId: sourceTagId);
      return loaded.copyWith(sourceTags: previousTags);
    });
  }

  Future<void> loadMore() async {
    final current = state.asData?.value;
    if (current == null ||
        current.isLoadingMore ||
        current.nextCursor == null) {
      return;
    }
    state = AsyncData(current.copyWith(isLoadingMore: true));
    try {
      final page = await ref
          .read(newsRepositoryProvider)
          .loadNewsPage(
            NewsPageRequest(
              categoryId: categoryId,
              cursor: current.nextCursor,
              sourceTagId: current.selectedSourceTagId,
            ),
          );
      state = AsyncData(
        NewsListUiState(
          items: [...current.items, ...page.items],
          sourceTags: current.sourceTags,
          nextCursor: page.nextCursor,
          selectedSourceTagId: current.selectedSourceTagId,
        ),
      );
    } catch (error, stackTrace) {
      state = AsyncError(error, stackTrace);
    }
  }

  static List<SourceTagUi> _tags(List<NewsItem> items) {
    final labels = <String, String>{};
    for (final item in items) {
      labels[item.sourceTagId] = item.sourceTagLabel;
    }
    return [
      for (final entry in labels.entries)
        SourceTagUi(id: entry.key, label: entry.value),
    ];
  }
}
