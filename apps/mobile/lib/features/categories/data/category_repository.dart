import '../../../core/api/api_failure.dart';
import 'category.dart';
import 'category_dto.dart';
import 'category_remote_service.dart';

abstract interface class CategoryRepository {
  Future<List<Category>> loadCategories();

  Future<Category> createCategory(CategoryDraft draft);

  Future<Category> updateCategory(String categoryId, CategoryDraft draft);

  Future<void> deleteCategory(String categoryId);
}

enum CategoryRepositoryFailure implements Exception { unavailable }

final class RemoteCategoryRepository implements CategoryRepository {
  RemoteCategoryRepository(this._remoteService);

  final CategoryRemoteService _remoteService;

  @override
  Future<List<Category>> loadCategories() async {
    try {
      final categories = await _remoteService.loadCategories();
      return categories.map(_toCategory).toList(growable: false);
    } on ApiFailure {
      throw CategoryRepositoryFailure.unavailable;
    }
  }

  @override
  Future<Category> createCategory(CategoryDraft draft) async {
    try {
      return _toCategory(await _remoteService.createCategory(draft));
    } on ApiFailure {
      throw CategoryRepositoryFailure.unavailable;
    }
  }

  @override
  Future<Category> updateCategory(
    String categoryId,
    CategoryDraft draft,
  ) async {
    try {
      return _toCategory(
        await _remoteService.updateCategory(categoryId, draft),
      );
    } on ApiFailure {
      throw CategoryRepositoryFailure.unavailable;
    }
  }

  @override
  Future<void> deleteCategory(String categoryId) async {
    try {
      await _remoteService.deleteCategory(categoryId);
    } on ApiFailure {
      throw CategoryRepositoryFailure.unavailable;
    }
  }

  static Category _toCategory(CategoryDto dto) {
    return Category(
      id: dto.id,
      name: dto.name,
      searchKeywords: dto.searchKeywords,
      specialRequirements: dto.specialRequirements,
      contentLanguage: dto.contentLanguage,
      sourceSettings: [
        for (final source in dto.sourceSettings)
          SourceSetting(
            id: source.id,
            label: source.label,
            websiteInput: source.websiteInput,
            normalizedHost: source.normalizedHost,
            kind: source.kind,
            position: source.position,
          ),
      ],
    );
  }
}
