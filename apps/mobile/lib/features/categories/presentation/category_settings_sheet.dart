import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../l10n/app_localizations.dart';
import '../application/category_controller.dart';
import '../application/category_ui_state.dart';
import '../data/category.dart';

const categoryNameFieldKey = Key('category-name-field');
const searchKeywordsFieldKey = Key('search-keywords-field');
const specialRequirementsFieldKey = Key('special-requirements-field');
const contentLanguageFieldKey = Key('content-language-field');
const addSourceSettingButtonKey = Key('add-source-setting-button');
const addSourceDialogFieldKey = Key('add-source-dialog-field');
const saveCategoryButtonKey = Key('save-category-button');

Key sourceSettingFieldKey(int index) => ValueKey('source-setting-$index');

final class CategorySettingsSheet extends ConsumerWidget {
  const CategorySettingsSheet({this.category, super.key});

  final CategoryItemUi? category;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final screenState = ref.watch(categoriesControllerProvider).asData?.value;
    final existing = category;
    return CategorySettingsForm(
      initialCategory: existing,
      isSubmitting: screenState?.isSubmitting ?? false,
      saveFailed: screenState?.saveFailed ?? false,
      onSave: (draft) => existing == null
          ? ref
                .read(categoriesControllerProvider.notifier)
                .createCategory(draft)
          : ref
                .read(categoriesControllerProvider.notifier)
                .updateCategory(existing.id, draft),
    );
  }
}

final class CategorySettingsForm extends StatefulWidget {
  const CategorySettingsForm({
    required this.isSubmitting,
    required this.saveFailed,
    required this.onSave,
    this.initialCategory,
    super.key,
  });

  final bool isSubmitting;
  final bool saveFailed;
  final Future<bool> Function(CategoryDraft draft) onSave;
  final CategoryItemUi? initialCategory;

  @override
  State<CategorySettingsForm> createState() => _CategorySettingsFormState();
}

