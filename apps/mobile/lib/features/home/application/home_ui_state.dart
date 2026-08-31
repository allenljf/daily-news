import 'package:flutter/foundation.dart';

enum HomeRunUiStatus { idle, queued, running }

@immutable
final class HomeUiState {
  const HomeUiState({required this.lastSuccessfulAt, required this.runStatus});

  final DateTime? lastSuccessfulAt;
  final HomeRunUiStatus runStatus;

  bool get canRequestUpdate => runStatus == HomeRunUiStatus.idle;
}

enum HomeLoadFailure implements Exception { unavailable }
