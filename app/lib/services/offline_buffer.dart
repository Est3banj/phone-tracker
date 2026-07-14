import 'dart:convert';
import 'package:sqflite/sqflite.dart';
import 'package:path_provider/path_provider.dart';
import '../models/location_report.dart';

/// Offline SQLite buffer for location reports.
/// FIFO queue, max 10k entries, 7-day expiry.
class OfflineBuffer {
  Database? _db;
  static const int maxEntries = 10000;
  static const Duration maxAge = Duration(days: 7);

  Future<void> init() async {
    final dir = await getApplicationDocumentsDirectory();
    _db = await openDatabase(
      '${dir.path}/offline_buffer.db',
      version: 1,
      onCreate: (db, version) async {
        await db.execute('''
          CREATE TABLE pending_reports (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            payload TEXT NOT NULL,
            created_at TEXT NOT NULL
          )
        ''');
      },
    );
  }

  /// Buffer a report (FIFO, max 10k)
  Future<void> buffer(LocationReport report) async {
    if (_db == null) return;

    await _db!.transaction((txn) async {
      // Check count
      final count = Sqflite.firstIntValue(
        await txn.rawQuery('SELECT COUNT(*) FROM pending_reports'),
      )!;

      if (count >= maxEntries) {
        // Remove oldest
        await txn.rawDelete(
          'DELETE FROM pending_reports WHERE id IN (SELECT id FROM pending_reports ORDER BY id ASC LIMIT 1)',
        );
      }

      await txn.rawInsert(
        'INSERT INTO pending_reports (payload, created_at) VALUES (?, ?)',
        [jsonEncode(report.toJson()), report.timestamp.toIso8601String()],
      );
    });
  }

  /// Flush all pending reports and return them in FIFO order
  Future<List<Map<String, dynamic>>> flush() async {
    if (_db == null) return [];

    final reports = await _db!.rawQuery(
      'SELECT id, payload FROM pending_reports ORDER BY id ASC',
    );

    // Clear the buffer
    await _db!.rawDelete('DELETE FROM pending_reports');

    // Also clean old entries
    await _cleanOldEntries();

    return reports
        .map((r) => jsonDecode(r['payload'] as String) as Map<String, dynamic>)
        .toList();
  }

  /// Remove entries older than 7 days
  Future<void> _cleanOldEntries() async {
    if (_db == null) return;
    final cutoff = DateTime.now().subtract(maxAge).toIso8601String();
    await _db!.rawDelete(
      'DELETE FROM pending_reports WHERE created_at < ?',
      [cutoff],
    );
  }

  /// Get pending count
  Future<int> pendingCount() async {
    if (_db == null) return 0;
    return Sqflite.firstIntValue(
      await _db!.rawQuery('SELECT COUNT(*) FROM pending_reports'),
    )!;
  }

  Future<void> dispose() async {
    await _db?.close();
    _db = null;
  }
}