final class _CategorySettingsFormState extends State<CategorySettingsForm> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _keywordsController = TextEditingController();
  final _specialRequirementsController = TextEditingController();
  String _contentLanguage = traditionalChineseContentLanguage;
  final List<TextEditingController> _sourceControllers = [
    TextEditingController(),
  ];

  @override
  void initState() {
    super.initState();
    final initial = widget.initialCategory;
    if (initial == null) {
      return;
    }
    _nameController.text = initial.name;
    _keywordsController.text = initial.searchKeywords ?? '';
    _specialRequirementsController.text = initial.specialRequirements ?? '';
    _contentLanguage = initial.contentLanguage;
    final sources = [
      for (final source in initial.sourceSettings)
        if (source.websiteInput.trim().isNotEmpty) source.websiteInput.trim(),
    ];
    if (sources.isEmpty) {
      return;
    }
    for (final controller in _sourceControllers) {
      controller.dispose();
    }
    _sourceControllers
      ..clear()
      ..addAll([
        for (final source in sources) TextEditingController(text: source),
      ]);
  }

  @override
  void dispose() {
    _nameController.dispose();
    _keywordsController.dispose();
    _specialRequirementsController.dispose();
    for (final controller in _sourceControllers) {
      controller.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context);
    return SafeArea(
      child: SingleChildScrollView(
        padding: EdgeInsets.fromLTRB(
          AppSpacing.large,
          AppSpacing.large,
          AppSpacing.large,
          MediaQuery.viewInsetsOf(context).bottom + AppSpacing.large,
        ),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextFormField(
                key: categoryNameFieldKey,
                controller: _nameController,
                decoration: InputDecoration(
                  labelText: localizations.categoryName,
                ),
                textInputAction: TextInputAction.next,
                validator: (value) => value == null || value.trim().isEmpty
                    ? localizations.categoryNameRequired
                    : null,
              ),
              const SizedBox(height: AppSpacing.medium),
              TextFormField(
                key: searchKeywordsFieldKey,
                controller: _keywordsController,
                decoration: InputDecoration(
                  labelText: localizations.searchKeywords,
                ),
                textInputAction: TextInputAction.next,
              ),
              const SizedBox(height: AppSpacing.medium),
              Row(
                children: [
                  Expanded(
                    child: Text(
                      localizations.searchWebsites,
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                  ),
                  IconButton(
                    key: addSourceSettingButtonKey,
                    tooltip: localizations.addSearchWebsite,
                    onPressed: widget.isSubmitting ? null : _addSourceSetting,
                    icon: const Icon(Icons.add),
                  ),
                ],
              ),
              for (var index = 0; index < _sourceControllers.length; index += 1)
                Padding(
                  padding: const EdgeInsets.only(bottom: AppSpacing.small),
                  child: TextFormField(
                    key: sourceSettingFieldKey(index),
                    controller: _sourceControllers[index],
                    decoration: InputDecoration(
                      labelText: index == 0
                          ? localizations.unspecifiedWebsite
                          : localizations.websiteNameOrUrl,
                    ),
                    textInputAction: TextInputAction.next,
                  ),
                ),
              const SizedBox(height: AppSpacing.small),
              TextFormField(
                key: specialRequirementsFieldKey,
                controller: _specialRequirementsController,
                decoration: InputDecoration(
                  labelText: localizations.specialRequirements,
                ),
                minLines: 3,
                maxLines: 5,
              ),
              const SizedBox(height: AppSpacing.medium),
              DropdownButtonFormField<String>(
                key: contentLanguageFieldKey,
                initialValue: _contentLanguage,
                decoration: InputDecoration(
                  labelText: localizations.contentLanguage,
                ),
                items: [
                  DropdownMenuItem(
                    value: traditionalChineseContentLanguage,
                    child: Text(localizations.traditionalChinese),
                  ),
                  DropdownMenuItem(
                    value: englishContentLanguage,
                    child: Text(localizations.english),
                  ),
                ],
                onChanged: widget.isSubmitting
                    ? null
                    : (value) {
                        if (value != null) {
                          setState(() => _contentLanguage = value);
                        }
                      },
              ),
              if (widget.saveFailed) ...[
                const SizedBox(height: AppSpacing.medium),
                Text(
                  localizations.categorySaveFailed,
                  style: Theme.of(context).textTheme.bodyMedium
                      ?.copyWith(color: Theme.of(context).colorScheme.error),
                ),
              ],
              const SizedBox(height: AppSpacing.large),
              FilledButton(
                key: saveCategoryButtonKey,
                onPressed: widget.isSubmitting ? null : _save,
                child: Text(
                  widget.isSubmitting
                      ? localizations.saving
                      : localizations.saveSettings,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _addSourceSetting() async {
    final source = await showDialog<String>(
      context: context,
      builder: (dialogContext) => const _AddSourceDialog(),
    );
    if (!mounted || source == null || source.isEmpty) {
      return;
    }
    setState(() => _sourceControllers.add(TextEditingController(text: source)));
  }

  Future<void> _save() async {
    if (!(_formKey.currentState?.validate() ?? false)) {
      return;
    }
    final draft = CategoryDraft(
      name: _nameController.text.trim(),
      searchKeywords: _nullableTrimmed(_keywordsController.text),
      specialRequirements: _nullableTrimmed(
        _specialRequirementsController.text,
      ),
      contentLanguage: _contentLanguage,
      sourceSettings: [
        for (final controller in _sourceControllers)
          if (controller.text.trim().isNotEmpty)
            SourceSettingDraft(
              label: controller.text.trim(),
              websiteInput: controller.text.trim(),
              kind: 'website',
            ),
      ],
    );
    final saved = await widget.onSave(draft);
    if (!mounted || !saved) {
      return;
    }
    Navigator.of(context).pop();
  }
}

final class _AddSourceDialog extends StatefulWidget {
  const _AddSourceDialog();

  @override
  State<_AddSourceDialog> createState() => _AddSourceDialogState();
}

final class _AddSourceDialogState extends State<_AddSourceDialog> {
  final _controller = TextEditingController();

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context);
    return AlertDialog(
      title: Text(localizations.addSearchWebsite),
      content: TextField(
        key: addSourceDialogFieldKey,
        controller: _controller,
        autofocus: true,
        decoration: InputDecoration(labelText: localizations.websiteNameOrUrl),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: Text(localizations.cancel),
        ),
        FilledButton(
          onPressed: () => Navigator.of(context).pop(_controller.text.trim()),
          child: Text(localizations.add),
        ),
      ],
    );
  }
}

String? _nullableTrimmed(String value) {
  final trimmed = value.trim();
  return trimmed.isEmpty ? null : trimmed;
}
