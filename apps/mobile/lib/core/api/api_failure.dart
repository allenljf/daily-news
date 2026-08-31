import 'problem_details_dto.dart';

enum ApiFailureKind { problem, http, network, cancelled, invalidResponse }

final class ApiFailure implements Exception {
  const ApiFailure._({
    required this.kind,
    required this.title,
    this.statusCode,
    this.detail,
    this.type,
    this.instance,
  });

  factory ApiFailure.problem(ProblemDetailsDto problem) => ApiFailure._(
    kind: ApiFailureKind.problem,
    title: problem.title,
    statusCode: problem.status,
    detail: problem.detail,
    type: problem.type,
    instance: problem.instance,
  );

  const ApiFailure.network()
    : this._(kind: ApiFailureKind.network, title: 'Network request failed');

  const ApiFailure.http(int statusCode)
    : this._(
        kind: ApiFailureKind.http,
        title: 'HTTP request failed',
        statusCode: statusCode,
      );

  const ApiFailure.cancelled()
    : this._(kind: ApiFailureKind.cancelled, title: 'Request cancelled');

  const ApiFailure.invalidResponse()
    : this._(
        kind: ApiFailureKind.invalidResponse,
        title: 'Invalid server response',
      );

  final ApiFailureKind kind;
  final String title;
  final int? statusCode;
  final String? detail;
  final String? type;
  final String? instance;

  @override
  String toString() => 'ApiFailure($kind, $statusCode, $title)';
}
