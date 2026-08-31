final class LatestIngestionRunDto {
  const LatestIngestionRunDto({
    required this.lastSuccessfulAt,
    required this.activeRun,
  });

  factory LatestIngestionRunDto.fromJson(Map<String, Object?> json) {
    final lastSuccessfulAt = _optionalDateTime(json['last_successful_at']);
    final activeRunJson = json['active_run'];
    if (activeRunJson != null && activeRunJson is! Map) {
      throw const FormatException('active_run must be a JSON object or null.');
    }

    return LatestIngestionRunDto(
      lastSuccessfulAt: lastSuccessfulAt,
      activeRun: activeRunJson == null
          ? null
          : IngestionRunDto.fromJson(
              Map<String, Object?>.from(activeRunJson as Map),
            ),
    );
  }

  final DateTime? lastSuccessfulAt;
  final IngestionRunDto? activeRun;
}

final class IngestionRunDto {
  const IngestionRunDto({
    required this.id,
    required this.trigger,
    required this.status,
    required this.startedAt,
    required this.finishedAt,
  });

  factory IngestionRunDto.fromJson(Map<String, Object?> json) {
    return IngestionRunDto(
      id: _requiredString(json, 'id'),
      trigger: _requiredString(json, 'trigger'),
      status: _requiredString(json, 'status'),
      startedAt: _requiredDateTime(json, 'started_at'),
      finishedAt: _optionalDateTime(json['finished_at']),
    );
  }

  final String id;
  final String trigger;
  final String status;
  final DateTime startedAt;
  final DateTime? finishedAt;
}

String _requiredString(Map<String, Object?> json, String key) {
  final value = json[key];
  if (value is String && value.isNotEmpty) {
    return value;
  }
  throw FormatException('$key must be a non-empty string.');
}

DateTime _requiredDateTime(Map<String, Object?> json, String key) {
  final value = _optionalDateTime(json[key]);
  if (value != null) {
    return value;
  }
  throw FormatException('$key must be an ISO-8601 timestamp.');
}

DateTime? _optionalDateTime(Object? value) {
  if (value == null) {
    return null;
  }
  if (value case final String raw) {
    final parsed = DateTime.tryParse(raw);
    if (parsed != null) {
      return parsed;
    }
  }
  throw const FormatException('Expected an ISO-8601 timestamp or null.');
}
