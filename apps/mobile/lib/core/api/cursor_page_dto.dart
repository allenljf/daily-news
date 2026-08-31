import 'package:json_annotation/json_annotation.dart';

part 'cursor_page_dto.g.dart';

@JsonSerializable(genericArgumentFactories: true)
final class CursorPageDto<T> {
  const CursorPageDto({required this.items, required this.nextCursor});

  factory CursorPageDto.fromJson(
    Map<String, Object?> json,
    T Function(Object? json) fromJsonT,
  ) => _$CursorPageDtoFromJson(json, fromJsonT);

  final List<T> items;

  @JsonKey(name: 'next_cursor')
  final String? nextCursor;

  Map<String, Object?> toJson(Object? Function(T value) toJsonT) =>
      _$CursorPageDtoToJson(this, toJsonT);
}
