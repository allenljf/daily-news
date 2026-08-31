import '../../../core/api/api_failure.dart';
import 'news.dart';
import 'news_remote_service.dart';

abstract interface class NewsRepository {
  Future<NewsPage> loadNewsPage(NewsPageRequest request);
  Future<NewsDetail> loadNewsDetail(String categoryId, String newsId);
  Future<void> setPermanent(String newsId, bool permanent);
  Future<void> deleteNews(String newsId);
}

enum NewsRepositoryFailure implements Exception { unavailable }

final class RemoteNewsRepository implements NewsRepository {
  RemoteNewsRepository(this._service);
  final NewsRemoteService _service;

  @override
  Future<NewsPage> loadNewsPage(NewsPageRequest request) =>
      _translate(() => _service.loadPage(request));

  @override
  Future<NewsDetail> loadNewsDetail(String categoryId, String newsId) =>
      _translate(() => _service.loadDetail(categoryId, newsId));

  @override
  Future<void> setPermanent(String newsId, bool permanent) =>
      _translate(() => _service.setPermanent(newsId, permanent));

  @override
  Future<void> deleteNews(String newsId) =>
      _translate(() => _service.deleteNews(newsId));

  static Future<T> _translate<T>(Future<T> Function() action) async {
    try {
      return await action();
    } on ApiFailure {
      throw NewsRepositoryFailure.unavailable;
    }
  }
}
