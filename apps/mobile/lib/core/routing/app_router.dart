import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/home/presentation/home_screen.dart';
import '../auth/auth_gate.dart';

final appRouterProvider = Provider<GoRouter>((ref) {
  final router = GoRouter(
    routes: [
      GoRoute(
        path: '/',
        builder: (context, state) =>
            const AuthGate(authenticatedBuilder: _buildAuthenticatedHome),
      ),
    ],
  );
  ref.onDispose(router.dispose);
  return router;
});

Widget _buildAuthenticatedHome(BuildContext context) => const HomeScreen();
