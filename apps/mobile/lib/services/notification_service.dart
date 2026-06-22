import 'package:awesome_notifications/awesome_notifications.dart';
import 'package:flutter/material.dart';

class NotificationService {
  static bool _initialized = false;

  /// Initialize notification service
  static Future<void> initialize() async {
    if (_initialized) return;

    await AwesomeNotifications().initialize(
      'resource://drawable/ic_notification',
      [
        NotificationChannel(
          channelKey: 'sync_channel',
          channelName: 'Sync Notifications',
          channelDescription: 'Notifications for sync operations',
          defaultColor: Colors.blue,
          ledColor: Colors.blue,
          importance: NotificationImportance.High,
          channelShowBadge: true,
        ),
        NotificationChannel(
          channelKey: 'security_channel',
          channelName: 'Security Alerts',
          channelDescription: 'Important security notifications',
          defaultColor: Colors.red,
          ledColor: Colors.red,
          importance: NotificationImportance.High,
          channelShowBadge: true,
          soundSource: 'resource://raw/security_alert',
        ),
        NotificationChannel(
          channelKey: 'session_channel',
          channelName: 'Session Notifications',
          channelDescription: 'Session related notifications',
          defaultColor: Colors.green,
          ledColor: Colors.green,
          importance: NotificationImportance.Default,
        ),
        NotificationChannel(
          channelKey: 'general_channel',
          channelName: 'General',
          channelDescription: 'General notifications',
          defaultColor: Colors.grey,
          importance: NotificationImportance.Low,
        ),
      ],
    );

    // Request permission
    await AwesomeNotifications().requestPermissionToSendNotifications();

    _initialized = true;
  }

  /// Show sync completion notification
  static Future<void> showSyncCompleteNotification({
    required int uploaded,
    required int downloaded,
    int conflicts = 0,
  }) async {
    String body = 'Uploaded: $uploaded, Downloaded: $downloaded';
    if (conflicts > 0) {
      body += ' | $conflicts conflict${conflicts > 1 ? 's' : ''} need attention';
    }

    await AwesomeNotifications().createNotification(
      content: NotificationContent(
        id: 100,
        channelKey: 'sync_channel',
        title: 'Sync Complete',
        body: body,
        notificationLayout: NotificationLayout.Default,
      ),
    );
  }

  /// Show sync conflict notification
  static Future<void> showSyncConflictNotification({required int conflicts}) async {
    await AwesomeNotifications().createNotification(
      content: NotificationContent(
        id: 101,
        channelKey: 'sync_channel',
        title: 'Sync Conflicts',
        body: '$conflicts item${conflicts > 1 ? 's' : ''} need your attention',
        notificationLayout: NotificationLayout.Default,
        payload: {'screen': '/sync-conflicts'},
      ),
      actionButtons: [
        NotificationActionButton(
          key: 'resolve',
          label: 'Resolve',
        ),
        NotificationActionButton(
          key: 'dismiss',
          label: 'Dismiss',
        ),
      ],
    );
  }

  /// Show security alert
  static Future<void> showSecurityAlert({
    required String title,
    required String body,
  }) async {
    await AwesomeNotifications().createNotification(
      content: NotificationContent(
        id: 200,
        channelKey: 'security_channel',
        title: title,
        body: body,
        notificationLayout: NotificationLayout.BigText,
        criticalAlert: true,
      ),
    );
  }

  /// Show session notification
  static Future<void> showSessionNotification({
    required String title,
    required String body,
  }) async {
    await AwesomeNotifications().createNotification(
      content: NotificationContent(
        id: 300,
        channelKey: 'session_channel',
        title: title,
        body: body,
      ),
    );
  }

  /// Show local auth failure alert
  static Future<void> showAuthFailureNotification() async {
    await AwesomeNotifications().createNotification(
      content: NotificationContent(
        id: 201,
        channelKey: 'security_channel',
        title: 'Authentication Failed',
        body: 'Multiple failed biometric attempts detected. Please verify your identity.',
        criticalAlert: true,
      ),
    );
  }

  /// Cancel notification
  static Future<void> cancel(int id) async {
    await AwesomeNotifications().cancel(id);
  }

  /// Cancel all notifications
  static Future<void> cancelAll() async {
    await AwesomeNotifications().cancelAll();
  }
}
