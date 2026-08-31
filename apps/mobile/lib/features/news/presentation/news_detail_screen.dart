import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../l10n/app_localizations.dart';
import '../application/news_detail_controller.dart';

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
    final value = ref.watch(newsDetailControllerProvider(args));
    return Scaffold(
      appBar: AppBar(title: Text(localizations.newsDetail)),
      body: value.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, _) => Center(child: Text(localizations.newsLoadFailed)),
        data: (state) {
          if (state.deleted) {
            return Center(child: Text(localizations.newsDeleted));
          }
          final item = state.detail.item;
          return ListView(
            padding: const EdgeInsets.all(AppSpacing.large),
            children: [
              Text(
                item.title,
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: AppSpacing.medium),
              Chip(label: Text(item.sourceTagLabel)),
              if (item.summary != null) ...[
                const SizedBox(height: AppSpacing.medium),
                Text(item.summary!),
              ],
              const SizedBox(height: AppSpacing.medium),
              SelectableText(item.canonicalUrl.toString()),
              const SizedBox(height: AppSpacing.large),
              FilledButton(
                onPressed: item.isPermanent
                    ? null
                    : () => unawaited(
                        ref
                            .read(newsDetailControllerProvider(args).notifier)
                            .setPermanent(),
                      ),
                child: Text(
                  item.isPermanent
                      ? localizations.savedPermanently
                      : localizations.makePermanent,
                ),
              ),
              const SizedBox(height: AppSpacing.small),
              OutlinedButton(
                onPressed: () => unawaited(
                  ref
                      .read(newsDetailControllerProvider(args).notifier)
                      .delete(),
                ),
                child: Text(localizations.deleteNews),
              ),
            ],
          );
        },
      ),
    );
  }
}
