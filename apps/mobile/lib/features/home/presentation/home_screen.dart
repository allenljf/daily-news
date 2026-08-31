import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../l10n/app_localizations.dart';
import '../application/home_controller.dart';
import '../application/home_ui_state.dart';

final class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final localizations = AppLocalizations.of(context);
    final state = ref.watch(homeControllerProvider);

    return Scaffold(
      appBar: AppBar(title: Text(localizations.appTitle)),
      body: state.when(
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
              Text(localizations.homeLoadFailed),
              const SizedBox(height: AppSpacing.medium),
              FilledButton(
                onPressed: () =>
                    ref.read(homeControllerProvider.notifier).reload(),
                child: Text(localizations.retry),
              ),
            ],
          ),
        ),
        data: (homeState) => HomeContent(
          state: homeState,
          onImmediateUpdatePressed: _manualUpdateDeferredToNewsFeature,
        ),
      ),
    );
  }
}

void _manualUpdateDeferredToNewsFeature() {}

final class HomeContent extends StatelessWidget {
  const HomeContent({
    required this.state,
    required this.onImmediateUpdatePressed,
    super.key,
  });

  final HomeUiState state;
  final VoidCallback onImmediateUpdatePressed;

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context);
    final updatedAt = state.lastSuccessfulAt;
    final formattedUpdatedAt = updatedAt == null
        ? localizations.neverUpdated
        : DateFormat.yMd(Localizations.localeOf(context).toLanguageTag())
              .add_Hm()
              .format(updatedAt.toLocal());

    return Align(
      alignment: Alignment.topRight,
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.large),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Text(
              localizations.lastUpdated,
              style: Theme.of(context).textTheme.labelLarge,
            ),
            const SizedBox(height: AppSpacing.small),
            Text(formattedUpdatedAt),
            if (state.runStatus != HomeRunUiStatus.idle) ...[
              const SizedBox(height: AppSpacing.small),
              Text(switch (state.runStatus) {
                HomeRunUiStatus.queued => localizations.runQueued,
                HomeRunUiStatus.running => localizations.runRunning,
                HomeRunUiStatus.idle => '',
              }),
            ],
            const SizedBox(height: AppSpacing.medium),
            FilledButton.icon(
              onPressed: state.canRequestUpdate
                  ? onImmediateUpdatePressed
                  : null,
              icon: const Icon(Icons.refresh),
              label: Text(localizations.updateNow),
            ),
          ],
        ),
      ),
    );
  }
}
