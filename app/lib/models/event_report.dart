class EventReport {
  final String eventType;
  final Map<String, dynamic>? payload;
  final DateTime timestamp;

  EventReport({
    required this.eventType,
    this.payload,
    required this.timestamp,
  });

  static const String simChange = 'sim_change';
  static const String batteryLow = 'battery_low';
  static const String wifiDisconnected = 'wifi_disconnected';
  static const String powerOn = 'power_on';

  Map<String, dynamic> toJson() => {
        'type': 'event',
        'event_type': eventType,
        'payload': payload,
        'ts': timestamp.toIso8601String(),
      };

  factory EventReport.fromJson(Map<String, dynamic> json) {
    return EventReport(
      eventType: json['event_type'] as String,
      payload: json['payload'] as Map<String, dynamic>?,
      timestamp: json['ts'] != null
          ? DateTime.parse(json['ts'])
          : DateTime.now(),
    );
  }
}
