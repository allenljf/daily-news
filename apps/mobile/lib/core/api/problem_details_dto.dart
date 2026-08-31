import 'package:json_annotation/json_annotation.dart';

part 'problem_details_dto.g.dart';

@JsonSerializable()
final class ProblemDetailsDto {
  const ProblemDetailsDto({
    required this.title,
    required this.status,
    this.type,
    this.detail,
    this.instance,
  });

  factory ProblemDetailsDto.fromJson(Map<String, Object?> json) =>
      _$ProblemDetailsDtoFromJson(json);

  final String title;
  final int status;
  final String? type;
  final String? detail;
  final String? instance;

  Map<String, Object?> toJson() => _$ProblemDetailsDtoToJson(this);
}
