class LocationReport {
  final double latitude;
  final double longitude;
  final double? altitude;
  final double? accuracy;
  final double? speed;
  final int battery;
  final bool isCharging;
  final DateTime timestamp;

  LocationReport({
    required this.latitude,
    required this.longitude,
    this.altitude,
    this.accuracy,
    this.speed,
    required this.battery,
    required this.isCharging,
    required this.timestamp,
  });

  Map<String, dynamic> toJson() => {
        'type': 'location',
        'lat': latitude,
        'lng': longitude,
        'alt': altitude,
        'accuracy': accuracy,
        'speed': speed,
        'battery': battery,
        'charging': isCharging,
        'ts': timestamp.toIso8601String(),
      };

  factory LocationReport.fromJson(Map<String, dynamic> json) {
    return LocationReport(
      latitude: (json['lat'] as num).toDouble(),
      longitude: (json['lng'] as num).toDouble(),
      altitude: (json['alt'] as num?)?.toDouble(),
      accuracy: (json['accuracy'] as num?)?.toDouble(),
      speed: (json['speed'] as num?)?.toDouble(),
      battery: (json['battery'] as num).toInt(),
      isCharging: json['charging'] == true,
      timestamp: json['ts'] != null
          ? DateTime.parse(json['ts'])
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toSqlite() => {
        'latitude': latitude,
        'longitude': longitude,
        'altitude': altitude,
        'accuracy': accuracy,
        'speed': speed,
        'battery_level': battery,
        'is_charging': isCharging ? 1 : 0,
        'received_at': timestamp.toIso8601String(),
      };
}
