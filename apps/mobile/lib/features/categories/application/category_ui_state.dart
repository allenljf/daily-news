import 'package:flutter/foundation.dart';

@immutable
final class CategoryItemUi {
  const CategoryItemUi({required this.id, required this.name});

  final String id;
  final String name;
}

@immutable
final class CategoriesUiState {
  CategoriesUiState({
    required List<CategoryItemUi> categories,
    this.isSubmitting = false,
    this.saveFailed = false,
  }) : categories = List.unmodifiable(categories);

  final List<CategoryItemUi> categories;
  final bool isSubmitting;
  final bool saveFailed;

  CategoriesUiState copyWith({
    List<CategoryItemUi>? categories,
    bool? isSubmitting,
    bool? saveFailed,
  }) {
    return CategoriesUiState(
      categories: categories ?? this.categories,
      isSubmitting: isSubmitting ?? this.isSubmitting,
      saveFailed: saveFailed ?? this.saveFailed,
    );
  }
}
