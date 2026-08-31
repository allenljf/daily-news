import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../l10n/app_localizations.dart';
import '../application/category_controller.dart';
import 'category_grid.dart';
import 'category_settings_sheet.dart';

final class CategoryHomeSection extends ConsumerWidget {
  const CategoryHomeSection({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final localizations = AppLocalizations.of(context);
    return ref
        .watch(categoriesControllerProvider)
        .when(
          loading: () => Center(
            child: Semantics(
              label: localizations.loading,
              child: const CircularProgressIndicator(),
            ),
          ),
          error: (_, _) => Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(localizations.categoriesLoadFailed),
                const SizedBox(height: AppSpacing.medium),
                FilledButton(
                  onPressed: () =>
                      ref.read(categoriesControllerProvider.notifier).reload(),
                  child: Text(localizations.retry),
                ),
              ],
            ),
          ),
          data: (state) => CategoryGrid(
            categories: state.categories,
            onCategoryPressed: _newsRouteDeferredToF5,
            onAddCategoryPressed: () => showModalBottomSheet<void>(
              context: context,
              isScrollControlled: true,
              builder: (context) => const CategorySettingsSheet(),
            ),
          ),
        );
  }
}

void _newsRouteDeferredToF5(String categoryId) {}
