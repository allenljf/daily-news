import 'package:daily_news_mobile/app.dart';
import 'package:daily_news_mobile/core/auth/auth_providers.dart';
import 'package:daily_news_mobile/core/auth/sign_in_screen.dart';
import 'package:daily_news_mobile/core/http/http_providers.dart';
import 'package:daily_news_mobile/features/categories/presentation/category_grid.dart';
import 'package:daily_news_mobile/features/categories/presentation/category_settings_sheet.dart';
import 'package:daily_news_mobile/features/news/presentation/article_web_view.dart';
import 'package:daily_news_mobile/features/news/presentation/manual_refresh_control.dart';
import 'package:daily_news_mobile/features/news/presentation/news_detail_screen.dart';
import 'package:daily_news_mobile/features/news/presentation/news_list_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

import '../test_support/fake_api_server.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('authenticated reader completes the daily news workflow', (
    tester,
  ) async {
    final fakeApi = FakeDailyNewsApiServer();
    addTearDown(fakeApi.close);

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authGatewayProvider.overrideWithValue(fakeApi.authGateway),
          apiBaseUrlProvider.overrideWithValue(
            Uri.parse('https://api.example.test/v1/'),
          ),
          httpClientAdapterProvider.overrideWithValue(fakeApi.adapter),
          articleWebViewBuilderProvider.overrideWithValue(
            (context, url) => const SizedBox.shrink(),
          ),
        ],
        child: const DailyNewsApp(locale: Locale('zh')),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(signInButtonKey));
    await tester.pumpAndSettle();
    expect(fakeApi.signedIn, isTrue);

    await tester.tap(find.byKey(addCategoryTileKey));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(categoryNameFieldKey), '科技');
    await tester.pumpAndSettle();
    await tester.ensureVisible(find.byKey(saveCategoryButtonKey));
    await tester.tap(find.byKey(saveCategoryButtonKey));
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey('category-category-1')), findsOneWidget);
    expect(find.text('科技'), findsOneWidget);

    await tester.tap(find.byKey(manualUpdateButtonKey));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(confirmManualUpdateButtonKey));
    await tester.pumpAndSettle();
    expect(find.text('排隊中'), findsOneWidget);
    expect(fakeApi.manualRunRequests, 1);

    await tester.tap(find.byKey(const ValueKey('category-category-1')));
    await tester.pumpAndSettle();
    expect(find.byKey(newsListItemKey('article-1')), findsOneWidget);
    expect(find.text('新的科技新聞'), findsOneWidget);

    await tester.tap(find.byKey(newsListItemKey('article-1')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(makePermanentButtonKey));
    await tester.pumpAndSettle();
    expect(find.byTooltip('恢復原本時效'), findsOneWidget);
    expect(fakeApi.articleIsPermanent, isTrue);

    await tester.tap(find.byKey(makePermanentButtonKey));
    await tester.pumpAndSettle();
    expect(find.byTooltip('設為永久'), findsOneWidget);
    expect(fakeApi.articleIsPermanent, isFalse);

    await tester.tap(find.byKey(makePermanentButtonKey));
    await tester.pumpAndSettle();
    expect(fakeApi.articleIsPermanent, isTrue);

    await tester.tap(find.byKey(deleteNewsButtonKey));
    await tester.pumpAndSettle();
    expect(fakeApi.articleDeleted, isTrue);
    expect(find.byType(NewsDetailScreen), findsNothing);
    expect(find.text('新的科技新聞'), findsNothing);
  });
}
