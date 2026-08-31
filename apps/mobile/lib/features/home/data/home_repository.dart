import '../../../core/api/api_failure.dart';
import 'home_refresh_status.dart';
import 'home_remote_service.dart';

abstract interface class HomeRepository {
  Future<HomeRefreshStatus> loadRefreshStatus();
}

enum HomeRepositoryFailureKind { unavailable, invalidResponse }

final class HomeRepositoryFailure implements Exception {
  const HomeRepositoryFailure(this.kind);

  final HomeRepositoryFailureKind kind;
}

final class RemoteHomeRepository implements HomeRepository {
  RemoteHomeRepository(this._remoteService);

  final HomeRemoteService _remoteService;

  @override
  Future<HomeRefreshStatus> loadRefreshStatus() async {
    try {
      final dto = await _remoteService.loadLatestIngestionRun();
      return HomeRefreshStatus(
        lastSuccessfulAt: dto.lastSuccessfulAt,
        activeRunStatus: switch (dto.activeRun?.status) {
          'queued' => ActiveRunStatus.queued,
          'running' => ActiveRunStatus.running,
          _ => ActiveRunStatus.none,
        },
      );
    } on ApiFailure catch (failure) {
      throw HomeRepositoryFailure(
        failure.kind == ApiFailureKind.invalidResponse
            ? HomeRepositoryFailureKind.invalidResponse
            : HomeRepositoryFailureKind.unavailable,
      );
    } on FormatException {
      throw const HomeRepositoryFailure(
        HomeRepositoryFailureKind.invalidResponse,
      );
    } on TypeError {
      throw const HomeRepositoryFailure(
        HomeRepositoryFailureKind.invalidResponse,
      );
    }
  }
}
