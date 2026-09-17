import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../l10n/app_localizations.dart';
import '../application/news_detail_controller.dart';
import '../data/news_repository.dart';
import 'article_web_view.dart';

const makePermanentButtonKey = Key('make-permanent-button');
const deleteNewsButtonKey = Key('delete-news-button');

final class NewsDetailScreen extends ConsumerWidget {
  const NewsDetailScreen({
    required this.categoryId,
    required this.newsId,
    super.key,
  });

  final String categoryId;
  final String newsId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final localizations = AppLocalizations.of(context);
    final args = (categoryId: categoryId, newsId: newsId);
    final provider = newsDetailControllerProvider(args);
    final value = ref.watch(provider);
    final item = value.asData?.value.item;
    return Scaffold(
      appBar: AppBar(
        title: Text(localizations.newsDetail),
        actions: [
          if (item != null) ...[
            IconButton(
              key: makePermanentButtonKey,
              tooltip: item.isPermanent
                  ? localizations.restoreExpiry
                  : localizations.makePermanent,
              onPressed: () =>
                  unawaited(ref.read(provider.notifier).togglePermanent()),
              icon: Icon(
                item.isPermanent ? Icons.bookmark : Icons.bookmark_border,
              ),
            ),
            IconButton(
              key: deleteNewsButtonKey,
              tooltip: localizations.deleteNews,
              onPressed: () => unawaited(_delete(context, ref, localizations)),
              icon: const Icon(Icons.delete_outline),
            ),
          ],
        ],
      ),
      body: value.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, _) => Center(child: Text(localizations.newsLoadFailed)),
        data: (detail) {
          final item = detail.item;
          final summary = item.summary;
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(
                  AppSpacing.large,
                  AppSpacing.large,
                  AppSpacing.large,
                  AppSpacing.medium,
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      item.title,
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                    const SizedBox(height: AppSpacing.small),
                    Chip(label: Text(item.sourceTagLabel)),
                    if (summary != null && summary.isNotEmpty) ...[
                      const SizedBox(height: AppSpacing.small),
                      Text(
                        summary,
                        maxLines: 3,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ],
                ),
              ),
              const Divider(height: 1),
              Expanded(child: ArticleWebView(url: item.canonicalUrl)),
            ],
          );
        },
      ),
    );
  }

  Future<void> _delete(
    BuildContext context,
    WidgetRef ref,
    AppLocalizations localizations,
  ) async {
    final provider = newsDetailControllerProvider((
      categoryId: categoryId,
      newsId: newsId,
    ));
    try {
      await ref.read(provider.notifier).delete();
    } on NewsRepositoryFailure {
      if (context.mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(localizations.newsDeleteFailed)));
      }
      return;
    }
    if (context.mounted && context.canPop()) {
      context.pop();
    }
  }
}
