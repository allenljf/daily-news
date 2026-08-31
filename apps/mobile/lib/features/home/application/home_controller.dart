import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/home_refresh_status.dart';
import '../data/home_repository.dart';
import 'home_providers.dart';
import 'home_ui_state.dart';

final homeControllerProvider =
    AsyncNotifierProvider<HomeController, HomeUiState>(HomeController.new);

final class HomeController extends AsyncNotifier<HomeUiState> {
  @override
  Future<HomeUiState> build() => _load(ref.watch(homeRepositoryProvider));

  Future<void> reload() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(
      () => _load(ref.read(homeRepositoryProvider)),
    );
  }

  Future<HomeUiState> _load(HomeRepository repository) async {
    try {
      final status = await repository.loadRefreshStatus();
      return HomeUiState(
        lastSuccessfulAt: status.lastSuccessfulAt,
        runStatus: switch (status.activeRunStatus) {
          ActiveRunStatus.none => HomeRunUiStatus.idle,
          ActiveRunStatus.queued => HomeRunUiStatus.queued,
          ActiveRunStatus.running => HomeRunUiStatus.running,
        },
      );
    } on HomeRepositoryFailure {
      throw HomeLoadFailure.unavailable;
    }
  }
}
