// ignore: unused_import
import 'package:intl/intl.dart' as intl;

import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Chinese (`zh`).
class AppLocalizationsZh extends AppLocalizations {
  AppLocalizationsZh([String locale = 'zh']) : super(locale);

  @override
  String get appTitle => '每日新聞';

  @override
  String get loading => '正在載入…';

  @override
  String get signInWithGoogle => '使用 Google 登入';

  @override
  String get signInFailed => '登入失敗，請重新嘗試';

  @override
  String get lastUpdated => '最近更新時間';

  @override
  String get neverUpdated => '尚未更新';

  @override
  String get updateNow => '立即更新';

  @override
  String get runQueued => '排隊中';

  @override
  String get runRunning => '更新中';

  @override
  String get homeLoadFailed => '無法讀取更新狀態';

  @override
  String get retry => '重試';
}
