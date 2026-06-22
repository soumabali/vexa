import 'dart:async';
import 'package:workmanager/workmanager.dart';
import 'sync_service.dart';
import 'notification_service.dart';

const String _syncTaskName = 'ssh_manager_background_sync';
const String _cleanupTaskName = 'ssh_manager_cleanup';

@pragma('vm:entry-point')
void callbackDispatcher() {
  Workmanager().executeTask((task, inputData) async {
    switch (task) {
      case _syncTaskName:
        return await _performBackgroundSync();
      case _cleanupTaskName:
        return await _performCleanup();
      default:
        return Future.value(true);
    }
  });
}

Future<bool> _performBackgroundSync() async {
  try {
    final syncService = SyncService();
    final result = await syncService.sync();
    
    if (result.success && result.conflicts > 0) {
      await NotificationService.showSyncConflictNotification(
        conflicts: result.conflicts,
      );
    }
    
    return result.success;
  } catch (e) {
    return false;
  }
}

Future<bool> _performCleanup() async {
  try {
    // Cleanup old sessions, failed sync items, etc.
    return true;
  } catch (e) {
    return false;
  }
}

class BackgroundService {
  static bool _initialized = false;

  /// Initialize background service
  static Future<void> initialize() async {
    if (_initialized) return;

    await Workmanager().initialize(
      callbackDispatcher,
      isInDebugMode: false,
    );

    _initialized = true;
  }

  /// Schedule periodic sync
  static Future<void> schedulePeriodicSync({
    Duration frequency = const Duration(hours: 1),
  }) async {
    await Workmanager().registerPeriodicTask(
      'periodic-sync',
      _syncTaskName,
      frequency: frequency,
      constraints: Constraints(
        networkType: NetworkType.connected,
        requiresBatteryNotLow: true,
        requiresStorageNotLow: false,
      ),
      existingWorkPolicy: ExistingWorkPolicy.replace,
    );
  }

  /// Schedule daily cleanup
  static Future<void> scheduleDailyCleanup() async {
    await Workmanager().registerPeriodicTask(
      'daily-cleanup',
      _cleanupTaskName,
      frequency: const Duration(days: 1),
      constraints: Constraints(
        networkType: NetworkType.not_required,
        requiresBatteryNotLow: true,
      ),
      existingWorkPolicy: ExistingWorkPolicy.keep,
    );
  }

  /// Cancel all background tasks
  static Future<void> cancelAll() async {
    await Workmanager().cancelAll();
  }

  /// Cancel specific task
  static Future<void> cancelTask(String taskId) async {
    await Workmanager().cancelByUniqueName(taskId);
  }
}
