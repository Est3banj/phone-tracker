import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:local_auth/local_auth.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class LockScreen extends ConsumerStatefulWidget {
  const LockScreen({super.key});

  @override
  ConsumerState<LockScreen> createState() => _LockScreenState();
}

class _LockScreenState extends ConsumerState<LockScreen> {
  final _pinController = TextEditingController();
  final _localAuth = LocalAuthentication();
  final _storage = const FlutterSecureStorage();
  String? _storedPin;
  bool _isAuthenticating = false;
  int _failCount = 0;
  bool _isLocked = false;

  @override
  void initState() {
    super.initState();
    _loadPin();
    _tryBiometric();
  }

  Future<void> _loadPin() async {
    final pin = await _storage.read(key: 'app_pin');
    setState(() => _storedPin = pin);
  }

  Future<void> _tryBiometric() async {
    try {
      final available = await _localAuth.canCheckBiometrics;
      if (!available) return;

      final authenticated = await _localAuth.authenticate(
        localizedReason: 'Unlock Phone Tracker',
        options: const AuthenticationOptions(
          stickyAuth: true,
          biometricOnly: false,
        ),
      );

      if (authenticated && mounted) {
        Navigator.pushReplacementNamed(context, '/dashboard');
      }
    } catch (_) {}
  }

  void _submitPin() {
    if (_isLocked) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Locked out for 60 seconds')),
      );
      return;
    }

    if (_pinController.text == _storedPin) {
      _failCount = 0;
      Navigator.pushReplacementNamed(context, '/dashboard');
    } else {
      _failCount++;
      if (_failCount >= 5) {
        setState(() => _isLocked = true);
        // Send auth failure event
        Future.delayed(const Duration(seconds: 60), () {
          if (mounted) setState(() => _isLocked = false);
        });
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Incorrect PIN (${5 - _failCount} attempts left)')),
      );
    }
  }

  @override
  void dispose() {
    _pinController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.lock_outline,
                size: 80,
                color: Theme.of(context).colorScheme.primary,
              ),
              const SizedBox(height: 24),
              Text(
                'Phone Tracker',
                style: Theme.of(context).textTheme.headlineMedium,
              ),
              const SizedBox(height: 32),
              if (_storedPin != null) ...[
                TextField(
                  controller: _pinController,
                  obscureText: true,
                  maxLength: 6,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(
                    labelText: 'Enter PIN',
                    counterText: '',
                  ),
                ),
                const SizedBox(height: 16),
                FilledButton(
                  onPressed: _isLocked ? null : _submitPin,
                  child: const Text('Unlock'),
                ),
              ],
              if (_storedPin == null) ...[
                TextField(
                  controller: _pinController,
                  obscureText: true,
                  maxLength: 6,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(
                    labelText: 'Set PIN (6 digits)',
                    counterText: '',
                  ),
                ),
                const SizedBox(height: 16),
                FilledButton(
                  onPressed: () async {
                    if (_pinController.text.length == 6) {
                      await _storage.write(
                        key: 'app_pin',
                        value: _pinController.text,
                      );
                      setState(() => _storedPin = _pinController.text);
                    }
                  },
                  child: const Text('Set PIN'),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
