import 'category.dart';

final class CategoryDto {
  const CategoryDto({
    required this.id,
    required this.name,
    required this.searchKeywords,
    required this.specialRequirements,
    required this.contentLanguage,
    required this.sourceSettings,
  });

  factory CategoryDto.fromJson(Map<String, Object?> json) {
    final sources = json['source_settings'];
    if (sources is! List) {
      throw const FormatException('source_settings must be a list.');
    }
    return CategoryDto(
      id: _requiredString(json, 'id'),
      name: _requiredString(json, 'name'),
      searchKeywords: _optionalString(json['search_keywords']),
      specialRequirements: _optionalString(json['special_requirements']),
      contentLanguage: json.containsKey('content_language')
          ? _requiredString(json, 'content_language')
          : traditionalChineseContentLanguage,
      sourceSettings: [
        for (final source in sources)
          if (source is Map)
            SourceSettingDto.fromJson(Map<String, Object?>.from(source))
          else
            throw const FormatException(
              'source_settings entries must be objects.',
            ),
      ],
    );
  }

  final String id;
  final String name;
  final String? searchKeywords;
  final String? specialRequirements;
  final String contentLanguage;
  final List<SourceSettingDto> sourceSettings;
}

final class SourceSettingDto {
  const SourceSettingDto({
    required this.id,
    required this.label,
    required this.websiteInput,
    required this.normalizedHost,
    required this.kind,
    required this.position,
  });

  factory SourceSettingDto.fromJson(Map<String, Object?> json) {
    final position = json['position'];
    if (position is! int) {
      throw const FormatException('position must be an integer.');
    }
    return SourceSettingDto(
      id: _requiredString(json, 'id'),
      label: _requiredString(json, 'label', allowEmpty: true),
      websiteInput: _requiredString(json, 'website_input', allowEmpty: true),
      normalizedHost: _optionalString(json['normalized_host']),
      kind: _requiredString(json, 'kind'),
      position: position,
    );
  }

  final String id;
  final String label;
  final String websiteInput;
  final String? normalizedHost;
  final String kind;
  final int position;
}

String _requiredString(
  Map<String, Object?> json,
  String key, {
  bool allowEmpty = false,
}) {
  final value = json[key];
  if (value is String && (allowEmpty || value.isNotEmpty)) {
    return value;
  }
  throw FormatException('$key must be a string.');
}

String? _optionalString(Object? value) {
  if (value == null || value is String) {
    return value as String?;
  }
  throw const FormatException('Expected a string or null.');
}
