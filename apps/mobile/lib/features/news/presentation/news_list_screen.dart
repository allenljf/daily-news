import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../l10n/app_localizations.dart';
import '../../categories/application/category_controller.dart';
import '../../categories/application/category_ui_state.dart';
import '../../categories/presentation/category_settings_sheet.dart';
import '../application/news_list_controller.dart';

Key newsListItemKey(String newsId) => ValueKey('news-item-$newsId');

const editCategoryButtonKey = Key('edit-category-button');

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
    final category = _findCategory(
      ref.watch(categoriesControllerProvider).asData?.value,
      widget.categoryId,
    );
    final categoryTitle = category?.name ?? localizations.news;
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
      appBar: AppBar(
        title: Text(categoryTitle),
        actions: [
          IconButton(
            key: editCategoryButtonKey,
            tooltip: localizations.editCategory,
            onPressed: category == null
                ? null
                : () => _openCategorySettings(context, category),
            icon: const Icon(Icons.edit_outlined),
          ),
          IconButton(
            tooltip: localizations.deleteCategory,
            onPressed: () => _confirmCategoryDeletion(context),
            icon: const Icon(Icons.delete_outline),
          ),
        ],
      ),
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

  Future<void> _openCategorySettings(
    BuildContext context,
    CategoryItemUi category,
  ) async {
    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (context) => CategorySettingsSheet(category: category),
    );
  }

  Future<void> _confirmCategoryDeletion(BuildContext context) async {
    final localizations = AppLocalizations.of(context);
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(localizations.deleteCategory),
        content: Text(localizations.deleteCategoryExplanation),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(false),
            child: Text(localizations.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.of(dialogContext).pop(true),
            child: Text(localizations.delete),
          ),
        ],
      ),
    );
    if (!context.mounted || confirmed != true) {
      return;
    }
    final deleted = await ref
        .read(categoriesControllerProvider.notifier)
        .deleteCategory(widget.categoryId);
    if (!context.mounted) {
      return;
    }
    if (deleted) {
      context.go('/');
      return;
    }
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(localizations.categoryDeleteFailed)));
  }
}

CategoryItemUi? _findCategory(
  CategoriesUiState? categoryState,
  String categoryId,
) {
  for (final category
      in categoryState?.categories ?? const <CategoryItemUi>[]) {
    if (category.id == categoryId) {
      return category;
    }
  }
  return null;
}
