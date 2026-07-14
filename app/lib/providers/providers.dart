import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/websocket_service.dart';
import '../services/location_service.dart';
import '../services/battery_service.dart';
import '../services/event_service.dart';
import '../services/offline_buffer.dart';
import '../models/location_report.dart';
import '../models/event_report.dart';
import 'dart:async';

// -- Services --
final wsServiceProvider = Provider<WebSocketService>((ref) {
  throw UnimplementedError('WS service must be initialized with URL and token');
});

final locationServiceProvider = Provider<LocationService>((ref) {
  return LocationService();
});

final batteryServiceProvider = Provider<BatteryService>((ref) {
  return BatteryService();
});

final eventServiceProvider = Provider<EventService>((ref) {
  return EventService();
});

final offlineBufferProvider = Provider<OfflineBuffer>((ref) {
  return OfflineBuffer();
});

// -- State providers --
final wsStateProvider = StreamProvider<Map<String, dynamic>>((ref) {
  final ws = ref.watch(wsServiceProvider);
  return ws.messages ?? const Stream.empty();
});

final locationReportsProvider = StreamProvider<LocationReport>((ref) {
  final locService = ref.watch(locationServiceProvider);
  return locService.reports;
});

final eventReportsProvider = StreamProvider<EventReport>((ref) {
  final evtService = ref.watch(eventServiceProvider);
  return evtService.events;
});

final batteryStateProvider = FutureProvider.autoDispose<BatteryState>((ref) async {
  final battery = ref.watch(batteryServiceProvider);
  return await battery.getState();
});

// -- Connection status --
final connectionStatusProvider = StateProvider<String>((ref) => 'disconnected');

// -- History --
final locationHistoryProvider = StateProvider<List<LocationReport>>((ref) => []);
final eventHistoryProvider = StateProvider<List<EventReport>>((ref) => []);
