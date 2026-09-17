import 'package:flutter/material.dart';

import '../../l10n/app_localizations.dart';
import 'app_spacing.dart';
import 'app_theme.dart';

/// Shown while the app finishes starting up. It mirrors the native launch
/// screen (brand background plus the newspaper logo) so the hand-off from the
/// iOS LaunchScreen to the first Flutter frame is seamless.
final class BrandedLoadingScreen extends StatelessWidget {
  const BrandedLoadingScreen({super.key});

  static const double _logoSize = 110;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.brandIndigo,
      body: Stack(
        children: [
          const Center(
            child: Image(
              image: AssetImage('assets/images/launch_logo.png'),
              width: _logoSize,
              height: _logoSize,
            ),
          ),
          Positioned(
            left: 0,
            right: 0,
            bottom: 96,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const SizedBox(
                  width: 28,
                  height: 28,
                  child: CircularProgressIndicator(
                    strokeWidth: 3,
                    valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                  ),
                ),
                const SizedBox(height: AppSpacing.medium),
                Text(
                  AppLocalizations.of(context).loading,
                  style: const TextStyle(color: Colors.white),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
