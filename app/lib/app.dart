import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'screens/lock_screen.dart';
import 'screens/dashboard.dart';

class PhoneTrackerApp extends ConsumerWidget {
  const PhoneTrackerApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return MaterialApp(
      title: 'Phone Tracker',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: Colors.indigo,
          brightness: Brightness.dark,
        ),
        useMaterial3: true,
        fontFamily: 'RobotoMono',
      ),
      home: const LockScreen(),
      routes: {
        '/dashboard': (context) => const Dashboard(),
      },
    );
  }
}
