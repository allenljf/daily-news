import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../l10n/app_localizations.dart';
import '../application/manual_run_providers.dart';
import '../data/manual_run.dart';

const manualUpdateButtonKey = Key('manual-update-button');
const confirmManualUpdateButtonKey = Key('confirm-manual-update-button');

final class ManualRefreshControl extends ConsumerWidget {
  const ManualRefreshControl({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final localizations = AppLocalizations.of(context);
    final value = ref.watch(manualRunControllerProvider);
    final run = value.asData?.value;
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (run?.status case final status?)
          Text(
            status == ManualRunStatus.queued
                ? localizations.runQueued
                : status == ManualRunStatus.running
                ? localizations.runRunning
                : '',
          ),
        if (run?.failed == true) Text(localizations.manualRunFailed),
        FilledButton.icon(
          key: manualUpdateButtonKey,
          onPressed: value.isLoading || run?.isActive == true
              ? null
              : () => showManualRefreshDialog(context, ref),
          icon: const Icon(Icons.refresh),
          label: Text(localizations.updateNow),
        ),
      ],
    );
  }
}

Future<void> showManualRefreshDialog(BuildContext context, WidgetRef ref) {
  final localizations = AppLocalizations.of(context);
  return showDialog<void>(
    context: context,
    builder: (context) => AlertDialog(
      title: Text(localizations.manualRefreshTitle),
      content: Text(localizations.manualRefreshExplanation),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(localizations.cancel),
        ),
        FilledButton(
          key: confirmManualUpdateButtonKey,
          onPressed: () {
            Navigator.pop(context);
            unawaited(ref.read(manualRunControllerProvider.notifier).request());
          },
          child: Text(localizations.confirmUpdate),
        ),
      ],
    ),
  );
}
