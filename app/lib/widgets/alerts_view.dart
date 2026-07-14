import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/providers.dart';
import '../models/event_report.dart';

class AlertsView extends ConsumerWidget {
  const AlertsView({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final alerts = ref.watch(eventHistoryProvider);

    if (alerts.isEmpty) {
      return const Center(child: Text('No alerts yet'));
    }

    return ListView.builder(
      itemCount: alerts.length,
      itemBuilder: (context, index) {
        final alert = alerts[index];
        return ListTile(
          leading: _alertIcon(alert.eventType),
          title: Text(
            _alertTitle(alert.eventType),
            style: TextStyle(
              fontWeight: FontWeight.bold,
              color: _alertColor(alert.eventType),
            ),
          ),
          subtitle: Text(alert.timestamp.toIso8601String()),
          trailing: alert.payload != null
              ? const Icon(Icons.info_outline, size: 16)
              : null,
        );
      },
    );
  }

  Icon _alertIcon(String type) {
    switch (type) {
      case EventReport.simChange:
        return const Icon(Icons.sim_card, color: Colors.red);
      case EventReport.batteryLow:
        return const Icon(Icons.battery_alert, color: Colors.yellow);
      case EventReport.wifiDisconnected:
        return const Icon(Icons.wifi_off, color: Colors.yellow);
      case EventReport.powerOn:
        return const Icon(Icons.power, color: Colors.green);
      default:
        return const Icon(Icons.notification);
    }
  }

  Color _alertColor(String type) {
    switch (type) {
      case EventReport.simChange:
        return Colors.red.shade300;
      case EventReport.batteryLow:
        return Colors.yellow.shade700;
      case EventReport.wifiDisconnected:
        return Colors.yellow.shade700;
      case EventReport.powerOn:
        return Colors.green.shade300;
      default:
        return Colors.grey;
    }
  }

  String _alertTitle(String type) {
    switch (type) {
      case EventReport.simChange:
        return 'SIM Change Detected';
      case EventReport.batteryLow:
        return 'Battery Low (<15%)';
      case EventReport.wifiDisconnected:
        return 'WiFi Disconnected';
      case EventReport.powerOn:
        return 'Device Powered On';
      default:
        return type;
    }
  }
}
