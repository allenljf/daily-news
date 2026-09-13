import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../features/categories/application/category_controller.dart';
import '../../../features/categories/presentation/category_home_section.dart';
import '../../../features/news/application/manual_run_providers.dart';
import '../../../features/news/data/manual_run.dart';
import '../../../features/news/presentation/manual_refresh_control.dart';
import '../../../l10n/app_localizations.dart';
import '../application/home_controller.dart';
import '../application/home_ui_state.dart';

const homeDataRefreshButtonKey = Key('home-data-refresh-button');

final class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final localizations = AppLocalizations.of(context);
    final state = ref.watch(homeControllerProvider);
    final manualRun = ref.watch(manualRunControllerProvider);

    return Scaffold(
      body: SafeArea(
        child: state.when(
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
          data: (homeState) {
            final manualStatus = manualRun.asData?.value.status;
            final effectiveState = HomeUiState(
              lastSuccessfulAt: homeState.lastSuccessfulAt,
              runStatus: manualStatus == ManualRunStatus.queued
                  ? HomeRunUiStatus.queued
                  : manualStatus == ManualRunStatus.running
                  ? HomeRunUiStatus.running
                  : homeState.runStatus,
            );
            return HomeContent(
              state: effectiveState,
              onImmediateUpdatePressed: () =>
                  showManualRefreshDialog(context, ref),
              onRefreshPagePressed: () async {
                await Future.wait([
                  ref.read(homeControllerProvider.notifier).reload(),
                  ref.read(categoriesControllerProvider.notifier).reload(),
                ]);
              },
              categoryContent: const CategoryHomeSection(),
            );
          },
        ),
      ),
    );
  }
}

final class HomeContent extends StatelessWidget {
  const HomeContent({
    required this.state,
    required this.onImmediateUpdatePressed,
    required this.onRefreshPagePressed,
    required this.categoryContent,
    super.key,
  });

  final HomeUiState state;
  final VoidCallback onImmediateUpdatePressed;
  final Future<void> Function() onRefreshPagePressed;
  final Widget categoryContent;

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context);
    final updatedAt = state.lastSuccessfulAt;
    final formattedUpdatedAt = updatedAt == null
        ? localizations.neverUpdated
        : DateFormat.yMd(Localizations.localeOf(context).toLanguageTag())
              .add_Hm()
              .format(updatedAt.toLocal());

    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(
            AppSpacing.large,
            AppSpacing.medium,
            AppSpacing.large,
            0,
          ),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
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
                  ],
                ),
              ),
              const SizedBox(width: AppSpacing.medium),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  FilledButton.icon(
                    key: manualUpdateButtonKey,
                    onPressed: state.canRequestUpdate
                        ? onImmediateUpdatePressed
                        : null,
                    icon: const Icon(Icons.refresh),
                    label: Text(localizations.updateNow),
                  ),
                  const SizedBox(height: AppSpacing.small),
                  OutlinedButton.icon(
                    key: homeDataRefreshButtonKey,
                    onPressed: onRefreshPagePressed,
                    icon: const Icon(Icons.refresh_outlined),
                    label: Text(localizations.refreshPage),
                  ),
                ],
              ),
            ],
          ),
        ),
        Expanded(child: categoryContent),
      ],
    );
  }
}
