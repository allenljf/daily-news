enum ActiveRunStatus { none, queued, running }

final class HomeRefreshStatus {
  const HomeRefreshStatus({
    required this.lastSuccessfulAt,
    required this.activeRunStatus,
  });

  final DateTime? lastSuccessfulAt;
  final ActiveRunStatus activeRunStatus;
}
