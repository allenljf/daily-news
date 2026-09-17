import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:webview_flutter/webview_flutter.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../l10n/app_localizations.dart';

const articleWebViewKey = Key('article-web-view');
const articleWebViewBlockedKey = Key('article-web-view-blocked');
const articleWebViewErrorKey = Key('article-web-view-error');

typedef ArticleWebViewBuilder = Widget Function(BuildContext context, Uri url);

/// Composition seam so widget tests can replace the platform view.
final articleWebViewBuilderProvider = Provider<ArticleWebViewBuilder>(
  (ref) =>
      (context, url) => _PlatformArticleWebView(url: url),
);

bool isLoadableArticleUrl(Uri url) =>
    url.isScheme('https') && url.host.isNotEmpty;

bool _isAllowedNavigation(String rawUrl) {
  final uri = Uri.tryParse(rawUrl);
  return uri != null && (uri.isScheme('http') || uri.isScheme('https'));
}

final class ArticleWebView extends ConsumerWidget {
  const ArticleWebView({required this.url, super.key});

  final Uri url;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final localizations = AppLocalizations.of(context);
    if (!isLoadableArticleUrl(url)) {
      return Center(
        key: articleWebViewBlockedKey,
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.large),
          child: Text(
            localizations.newsOriginalPageUnavailable,
            textAlign: TextAlign.center,
          ),
        ),
      );
    }
    return ref.watch(articleWebViewBuilderProvider)(context, url);
  }
}

final class _PlatformArticleWebView extends StatefulWidget {
  const _PlatformArticleWebView({required this.url});

  final Uri url;

  @override
  State<_PlatformArticleWebView> createState() =>
      _PlatformArticleWebViewState();
}

final class _PlatformArticleWebViewState
    extends State<_PlatformArticleWebView> {
  late final WebViewController _controller;
  var _progress = 0;
  var _failed = false;

  @override
  void initState() {
    super.initState();
    _controller = WebViewController()
      ..setJavaScriptMode(JavaScriptMode.unrestricted)
      ..setNavigationDelegate(
        NavigationDelegate(
          onProgress: (progress) {
            if (mounted) {
              setState(() => _progress = progress);
            }
          },
          onWebResourceError: (error) {
            if ((error.isForMainFrame ?? false) && mounted) {
              setState(() => _failed = true);
            }
          },
          onNavigationRequest: (request) => _isAllowedNavigation(request.url)
              ? NavigationDecision.navigate
              : NavigationDecision.prevent,
        ),
      )
      ..loadRequest(widget.url);
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context);
    return Stack(
      children: [
        WebViewWidget(key: articleWebViewKey, controller: _controller),
        if (_progress < 100 && !_failed)
          const LinearProgressIndicator(minHeight: 2),
        if (_failed)
          ColoredBox(
            color: Theme.of(context).colorScheme.surface,
            child: Center(
              key: articleWebViewErrorKey,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(localizations.newsOriginalPageFailed),
                  const SizedBox(height: AppSpacing.small),
                  OutlinedButton(
                    onPressed: _reload,
                    child: Text(localizations.retry),
                  ),
                ],
              ),
            ),
          ),
      ],
    );
  }

  void _reload() {
    setState(() {
      _failed = false;
      _progress = 0;
    });
    unawaited(_controller.loadRequest(widget.url));
  }
}
