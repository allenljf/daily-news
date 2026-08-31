import '../../../core/api/api_client.dart';
import '../../../core/api/api_failure.dart';
import 'manual_run.dart';

abstract interface class ManualRunRepository {
  Future<ManualRun> requestManualRun();
}

enum ManualRunRepositoryFailure implements Exception { unavailable }

final class RemoteManualRunRepository implements ManualRunRepository {
  RemoteManualRunRepository(this._apiClient);
  final ApiClient _apiClient;

  @override
  Future<ManualRun> requestManualRun() async {
    try {
      return await _apiClient.post(
        'ingestion-runs',
        data: const <String, Object?>{},
        decode: (json) {
          if (json is! Map) throw const FormatException('Expected an object.');
          final object = Map<String, Object?>.from(json);
          return ManualRun(
            id: object['id'] as String,
            status: ManualRunStatus.values.byName(object['status'] as String),
          );
        },
      );
    } on ApiFailure {
      throw ManualRunRepositoryFailure.unavailable;
    }
  }
}
