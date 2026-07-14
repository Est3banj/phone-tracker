import 'dart:async';
import 'dart:convert';
import 'dart:math';
import 'package:web_socket_channel/web_socket_channel.dart';

/// WebSocket connection state machine states
enum WsState {
  disconnected,
  connecting,
  backoffWait,
  connected,
  shutdown,
}

class WebSocketService {
  final String _baseUrl;
  final String _token;
  WebSocketChannel? _channel;
  WsState _state = WsState.disconnected;
  int _backoffSeconds = 1;
  Timer? _backoffTimer;
  Timer? _pingTimer;
  Timer? _timeoutTimer;
  StreamController<Map<String, dynamic>>? _messageController;
  StreamSubscription? _channelSubscription;
  bool _shouldReconnect = true;

  /// Max backoff: 60 seconds
  static const int maxBackoff = 60;
  /// Server pings every 30s, timeout if nothing for 90s
  static const Duration pingInterval = Duration(seconds: 30);
  static const Duration timeoutDuration = Duration(seconds: 90);

  WsState get state => _state;
  Stream<Map<String, dynamic>>? get messages => _messageController?.stream;

  WebSocketService({required String baseUrl, required String token})
      : _baseUrl = baseUrl,
        _token = token;

  /// Connect with exponential backoff
  void connect() {
    if (_state == WsState.shutdown) return;
    _setState(WsState.connecting);
    _shouldReconnect = true;

    try {
      final uri = Uri.parse('$_baseUrl/ws?token=$_token');
      _channel = WebSocketChannel.connect(uri);

      _messageController = StreamController<Map<String, dynamic>>.broadcast();
      _channelSubscription = _channel!.stream.listen(
        _onMessage,
        onError: _onError,
        onDone: _onDone,
        cancelOnError: false,
      );

      _setState(WsState.connected);
      _backoffSeconds = 1; // Reset backoff on success
      _startPingMonitor();
    } catch (e) {
      _onError(e);
    }
  }

  void _onMessage(dynamic data) {
    try {
      final msg = jsonDecode(data as String) as Map<String, dynamic>;
      _resetTimeout();

      // Handle ping from server
      if (msg['type'] == 'ping') {
        send({'type': 'pong'});
        return;
      }

      _messageController?.add(msg);
    } catch (_) {}
  }

  void _onError(dynamic error) {
    _cleanup();
    if (_state == WsState.shutdown) return;
    _scheduleBackoff();
  }

  void _onDone() {
    _cleanup();
    if (_state == WsState.shutdown) return;
    _scheduleBackoff();
  }

  void _scheduleBackoff() {
    _setState(WsState.backoffWait);
    final seconds = min(_backoffSeconds, maxBackoff);
    _backoffTimer = Timer(Duration(seconds: seconds), () {
      if (_shouldReconnect && _state != WsState.shutdown) {
        _backoffSeconds = min(_backoffSeconds * 2, maxBackoff);
        connect();
      }
    });
  }

  void _startPingMonitor() {
    _pingTimer?.cancel();
    _timeoutTimer?.cancel();
    _resetTimeout();
  }

  void _resetTimeout() {
    _timeoutTimer?.cancel();
    _timeoutTimer = Timer(timeoutDuration, () {
      // No message for 90s — consider dead
      disconnect();
      if (_shouldReconnect) _scheduleBackoff();
    });
  }

  /// Send JSON message
  void send(Map<String, dynamic> message) {
    if (_channel != null && _state == WsState.connected) {
      _channel!.sink.add(jsonEncode(message));
    }
  }

  /// Graceful disconnect (keeps shouldReconnect)
  void disconnect() {
    _shouldReconnect = false;
    _cleanup();
    _setState(WsState.disconnected);
  }

  /// Permanent shutdown
  void shutdown() {
    _shouldReconnect = false;
    _setState(WsState.shutdown);
    _cleanup();
  }

  void _cleanup() {
    _channelSubscription?.cancel();
    _channel?.sink.close();
    _channel = null;
    _backoffTimer?.cancel();
    _pingTimer?.cancel();
    _timeoutTimer?.cancel();
    _messageController?.close();
    _messageController = null;
  }

  void _setState(WsState newState) {
    _state = newState;
    // Notify state change via a separate stream if needed
  }
}
