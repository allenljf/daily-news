import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/http/http_providers.dart';
import '../data/category_remote_service.dart';
import '../data/category_repository.dart';

final categoryRemoteServiceProvider = Provider<CategoryRemoteService>(
  (ref) => CategoryRemoteService(ref.watch(apiClientProvider)),
);

final categoryRepositoryProvider = Provider<CategoryRepository>(
  (ref) => RemoteCategoryRepository(ref.watch(categoryRemoteServiceProvider)),
);
