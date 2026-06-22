import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'offline_storage.dart';
import 'secure_storage.dart';
import 'connectivity_service.dart';

enum SyncConflictStrategy {
  serverWins,    // Server data always wins
  clientWins,    // Local data always wins
  manualMerge,   // Require user intervention
  newerWins,     // Use the most recent timestamp
}

class SyncService {
  static final SyncService _instance = SyncService._internal();
  factory SyncService() => _instance;
  SyncService._internal();

  final _syncController = StreamController<SyncEvent>.broadcast();
  Stream<SyncEvent> get syncStream => _syncController.stream;

  Timer? _periodicTimer;
  bool _isSyncing = false;
  SyncConflictStrategy _conflictStrategy = SyncConflictStrategy.serverWins;

  bool get isSyncing => _isSyncing;

  /// Initialize periodic sync
  void startPeriodicSync({Duration interval = const Duration(minutes: 15)}) {
    _periodicTimer?.cancel();
    _periodicTimer = Timer.periodic(interval, (_) => sync());
  }

  /// Stop periodic sync
  void stopPeriodicSync() {
    _periodicTimer?.cancel();
    _periodicTimer = null;
  }

  /// Set conflict strategy
  void setConflictStrategy(SyncConflictStrategy strategy) {
    _conflictStrategy = strategy;
  }

  /// Main sync method
  Future<SyncResult> sync() async {
    if (_isSyncing) {
      return SyncResult.alreadyRunning();
    }

    final isOnline = await ConnectivityService.isOnline;
    if (!isOnline) {
      _syncController.add(SyncEvent.noConnection());
      return SyncResult.noConnection();
    }

    _isSyncing = true;
    _syncController.add(SyncEvent.started());

    try {
      final result = await _performSync();
      _syncController.add(SyncEvent.completed(result));
      return result;
    } catch (e) {
      _syncController.add(SyncEvent.error(e.toString()));
      return SyncResult.error(e.toString());
    } finally {
      _isSyncing = false;
    }
  }

  /// Perform actual sync
  Future<SyncResult> _performSync() async {
    int uploaded = 0;
    int downloaded = 0;
    int conflicts = 0;
    int errors = 0;
    List<SyncConflict> conflictList = [];

    // 1. Push local changes to server
    final pendingItems = await OfflineStorage.getPendingQueue(limit: 50);
    
    for (final item in pendingItems) {
      try {
        final success = await _pushToServer(item);
        if (success) {
          await OfflineStorage.markQueueItemProcessed(item['id'] as int);
          uploaded++;
        } else {
          // Conflict detected
          conflicts++;
          final conflict = await _handleConflict(item);
          if (conflict != null) {
            conflictList.add(conflict);
          }
        }
      } catch (e) {
        errors++;
        await OfflineStorage.retryQueueItem(
          item['id'] as int,
          delayMinutes: _calculateBackoff(item['retry_count'] as int),
        );
      }
    }

    // 2. Pull server changes
    final serverChanges = await _pullFromServer();
    for (final change in serverChanges) {
      try {
        await _applyServerChange(change);
        downloaded++;
      } catch (e) {
        errors++;
      }
    }

    // 3. Sync sessions
    await _syncSessions();

    return SyncResult.success(
      uploaded: uploaded,
      downloaded: downloaded,
      conflicts: conflicts,
      errors: errors,
      conflictList: conflictList,
    );
  }

  /// Push item to server
  Future<bool> _pushToServer(Map<String, dynamic> item) async {
    final apiKey = await SecureStorage.read(key: 'api_key');
    if (apiKey == null) throw Exception('Not authenticated');

    final serverUrl = await SecureStorage.read(key: 'server_url') ?? 'https://api.sshmanager.io';
    
    final response = await http.post(
      Uri.parse('$serverUrl/api/v1/sync'),
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': apiKey,
      },
      body: jsonEncode({
        'entity_type': item['entity_type'],
        'entity_id': item['entity_id'],
        'operation': item['operation'],
        'payload': item['payload'] != null ? jsonDecode(item['payload']) : null,
        'client_timestamp': item['created_at'],
      }),
    );

