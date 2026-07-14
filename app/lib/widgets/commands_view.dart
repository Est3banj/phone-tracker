import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class CommandsView extends ConsumerStatefulWidget {
  const CommandsView({super.key});

  @override
  ConsumerState<CommandsView> createState() => _CommandsViewState();
}

class _CommandsViewState extends ConsumerState<CommandsView> {
  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        const Text(
          'Remote Commands',
          style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 16),
        _commandButton(
          icon: Icons.lock,
          label: 'Lock Device',
          color: Colors.orange,
          onPressed: () => _sendCommand('lock_device'),
        ),
        const SizedBox(height: 8),
        _commandButton(
          icon: Icons.delete_forever,
          label: 'Wipe Device',
          color: Colors.red,
          onPressed: () => _sendCommand('wipe_device'),
        ),
        const SizedBox(height: 8),
        _commandButton(
          icon: Icons.camera_alt,
          label: 'Capture Photo',
          color: Colors.blue,
          onPressed: () => _sendCommand('capture_photo'),
        ),
        const SizedBox(height: 8),
        _commandButton(
          icon: Icons.alarm,
          label: 'Trigger Alarm',
          color: Colors.purple,
          onPressed: () => _sendCommand('trigger_alarm'),
        ),
        const SizedBox(height: 24),
        const Text(
          'Command History',
          style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 8),
        const Center(child: Text('No commands sent yet')),
      ],
    );
  }

  Widget _commandButton({
    required IconData icon,
    required String label,
    required Color color,
    required VoidCallback onPressed,
  }) {
    return SizedBox(
      width: double.infinity,
      child: OutlinedButton.icon(
        onPressed: onPressed,
        icon: Icon(icon, color: color),
        label: Text(label),
        style: OutlinedButton.styleFrom(
          padding: const EdgeInsets.symmetric(vertical: 16),
          side: BorderSide(color: color),
        ),
      ),
    );
  }

  void _sendCommand(String action) {
    // In real impl: send via WebSocket service
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Command $action sent')),
    );
  }
}
