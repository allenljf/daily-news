import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../theme/branded_loading_screen.dart';
import 'auth_providers.dart';
import 'sign_in_screen.dart';

final class AuthGate extends ConsumerWidget {
  const AuthGate({required this.authenticatedBuilder, super.key});

  final WidgetBuilder authenticatedBuilder;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return ref
        .watch(authStateProvider)
        .when(
          data: (user) => user == null
              ? const SignInScreen()
              : authenticatedBuilder(context),
          error: (_, _) => const SignInScreen(),
          loading: () => const BrandedLoadingScreen(),
        );
  }
}
