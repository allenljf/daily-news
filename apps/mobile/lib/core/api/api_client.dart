import 'package:dio/dio.dart';

import 'api_cancel_token.dart';
import 'api_failure.dart';
import 'problem_details_dto.dart';

final class ApiClient {
  ApiClient(this._dio);

  final Dio _dio;

  Future<Map<String, Object?>> getJson(
    String path, {
    Map<String, Object?>? queryParameters,
    ApiCancelToken? cancelToken,
  }) => get(
    path,
    queryParameters: queryParameters,
    cancelToken: cancelToken,
    decode: _decodeJsonObject,
  );

  Future<T> get<T>(
    String path, {
    required T Function(Object? json) decode,
    Map<String, Object?>? queryParameters,
    ApiCancelToken? cancelToken,
  }) async {
    try {
      final response = await _dio.get<Object?>(
        path,
        queryParameters: queryParameters,
        cancelToken: cancelToken?.transportToken,
      );
      return decode(response.data);
    } on DioException catch (error) {
      throw _translate(error);
    } on FormatException {
      throw const ApiFailure.invalidResponse();
    } on TypeError {
      throw const ApiFailure.invalidResponse();
    }
  }

  Future<T> post<T>(
    String path, {
    required Object? data,
    required T Function(Object? json) decode,
    ApiCancelToken? cancelToken,
  }) async {
    try {
      final response = await _dio.post<Object?>(
        path,
        data: data,
        cancelToken: cancelToken?.transportToken,
      );
      return decode(response.data);
    } on DioException catch (error) {
      throw _translate(error);
    } on FormatException {
      throw const ApiFailure.invalidResponse();
    } on TypeError {
      throw const ApiFailure.invalidResponse();
    }
  }

  static Map<String, Object?> _decodeJsonObject(Object? json) {
    if (json case final Map<String, Object?> object) {
      return object;
    }
    throw const FormatException('Expected a JSON object.');
  }

  static ApiFailure _translate(DioException error) {
    if (CancelToken.isCancel(error)) {
      return const ApiFailure.cancelled();
    }

    final problem = _decodeProblem(error.response?.data);
    if (problem != null) {
      return ApiFailure.problem(problem);
    }
    final statusCode = error.response?.statusCode;
    return statusCode == null
        ? const ApiFailure.network()
        : ApiFailure.http(statusCode);
  }

  static ProblemDetailsDto? _decodeProblem(Object? data) {
    if (data is! Map) {
      return null;
    }
    final json = Map<String, Object?>.from(data);
    final detail = json['detail'];
    final problemJson = detail is Map
        ? Map<String, Object?>.from(detail)
        : json;
    try {
      return ProblemDetailsDto.fromJson(problemJson);
    } on TypeError {
      return null;
    }
  }
}
