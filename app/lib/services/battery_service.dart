import 'dart:async';
import 'package:flutter/services.dart';

class BatteryService {
  static const _channel = MethodChannel('phone_tracker/battery');
  final StreamController<BatteryState> _stateController =
      StreamController<BatteryState>.broadcast();

  Stream<BatteryState> get stateStream => _stateController.stream;

  /// Get current battery level (0-100)
  Future<int> getBatteryLevel() async {
    try {
      final level = await _channel.invokeMethod<int>('getBatteryLevel');
      return level ?? 0;
    } catch (_) {
      return 0;
    }
  }

  /// Get whether device is charging
  Future<bool> isCharging() async {
    try {
      final charging = await _channel.invokeMethod<bool>('isCharging');
      return charging ?? false;
    } catch (_) {
      return false;
    }
  }

  /// Get full battery state
  Future<BatteryState> getState() async {
    final level = await getBatteryLevel();
    final charging = await isCharging();
    return BatteryState(level: level, isCharging: charging);
  }

  void dispose() {
    _stateController.close();
  }
}

class BatteryState {
  final int level;
  final bool isCharging;

  BatteryState({required this.level, required this.isCharging});

  bool get isLow => level < 15;
}