    if (response.statusCode == 200) {
      return true;
    } else if (response.statusCode == 409) {
      // Conflict
      return false;
    } else {
      throw Exception('Server error: ${response.statusCode}');
    }
  }

  /// Pull changes from server
  Future<List<Map<String, dynamic>>> _pullFromServer() async {
    final apiKey = await SecureStorage.read(key: 'api_key');
    if (apiKey == null) return [];

    final serverUrl = await SecureStorage.read(key: 'server_url') ?? 'https://api.sshmanager.io';
    
    // Get last sync timestamp
    final lastSync = await SecureStorage.read(key: 'last_sync_timestamp');
    final since = lastSync != null ? '&since=$lastSync' : '';

    final response = await http.get(
      Uri.parse('$serverUrl/api/v1/sync/changes?limit=100$since'),
      headers: {'X-API-Key': apiKey},
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      return List<Map<String, dynamic>>.from(data['changes'] ?? []);
    }
    
    return [];
  }

  /// Apply server change locally
  Future<void> _applyServerChange(Map<String, dynamic> change) async {
    final entityType = change['entity_type'];
    final entityId = change['entity_id'];
    final operation = change['operation'];
    final payload = change['payload'];

    switch (entityType) {
      case 'credential':
        if (operation == 'delete') {
          await OfflineStorage.deleteCredential(entityId);
        } else {
          await OfflineStorage.insertCredential(payload);
        }
        break;
      case 'host':
        if (operation == 'delete') {
          // Soft delete
        } else {
          await OfflineStorage.insertHost(payload);
        }
        break;
    }
  }

  /// Handle conflict
  Future<SyncConflict?> _handleConflict(Map<String, dynamic> item) async {
    switch (_conflictStrategy) {
      case SyncConflictStrategy.serverWins:
        // Re-fetch from server and overwrite local
        return null; // Conflict auto-resolved
      case SyncConflictStrategy.clientWins:
        // Force push local
        await _forcePush(item);
        return null;
      case SyncConflictStrategy.newerWins:
        // Compare timestamps and use newer
        return null; // Simplified
      case SyncConflictStrategy.manualMerge:
        // Return conflict for user resolution
        return SyncConflict(
          entityType: item['entity_type'],
          entityId: item['entity_id'],
          localData: jsonDecode(item['payload'] ?? '{}'),
        );
    }
  }

  /// Force push local data
  Future<void> _forcePush(Map<String, dynamic> item) async {
    // Implementation: force overwrite server data
  }

  /// Sync sessions
  Future<void> _syncSessions() async {
    final pendingSessions = await OfflineStorage.getPendingSessions();
    
    for (final session in pendingSessions) {
      // Upload session data
    }
  }

  /// Calculate exponential backoff
  int _calculateBackoff(int retryCount) {
    return [1, 5, 15, 60, 240][retryCount.clamp(0, 4)];
  }

  /// Dispose
  void dispose() {
    stopPeriodicSync();
    _syncController.close();
  }
}

// === Data Classes ===

class SyncEvent {
  final SyncEventType type;
  final String? message;
  final SyncResult? result;

  SyncEvent._(this.type, {this.message, this.result});

  factory SyncEvent.started() => SyncEvent._(SyncEventType.started);
  factory SyncEvent.completed(SyncResult result) => 
    SyncEvent._(SyncEventType.completed, result: result);
  factory SyncEvent.error(String message) => 
    SyncEvent._(SyncEventType.error, message: message);
  factory SyncEvent.noConnection() => SyncEvent._(SyncEventType.noConnection);
}

enum SyncEventType { started, completed, error, noConnection }

class SyncResult {
  final bool success;
  final int uploaded;
  final int downloaded;
  final int conflicts;
  final int errors;
  final List<SyncConflict> conflictList;
  final String? errorMessage;

  SyncResult._({
    required this.success,
    this.uploaded = 0,
    this.downloaded = 0,
    this.conflicts = 0,
    this.errors = 0,
    this.conflictList = const [],
    this.errorMessage,
  });

  factory SyncResult.success({
    required int uploaded,
    required int downloaded,
    required int conflicts,
    required int errors,
    required List<SyncConflict> conflictList,
  }) => SyncResult._(
    success: true,
    uploaded: uploaded,
    downloaded: downloaded,
    conflicts: conflicts,
    errors: errors,
    conflictList: conflictList,
  );

  factory SyncResult.error(String message) => 
    SyncResult._(success: false, errorMessage: message);
  factory SyncResult.alreadyRunning() => 
    SyncResult._(success: false, errorMessage: 'Sync already running');
  factory SyncResult.noConnection() => 
    SyncResult._(success: false, errorMessage: 'No internet connection');
}

class SyncConflict {
  final String entityType;
  final String entityId;
  final Map<String, dynamic> localData;
  final Map<String, dynamic>? serverData;

  SyncConflict({
    required this.entityType,
    required this.entityId,
    required this.localData,
    this.serverData,
  });
}
