import 'package:flutter/foundation.dart';

import '../data/category.dart' show traditionalChineseContentLanguage;

@immutable
final class CategoryItemUi {
  CategoryItemUi({
    required this.id,
    required this.name,
    this.searchKeywords,
    this.specialRequirements,
    this.contentLanguage = traditionalChineseContentLanguage,
    List<SourceSettingItemUi> sourceSettings = const [],
  }) : sourceSettings = List.unmodifiable(sourceSettings);

  final String id;
  final String name;
  final String? searchKeywords;
  final String? specialRequirements;
  final String contentLanguage;
  final List<SourceSettingItemUi> sourceSettings;
}

@immutable
final class SourceSettingItemUi {
  const SourceSettingItemUi({
    required this.label,
    required this.websiteInput,
    required this.kind,
  });

  final String label;
  final String websiteInput;
  final String kind;
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
