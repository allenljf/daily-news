import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/http/http_providers.dart';
import '../data/manual_run.dart';
import '../data/manual_run_repository.dart';

final manualRunRepositoryProvider = Provider<ManualRunRepository>(
  (ref) => RemoteManualRunRepository(ref.watch(apiClientProvider)),
);

final class ManualRunUiState {
  const ManualRunUiState({this.status, this.failed = false});
  final ManualRunStatus? status;
  final bool failed;
  bool get isActive =>
      status == ManualRunStatus.queued || status == ManualRunStatus.running;
}

final manualRunControllerProvider =
    AsyncNotifierProvider<ManualRunController, ManualRunUiState>(
      ManualRunController.new,
    );

final class ManualRunController extends AsyncNotifier<ManualRunUiState> {
  @override
  ManualRunUiState build() => const ManualRunUiState();

  Future<void> request() async {
    final current = state.asData?.value;
    if (current?.isActive == true || state.isLoading) return;
    state = const AsyncLoading();
    try {
      final run = await ref
          .read(manualRunRepositoryProvider)
          .requestManualRun();
      state = AsyncData(ManualRunUiState(status: run.status));
    } on ManualRunRepositoryFailure {
      state = const AsyncData(ManualRunUiState(failed: true));
    }
  }
}
