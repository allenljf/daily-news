enum ManualRunStatus { queued, running, completed, failed }

final class ManualRun {
  const ManualRun({required this.id, required this.status});
  final String id;
  final ManualRunStatus status;
}
