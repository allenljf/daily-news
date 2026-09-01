import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../l10n/app_localizations.dart';
import '../application/news_list_controller.dart';

Key newsListItemKey(String newsId) => ValueKey('news-item-$newsId');

final class NewsListScreen extends ConsumerStatefulWidget {
  const NewsListScreen({required this.categoryId, super.key});
  final String categoryId;

  @override
  ConsumerState<NewsListScreen> createState() => _NewsListScreenState();
}

final class _NewsListScreenState extends ConsumerState<NewsListScreen> {
  late final ScrollController _scrollController;

  @override
  void initState() {
    super.initState();
    _scrollController = ScrollController()..addListener(_loadMore);
  }

  void _loadMore() {
    if (_scrollController.position.extentAfter < 200) {
      ref
          .read(newsListControllerProvider(widget.categoryId).notifier)
          .loadMore();
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context);
    final provider = newsListControllerProvider(widget.categoryId);
    ref.listen(provider, (previous, next) {
      final previousLength = previous?.asData?.value.items.length ?? 0;
      final nextLength = next.asData?.value.items.length ?? 0;
      if (nextLength > previousLength && previousLength > 0) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (_scrollController.hasClients) {
            _scrollController.jumpTo(
              _scrollController.position.maxScrollExtent,
            );
          }
        });
      }
    });
    final value = ref.watch(provider);
    return Scaffold(
      appBar: AppBar(title: Text(localizations.news)),
      body: value.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, _) => Center(child: Text(localizations.newsLoadFailed)),
        data: (state) => Column(
          children: [
            if (state.sourceTags.isNotEmpty)
              SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.all(AppSpacing.medium),
                child: Row(
                  children: [
                    for (final tag in state.sourceTags)
                      Padding(
                        padding: const EdgeInsets.only(right: AppSpacing.small),
                        child: FilterChip(
                          label: Text(tag.label),
                          selected: state.selectedSourceTagId == tag.id,
                          onSelected: (selected) => ref
                              .read(
                                newsListControllerProvider(widget.categoryId)
                                    .notifier,
                              )
                              .selectSourceTag(selected ? tag.id : null),
                        ),
                      ),
                  ],
                ),
              ),
            Expanded(
              child: state.items.isEmpty
                  ? Center(child: Text(localizations.noNews))
                  : ListView.builder(
                      controller: _scrollController,
                      itemCount:
                          state.items.length + (state.isLoadingMore ? 1 : 0),
                      itemBuilder: (context, index) {
                        if (index == state.items.length) {
                          return const Center(
                            child: CircularProgressIndicator(),
                          );
                        }
                        final item = state.items[index];
                        final time = DateFormat.yMd(
                          Localizations.localeOf(context).toLanguageTag(),
                        ).add_Hm().format(item.insertedAt.toLocal());
                        return ListTile(
                          key: newsListItemKey(item.id),
                          title: Text(item.title),
                          subtitle: Text('$time · ${item.sourceTagLabel}'),
                          trailing: item.isPermanent
                              ? const Icon(Icons.bookmark)
                              : null,
                          onTap: () => context.push(
                            '/categories/${widget.categoryId}/news/${item.id}',
                          ),
                        );
                      },
                    ),
            ),
          ],
        ),
      ),
    );
  }
}
