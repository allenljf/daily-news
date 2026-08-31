import 'package:flutter/material.dart';

class DailyNewsApp extends StatelessWidget {
  const DailyNewsApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Daily News',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.indigo),
        useMaterial3: true,
      ),
      home: const _ProtectedLoadingScreen(),
    );
  }
}

class _ProtectedLoadingScreen extends StatelessWidget {
  const _ProtectedLoadingScreen();

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            CircularProgressIndicator(),
            SizedBox(height: 16),
            Text('正在載入…'),
          ],
        ),
      ),
    );
  }
}
