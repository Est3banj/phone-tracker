class RemoteCommand {
  final String cmdId;
  final String action;
  final Map<String, dynamic>? params;
  final DateTime timestamp;
  final String status;

  RemoteCommand({
    required this.cmdId,
    required this.action,
    this.params,
    required this.timestamp,
    this.status = 'pending',
  });

  static const String lockDevice = 'lock_device';
  static const String wipeDevice = 'wipe_device';
  static const String capturePhoto = 'capture_photo';
  static const String triggerAlarm = 'trigger_alarm';

  Map<String, dynamic> toJson() => {
        'type': 'command',
        'cmd_id': cmdId,
        'action': action,
        'params': params,
        'ts': timestamp.toIso8601String(),
      };

  Map<String, dynamic> ackJson() => {
        'type': 'ack',
        'cmd_id': cmdId,
        'status': 'received',
        'ts': DateTime.now().toIso8601String(),
      };

  Map<String, dynamic> resultJson(String resultStatus, {String? error}) => {
        'type': 'result',
        'cmd_id': cmdId,
        'status': resultStatus,
        'error': error,
        'ts': DateTime.now().toIso8601String(),
      };

  factory RemoteCommand.fromJson(Map<String, dynamic> json) {
    return RemoteCommand(
      cmdId: json['cmd_id'] as String,
      action: json['action'] as String,
      params: json['params'] as Map<String, dynamic>?,
      timestamp: json['ts'] != null
          ? DateTime.parse(json['ts'])
          : DateTime.now(),
      status: json['status'] as String? ?? 'pending',
    );
  }
}
