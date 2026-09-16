const traditionalChineseContentLanguage = 'zh-Hant';
const englishContentLanguage = 'en';

final class Category {
  Category({
    required this.id,
    required this.name,
    this.searchKeywords,
    this.specialRequirements,
    this.contentLanguage = traditionalChineseContentLanguage,
    List<SourceSetting> sourceSettings = const [],
  }) : sourceSettings = List.unmodifiable(sourceSettings);

  final String id;
  final String name;
  final String? searchKeywords;
  final String? specialRequirements;
  final String contentLanguage;
  final List<SourceSetting> sourceSettings;
}

final class SourceSetting {
  const SourceSetting({
    required this.id,
    required this.label,
    required this.websiteInput,
    required this.kind,
    required this.position,
    this.normalizedHost,
  });

  final String id;
  final String label;
  final String websiteInput;
  final String? normalizedHost;
  final String kind;
  final int position;
}

final class CategoryDraft {
  CategoryDraft({
    required this.name,
    required this.searchKeywords,
    required this.specialRequirements,
    this.contentLanguage = traditionalChineseContentLanguage,
    required List<SourceSettingDraft> sourceSettings,
  }) : sourceSettings = List.unmodifiable(sourceSettings);

  final String name;
  final String? searchKeywords;
  final String? specialRequirements;
  final String contentLanguage;
  final List<SourceSettingDraft> sourceSettings;
}

final class SourceSettingDraft {
  const SourceSettingDraft({
    required this.label,
    required this.websiteInput,
    required this.kind,
  });

  final String label;
  final String websiteInput;
  final String kind;

  @override
  bool operator ==(Object other) {
    return other is SourceSettingDraft &&
        other.label == label &&
        other.websiteInput == websiteInput &&
        other.kind == kind;
  }

  @override
  int get hashCode => Object.hash(label, websiteInput, kind);
}
