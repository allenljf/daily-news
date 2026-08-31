import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/home/presentation/home_screen.dart';
import '../../features/news/presentation/news_detail_screen.dart';
import '../../features/news/presentation/news_list_screen.dart';
import '../auth/auth_gate.dart';

final appRouterProvider = Provider<GoRouter>((ref) {
  final router = GoRouter(
    routes: [
      GoRoute(
        path: '/',
        builder: (context, state) =>
            const AuthGate(authenticatedBuilder: _buildAuthenticatedHome),
      ),
      GoRoute(
        path: '/categories/:categoryId/news',
        builder: (context, state) =>
            NewsListScreen(categoryId: state.pathParameters['categoryId']!),
        routes: [
          GoRoute(
            path: ':newsId',
            builder: (context, state) => NewsDetailScreen(
              categoryId: state.pathParameters['categoryId']!,
              newsId: state.pathParameters['newsId']!,
            ),
          ),
        ],
      ),
    ],
  );
  ref.onDispose(router.dispose);
  return router;
});

Widget _buildAuthenticatedHome(BuildContext context) => const HomeScreen();
