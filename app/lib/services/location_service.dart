import 'dart:async';
import 'package:geolocator/geolocator.dart';
import '../models/location_report.dart';

class LocationService {
  StreamSubscription<Position>? _positionSubscription;
  Timer? _reportTimer;
  int _intervalSeconds = 60;
  bool _isRunning = false;
  Position? _lastPosition;

  /// Configurable interval: 15-300 seconds, default 60
  int get intervalSeconds => _intervalSeconds;

  /// Accuracy threshold: discard reports > 100m
  static const double maxAccuracy = 100.0;

  final StreamController<LocationReport> _reportController =
      StreamController<LocationReport>.broadcast();

  Stream<LocationReport> get reports => _reportController.stream;
  bool get isRunning => _isRunning;

  /// Start location service with configurable interval
  Future<void> start({int intervalSeconds = 60}) async {
    if (_isRunning) return;

    _intervalSeconds = intervalSeconds.clamp(15, 300);

    bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) {
      await Geolocator.openLocationSettings();
      return;
    }

    LocationPermission permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) return;
    }
    if (permission == LocationPermission.deniedForever) return;

    _isRunning = true;

    // Continuous position updates for best accuracy
    _positionSubscription = Geolocator.getPositionStream(
      locationSettings: LocationSettings(
        accuracy: LocationAccuracy.high,
        distanceFilter: 10,
        timeLimit: Duration(seconds: _intervalSeconds),
      ),
    ).listen((Position position) {
      _lastPosition = position;
    });

    // Timer-based reporting at configured interval
    _reportTimer = Timer.periodic(
      Duration(seconds: _intervalSeconds),
      (_) => _sendReport(),
    );
  }

  void _sendReport() {
    if (_lastPosition == null) return;

    // Accuracy filter: discard > 100m
    if (_lastPosition!.accuracy > maxAccuracy) return;

    final report = LocationReport(
      latitude: _lastPosition!.latitude,
      longitude: _lastPosition!.longitude,
      altitude: _lastPosition!.altitude,
      accuracy: _lastPosition!.accuracy,
      speed: _lastPosition!.speed,
      battery: 0, // Set by BatteryService
      isCharging: false,
      timestamp: DateTime.now(),
    );

    _reportController.add(report);
  }

  void updateInterval(int seconds) {
    _intervalSeconds = seconds.clamp(15, 300);
    _reportTimer?.cancel();
    if (_isRunning) {
      _reportTimer = Timer.periodic(
        Duration(seconds: _intervalSeconds),
        (_) => _sendReport(),
      );
    }
  }

  void stop() {
    _isRunning = false;
    _positionSubscription?.cancel();
    _reportTimer?.cancel();
    _positionSubscription = null;
    _reportTimer = null;
  }

  void dispose() {
    stop();
    _reportController.close();
  }
}
