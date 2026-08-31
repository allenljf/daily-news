import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/category.dart';
import '../data/category_repository.dart';
import 'category_providers.dart';
import 'category_ui_state.dart';

final categoriesControllerProvider =
    AsyncNotifierProvider<CategoriesController, CategoriesUiState>(
      CategoriesController.new,
    );

final class CategoriesController extends AsyncNotifier<CategoriesUiState> {
  @override
  Future<CategoriesUiState> build() {
    return _load(ref.watch(categoryRepositoryProvider));
  }

  Future<void> reload() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(
      () => _load(ref.read(categoryRepositoryProvider)),
    );
  }

  Future<bool> createCategory(CategoryDraft draft) async {
    final current = state.asData?.value;
    if (current == null || current.isSubmitting) {
      return false;
    }
    state = AsyncData(current.copyWith(isSubmitting: true, saveFailed: false));
    final repository = ref.read(categoryRepositoryProvider);
    try {
      await repository.createCategory(draft);
      state = AsyncData(await _load(repository));
      return true;
    } on CategoryRepositoryFailure {
      state = AsyncData(
        current.copyWith(isSubmitting: false, saveFailed: true),
      );
      return false;
    }
  }

  Future<CategoriesUiState> _load(CategoryRepository repository) async {
    final categories = await repository.loadCategories();
    return CategoriesUiState(
      categories: [
        for (final category in categories)
          CategoryItemUi(id: category.id, name: category.name),
      ],
    );
  }
}
