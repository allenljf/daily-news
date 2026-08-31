import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../l10n/app_localizations.dart';
import '../theme/app_spacing.dart';
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
          loading: () => Scaffold(
            body: Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const CircularProgressIndicator(),
                  const SizedBox(height: AppSpacing.medium),
                  Text(AppLocalizations.of(context).loading),
                ],
              ),
            ),
          ),
        );
  }
}
