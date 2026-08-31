import '../../../core/api/api_client.dart';
import 'category.dart';
import 'category_dto.dart';

final class CategoryRemoteService {
  CategoryRemoteService(this._apiClient);

  final ApiClient _apiClient;

  Future<List<CategoryDto>> loadCategories() {
    return _apiClient.get(
      'categories',
      decode: (json) {
        if (json is! List) {
          throw const FormatException('Expected a JSON list.');
        }
        return [
          for (final item in json)
            if (item is Map)
              CategoryDto.fromJson(Map<String, Object?>.from(item))
            else
              throw const FormatException('Category entries must be objects.'),
        ];
      },
    );
  }

  Future<CategoryDto> createCategory(CategoryDraft draft) {
    return _apiClient.post(
      'categories',
      data: {
        'name': draft.name,
        'search_keywords': draft.searchKeywords,
        'special_requirements': draft.specialRequirements,
        'source_settings': [
          for (final source in draft.sourceSettings)
            {
              'label': source.label,
              'website_input': source.websiteInput,
              'kind': source.kind,
            },
        ],
      },
      decode: (json) {
        if (json case final Map<String, Object?> object) {
          return CategoryDto.fromJson(object);
        }
        throw const FormatException('Expected a JSON object.');
      },
    );
  }
}
