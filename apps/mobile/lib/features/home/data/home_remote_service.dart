import '../../../core/api/api_client.dart';
import 'latest_ingestion_run_dto.dart';

final class HomeRemoteService {
  HomeRemoteService(this._apiClient);

  final ApiClient _apiClient;

  Future<LatestIngestionRunDto> loadLatestIngestionRun() {
    return _apiClient.get(
      'ingestion-runs/latest',
      decode: (json) {
        if (json case final Map<String, Object?> object) {
          return LatestIngestionRunDto.fromJson(object);
        }
        throw const FormatException('Expected a JSON object.');
      },
    );
  }
}
