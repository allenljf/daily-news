import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/http/http_providers.dart';
import '../data/news_remote_service.dart';
import '../data/news_repository.dart';

final newsRemoteServiceProvider = Provider<NewsRemoteService>(
  (ref) => NewsRemoteService(ref.watch(apiClientProvider)),
);

final newsRepositoryProvider = Provider<NewsRepository>(
  (ref) => RemoteNewsRepository(ref.watch(newsRemoteServiceProvider)),
);
