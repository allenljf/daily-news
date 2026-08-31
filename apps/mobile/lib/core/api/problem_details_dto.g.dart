// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'problem_details_dto.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ProblemDetailsDto _$ProblemDetailsDtoFromJson(Map<String, dynamic> json) =>
    ProblemDetailsDto(
      title: json['title'] as String,
      status: (json['status'] as num).toInt(),
      type: json['type'] as String?,
      detail: json['detail'] as String?,
      instance: json['instance'] as String?,
    );

Map<String, dynamic> _$ProblemDetailsDtoToJson(ProblemDetailsDto instance) =>
    <String, dynamic>{
      'title': instance.title,
      'status': instance.status,
      'type': instance.type,
      'detail': instance.detail,
      'instance': instance.instance,
    };
