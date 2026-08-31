import '../../../core/api/api_client.dart';
import 'news.dart';

final class NewsRemoteService {
  NewsRemoteService(this._apiClient);

  final ApiClient _apiClient;

  Future<NewsPage> loadPage(NewsPageRequest request) => _apiClient.get(
    'categories/${request.categoryId}/news',
    queryParameters: {
      'limit': request.limit,
      if (request.cursor != null) 'cursor': request.cursor,
      if (request.sourceTagId != null) 'sourceTagId': request.sourceTagId,
    },
    decode: (json) {
      final object = _object(json);
      final rawItems = object['items'];
      if (rawItems is! List) throw const FormatException('Expected items.');
      return NewsPage(
        items: rawItems.map((item) => _item(_object(item))).toList(),
        nextCursor: object['next_cursor'] as String?,
      );
    },
  );

  Future<NewsDetail> loadDetail(String categoryId, String newsId) =>
      _apiClient.get(
        'categories/$categoryId/news/$newsId',
        decode: (json) {
          final object = _object(json);
          return NewsDetail(
            item: _item(object),
            firstSeenAt: DateTime.parse(object['first_seen_at'] as String),
          );
        },
      );

  Future<void> setPermanent(String newsId, bool permanent) async {
    await _apiClient.patch<Object?>(
      'news/$newsId',
      data: {'permanent': permanent},
      decode: (json) => json,
    );
  }

  Future<void> deleteNews(String newsId) => _apiClient.delete('news/$newsId');

  static NewsItem _item(Map<String, Object?> json) {
    final expiresAt = json['expires_at'] as String?;
    return NewsItem(
      id: json['id'] as String,
      title: json['title'] as String,
      summary: json['summary'] as String?,
      canonicalUrl: Uri.parse(json['canonical_url'] as String),
      publishedAt: _date(json['published_at']),
      insertedAt: DateTime.parse(json['inserted_at'] as String),
      expiresAt: _date(expiresAt),
      sourceTagId: json['source_tag_id'] as String,
      sourceTagLabel: json['source_tag_label'] as String,
      isPermanent: expiresAt == null,
    );
  }

  static DateTime? _date(Object? value) =>
      value is String ? DateTime.parse(value) : null;

  static Map<String, Object?> _object(Object? value) {
    if (value is Map) return Map<String, Object?>.from(value);
    throw const FormatException('Expected an object.');
  }
}
