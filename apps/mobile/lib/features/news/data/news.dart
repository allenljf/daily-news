final class NewsPageRequest {
  const NewsPageRequest({
    required this.categoryId,
    this.cursor,
    this.sourceTagId,
    this.limit = 20,
  });

  final String categoryId;
  final String? cursor;
  final String? sourceTagId;
  final int limit;
}

final class NewsItem {
  const NewsItem({
    required this.id,
    required this.title,
    required this.canonicalUrl,
    required this.insertedAt,
    required this.sourceTagId,
    required this.sourceTagLabel,
    required this.isPermanent,
    this.summary,
    this.publishedAt,
    this.expiresAt,
  });

  final String id;
  final String title;
  final String? summary;
  final Uri canonicalUrl;
  final DateTime? publishedAt;
  final DateTime insertedAt;
  final DateTime? expiresAt;
  final String sourceTagId;
  final String sourceTagLabel;
  final bool isPermanent;

  NewsItem copyWith({bool? isPermanent, DateTime? expiresAt}) => NewsItem(
    id: id,
    title: title,
    summary: summary,
    canonicalUrl: canonicalUrl,
    publishedAt: publishedAt,
    insertedAt: insertedAt,
    expiresAt: isPermanent == true ? null : expiresAt ?? this.expiresAt,
    sourceTagId: sourceTagId,
    sourceTagLabel: sourceTagLabel,
    isPermanent: isPermanent ?? this.isPermanent,
  );
}

final class NewsPage {
  NewsPage({required List<NewsItem> items, required this.nextCursor})
    : items = List.unmodifiable(items);

  final List<NewsItem> items;
  final String? nextCursor;
}

final class NewsDetail {
  const NewsDetail({required this.item, required this.firstSeenAt});

  factory NewsDetail.fromItem(NewsItem item, {required DateTime firstSeenAt}) =>
      NewsDetail(item: item, firstSeenAt: firstSeenAt);

  final NewsItem item;
  final DateTime firstSeenAt;

  NewsDetail copyWith({NewsItem? item}) =>
      NewsDetail(item: item ?? this.item, firstSeenAt: firstSeenAt);
}
