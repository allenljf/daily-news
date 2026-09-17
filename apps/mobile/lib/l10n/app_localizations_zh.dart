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
  String get refreshPage => '刷新頁面';

  @override
  String get runQueued => '排隊中';

  @override
  String get runRunning => '更新中';

  @override
  String get homeLoadFailed => '無法讀取更新狀態';

  @override
  String get retry => '重試';

  @override
  String get addCategory => '新增新聞類別';

  @override
  String get categoriesLoadFailed => '無法讀取新聞類別';

  @override
  String get categoryName => '新聞類別';

  @override
  String get categoryNameRequired => '請輸入新聞類別';

  @override
  String get searchKeywords => '搜尋關鍵字';

  @override
  String get searchWebsites => '搜尋網站';

  @override
  String get unspecifiedWebsite => '未指定網站';

  @override
  String get addSearchWebsite => '新增搜尋網站';

  @override
  String get removeSearchWebsite => '移除搜尋網站';

  @override
  String get websiteNameOrUrl => '網站名稱或 URL';

  @override
  String get add => '新增';

  @override
  String get cancel => '取消';

  @override
  String get specialRequirements => '其他特殊需求';

  @override
  String get contentLanguage => '內容語言';

  @override
  String get english => '英文';

  @override
  String get traditionalChinese => '繁體中文';

  @override
  String get saveSettings => '儲存設定';

  @override
  String get saving => '儲存中…';

  @override
  String get categorySaveFailed => '儲存失敗，請重新嘗試';

  @override
  String get deleteCategory => '刪除新聞類別';

  @override
  String get editCategory => '編輯新聞類別';

  @override
  String get deleteCategoryExplanation => '會停止後續擷取並移除這個類別的關聯。其他類別仍引用的文章不會被刪除。';

  @override
  String get delete => '刪除';

  @override
  String get categoryDeleteFailed => '無法刪除新聞類別，請重新嘗試';

  @override
  String get news => '新聞';

  @override
  String get newsDetail => '新聞詳情';

  @override
  String get newsLoadFailed => '無法讀取新聞';

  @override
  String get noNews => '目前沒有資料';

  @override
  String get makePermanent => '設為永久';

  @override
  String get restoreExpiry => '恢復原本時效';

  @override
  String get deleteNews => '刪除新聞';

  @override
  String get newsDeleteFailed => '刪除失敗，請重新嘗試';

  @override
  String get newsOriginalPageUnavailable => '無法顯示原文網頁';

  @override
  String get newsOriginalPageFailed => '原文網頁載入失敗';

  @override
  String get manualRefreshTitle => '確認立即更新';

  @override
  String get manualRefreshExplanation =>
      '確認後會啟動後端背景擷取工作，所需時間可能不同；你可以離開 App，新聞不會立即出現。';

  @override
  String get confirmUpdate => '確認更新';

  @override
  String get manualRunFailed => '無法啟動更新，請重新嘗試';
}
