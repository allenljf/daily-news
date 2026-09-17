import 'package:flutter/material.dart';

abstract final class AppTheme {
  /// Matches the native iOS LaunchScreen and Android launcher background.
  static const Color brandIndigo = Color(0xFF4355B9);

  static const Color _seedColor = brandIndigo;

  static ThemeData light() => ThemeData(
    colorScheme: ColorScheme.fromSeed(seedColor: _seedColor),
    useMaterial3: true,
  );

  static ThemeData dark() => ThemeData(
    colorScheme: ColorScheme.fromSeed(
      seedColor: _seedColor,
      brightness: Brightness.dark,
    ),
    useMaterial3: true,
  );
}
