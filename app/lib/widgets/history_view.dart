import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/providers.dart';
import '../models/location_report.dart';

class HistoryView extends ConsumerWidget {
  const HistoryView({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final history = ref.watch(locationHistoryProvider);

    if (history.isEmpty) {
      return const Center(child: Text('No location records yet'));
    }

    return ListView.builder(
      itemCount: history.length,
      itemBuilder: (context, index) {
        final loc = history[index];
        return ListTile(
          leading: const Icon(Icons.location_on, size: 20),
          title: Text(
            '${loc.latitude.toStringAsFixed(4)}, ${loc.longitude.toStringAsFixed(4)}',
            style: const TextStyle(fontFamily: 'RobotoMono', fontSize: 14),
          ),
          subtitle: Text(
            '${loc.timestamp}  |  Bat: ${loc.battery}%${loc.isCharging ? " (charging)" : ""}',
            style: const TextStyle(fontSize: 12),
          ),
        );
      },
    );
  }
}
