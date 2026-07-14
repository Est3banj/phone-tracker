import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/providers.dart';
import '../models/location_report.dart';

class MapView extends ConsumerWidget {
  const MapView({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final latestLocations = ref.watch(locationReportsProvider);

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            latestLocations.when(
              data: (location) => _buildLocationInfo(location),
              error: (err, stack) => Text('Error: $err'),
              loading: () => const Text('Waiting for GPS data...'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildLocationInfo(LocationReport loc) {
    return Column(
      children: [
        const Icon(Icons.my_location, size: 64, color: Colors.indigo),
        const SizedBox(height: 16),
        Text(
          '${loc.latitude.toStringAsFixed(6)}, ${loc.longitude.toStringAsFixed(6)}',
          style: const TextStyle(fontSize: 18, fontFamily: 'RobotoMono'),
        ),
        const SizedBox(height: 8),
        if (loc.accuracy != null)
          Text('Accuracy: ${loc.accuracy!.toStringAsFixed(0)}m'),
        if (loc.speed != null)
          Text('Speed: ${loc.speed!.toStringAsFixed(1)} m/s'),
        const SizedBox(height: 16),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              loc.isCharging ? Icons.battery_charging_full : Icons.battery_full,
              size: 20,
            ),
            const SizedBox(width: 4),
            Text('${loc.battery}%'),
          ],
        ),
      ],
    );
  }
}
