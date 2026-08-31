import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/http/http_providers.dart';
import '../data/home_remote_service.dart';
import '../data/home_repository.dart';

final homeRemoteServiceProvider = Provider<HomeRemoteService>(
  (ref) => HomeRemoteService(ref.watch(apiClientProvider)),
);

final homeRepositoryProvider = Provider<HomeRepository>(
  (ref) => RemoteHomeRepository(ref.watch(homeRemoteServiceProvider)),
);
