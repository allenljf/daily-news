import 'package:flutter/material.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../l10n/app_localizations.dart';
import '../application/category_ui_state.dart';

const addCategoryTileKey = Key('add-category-tile');

final class CategoryGrid extends StatelessWidget {
  const CategoryGrid({
    required this.categories,
    required this.onCategoryPressed,
    required this.onAddCategoryPressed,
    super.key,
  });

  final List<CategoryItemUi> categories;
  final ValueChanged<String> onCategoryPressed;
  final VoidCallback onAddCategoryPressed;

  @override
  Widget build(BuildContext context) {
    return GridView.builder(
      padding: const EdgeInsets.all(AppSpacing.large),
      gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
        maxCrossAxisExtent: AppSizing.categoryTileMaxWidth,
        mainAxisSpacing: AppSpacing.medium,
        crossAxisSpacing: AppSpacing.medium,
        childAspectRatio: AppSizing.categoryTileAspectRatio,
      ),
      itemCount: categories.length + 1,
      itemBuilder: (context, index) {
        if (index == categories.length) {
          return _AddCategoryTile(onPressed: onAddCategoryPressed);
        }
        final category = categories[index];
        return CategoryTile(
          key: ValueKey('category-${category.id}'),
          name: category.name,
          onPressed: () => onCategoryPressed(category.id),
        );
      },
    );
  }
}

final class CategoryTile extends StatelessWidget {
  const CategoryTile({required this.name, required this.onPressed, super.key});

  final String name;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onPressed,
        child: Center(
          child: Padding(
            padding: const EdgeInsets.all(AppSpacing.medium),
            child: Text(
              name,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.titleMedium,
            ),
          ),
        ),
      ),
    );
  }
}

final class _AddCategoryTile extends StatelessWidget {
  const _AddCategoryTile({required this.onPressed});

  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context);
    return Card(
      key: addCategoryTileKey,
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onPressed,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.add),
            const SizedBox(height: AppSpacing.small),
            Text(localizations.addCategory),
          ],
        ),
      ),
    );
  }
}
