import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_zh.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations)!;
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[Locale('zh')];

  /// No description provided for @appTitle.
  ///
  /// In zh, this message translates to:
  /// **'每日新聞'**
  String get appTitle;

  /// No description provided for @loading.
  ///
  /// In zh, this message translates to:
  /// **'正在載入…'**
  String get loading;

  /// No description provided for @signInWithGoogle.
  ///
  /// In zh, this message translates to:
  /// **'使用 Google 登入'**
  String get signInWithGoogle;

  /// No description provided for @signInFailed.
  ///
  /// In zh, this message translates to:
  /// **'登入失敗，請重新嘗試'**
  String get signInFailed;

  /// No description provided for @lastUpdated.
  ///
  /// In zh, this message translates to:
  /// **'最近更新時間'**
  String get lastUpdated;

  /// No description provided for @neverUpdated.
  ///
  /// In zh, this message translates to:
  /// **'尚未更新'**
  String get neverUpdated;

  /// No description provided for @updateNow.
  ///
  /// In zh, this message translates to:
  /// **'立即更新'**
  String get updateNow;

  /// No description provided for @runQueued.
  ///
  /// In zh, this message translates to:
  /// **'排隊中'**
  String get runQueued;

  /// No description provided for @runRunning.
  ///
  /// In zh, this message translates to:
  /// **'更新中'**
  String get runRunning;

  /// No description provided for @homeLoadFailed.
  ///
  /// In zh, this message translates to:
  /// **'無法讀取更新狀態'**
  String get homeLoadFailed;

  /// No description provided for @retry.
  ///
  /// In zh, this message translates to:
  /// **'重試'**
  String get retry;

  /// No description provided for @addCategory.
  ///
  /// In zh, this message translates to:
  /// **'新增新聞類別'**
  String get addCategory;

  /// No description provided for @categoriesLoadFailed.
  ///
  /// In zh, this message translates to:
  /// **'無法讀取新聞類別'**
  String get categoriesLoadFailed;

  /// No description provided for @categoryName.
  ///
  /// In zh, this message translates to:
  /// **'新聞類別'**
  String get categoryName;

  /// No description provided for @categoryNameRequired.
  ///
  /// In zh, this message translates to:
  /// **'請輸入新聞類別'**
  String get categoryNameRequired;

  /// No description provided for @searchKeywords.
  ///
  /// In zh, this message translates to:
  /// **'搜尋關鍵字'**
  String get searchKeywords;

  /// No description provided for @searchWebsites.
  ///
  /// In zh, this message translates to:
  /// **'搜尋網站'**
  String get searchWebsites;

  /// No description provided for @unspecifiedWebsite.
  ///
  /// In zh, this message translates to:
  /// **'未指定網站'**
  String get unspecifiedWebsite;

  /// No description provided for @addSearchWebsite.
  ///
  /// In zh, this message translates to:
  /// **'新增搜尋網站'**
  String get addSearchWebsite;

  /// No description provided for @websiteNameOrUrl.
  ///
  /// In zh, this message translates to:
  /// **'網站名稱或 URL'**
  String get websiteNameOrUrl;

  /// No description provided for @add.
  ///
  /// In zh, this message translates to:
  /// **'新增'**
  String get add;

  /// No description provided for @cancel.
  ///
  /// In zh, this message translates to:
  /// **'取消'**
  String get cancel;

  /// No description provided for @specialRequirements.
  ///
  /// In zh, this message translates to:
  /// **'其他特殊需求'**
  String get specialRequirements;

  /// No description provided for @saveSettings.
  ///
  /// In zh, this message translates to:
  /// **'儲存設定'**
  String get saveSettings;

  /// No description provided for @saving.
  ///
  /// In zh, this message translates to:
  /// **'儲存中…'**
  String get saving;

  /// No description provided for @categorySaveFailed.
  ///
  /// In zh, this message translates to:
  /// **'儲存失敗，請重新嘗試'**
  String get categorySaveFailed;

  /// No description provided for @news.
  ///
  /// In zh, this message translates to:
  /// **'新聞'**
  String get news;

  /// No description provided for @newsDetail.
  ///
  /// In zh, this message translates to:
  /// **'新聞詳情'**
  String get newsDetail;

  /// No description provided for @newsLoadFailed.
  ///
  /// In zh, this message translates to:
  /// **'無法讀取新聞'**
  String get newsLoadFailed;

  /// No description provided for @noNews.
  ///
  /// In zh, this message translates to:
  /// **'目前沒有新聞'**
  String get noNews;

  /// No description provided for @makePermanent.
  ///
  /// In zh, this message translates to:
  /// **'設為永久'**
  String get makePermanent;

  /// No description provided for @savedPermanently.
  ///
  /// In zh, this message translates to:
  /// **'已永久保存'**
  String get savedPermanently;

  /// No description provided for @deleteNews.
  ///
  /// In zh, this message translates to:
  /// **'刪除新聞'**
  String get deleteNews;

  /// No description provided for @newsDeleted.
  ///
  /// In zh, this message translates to:
  /// **'新聞已刪除'**
  String get newsDeleted;

  /// No description provided for @manualRefreshTitle.
  ///
  /// In zh, this message translates to:
  /// **'確認立即更新'**
  String get manualRefreshTitle;

  /// No description provided for @manualRefreshExplanation.
  ///
  /// In zh, this message translates to:
  /// **'確認後會啟動後端背景擷取工作，所需時間可能不同；你可以離開 App，新聞不會立即出現。'**
  String get manualRefreshExplanation;

  /// No description provided for @confirmUpdate.
  ///
  /// In zh, this message translates to:
  /// **'確認更新'**
  String get confirmUpdate;

  /// No description provided for @manualRunFailed.
  ///
  /// In zh, this message translates to:
  /// **'無法啟動更新，請重新嘗試'**
  String get manualRunFailed;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['zh'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'zh':
      return AppLocalizationsZh();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
