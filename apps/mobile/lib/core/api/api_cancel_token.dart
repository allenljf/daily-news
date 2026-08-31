import 'package:dio/dio.dart';

final class ApiCancelToken {
  final CancelToken _token = CancelToken();

  bool get isCancelled => _token.isCancelled;

  void cancel() => _token.cancel();

  CancelToken get transportToken => _token;
}
