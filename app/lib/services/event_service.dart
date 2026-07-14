import 'dart:async';
import 'package:flutter/services.dart';
import '../models/event_report.dart';

class EventService {
  static const _channel = MethodChannel('phone_tracker/events');

  final StreamController<EventReport> _eventController =
      StreamController<EventReport>.broadcast();

  Stream<EventReport> get events => _eventController.stream;

  /// Listen for native events from Android
  void startListening() {
    _channel.setMethodCallHandler((call) async {
      switch (call.method) {
        case 'onSimChange':
          _emitEvent(EventReport(
            eventType: EventReport.simChange,
            payload: call.arguments as Map<String, dynamic>?,
            timestamp: DateTime.now(),
          ));
          break;
        case 'onBatteryLow':
          _emitEvent(EventReport(
            eventType: EventReport.batteryLow,
            payload: call.arguments as Map<String, dynamic>?,
            timestamp: DateTime.now(),
          ));
          break;
        case 'onWiFiDisconnected':
          _emitEvent(EventReport(
            eventType: EventReport.wifiDisconnected,
            payload: call.arguments as Map<String, dynamic>?,
            timestamp: DateTime.now(),
          ));
          break;
        case 'onPowerOn':
          _emitEvent(EventReport(
            eventType: EventReport.powerOn,
            payload: call.arguments as Map<String, dynamic>?,
            timestamp: DateTime.now(),
          ));
          break;
      }
    });
  }

  void _emitEvent(EventReport event) {
    _eventController.add(event);
  }

  void dispose() {
    _channel.setMethodCallHandler(null);
    _eventController.close();
  }
}
