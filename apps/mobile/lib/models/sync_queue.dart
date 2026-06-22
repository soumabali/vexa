import 'dart:convert';

enum SyncOperation {
  create,
  update,
  delete,
}

enum SyncStatus {
  pending,
  processing,
  completed,
  failed,
  retrying,
}

enum EntityType {
  credential,
  host,
  session,
  settings,
}

class SyncQueueItem {
  final int? id;
  final EntityType entityType;
  final String entityId;
  final SyncOperation operation;
  final Map<String, dynamic>? payload;
  final int priority;
  int retryCount;
  final int maxRetries;
  final DateTime createdAt;
  DateTime scheduledAt;
  SyncStatus status;

  SyncQueueItem({
    this.id,
    required this.entityType,
    required this.entityId,
    required this.operation,
    this.payload,
    this.priority = 5,
    this.retryCount = 0,
    this.maxRetries = 3,
    required this.createdAt,
    required this.scheduledAt,
    this.status = SyncStatus.pending,
  });

  factory SyncQueueItem.fromMap(Map<String, dynamic> map) {
    return SyncQueueItem(
      id: map['id'],
      entityType: EntityType.values.firstWhere(
        (e) => e.toString() == 'EntityType.${map['entity_type']}',
        orElse: () => EntityType.credential,
      ),
      entityId: map['entity_id'],
      operation: SyncOperation.values.firstWhere(
        (e) => e.toString() == 'SyncOperation.${map['operation']}',
        orElse: () => SyncOperation.create,
      ),
      payload: map['payload'] != null ? jsonDecode(map['payload']) : null,
      priority: map['priority'] ?? 5,
      retryCount: map['retry_count'] ?? 0,
      maxRetries: map['max_retries'] ?? 3,
      createdAt: DateTime.fromMillisecondsSinceEpoch(map['created_at']),
      scheduledAt: DateTime.fromMillisecondsSinceEpoch(map['scheduled_at']),
      status: SyncStatus.values.firstWhere(
        (e) => e.toString() == 'SyncStatus.${map['status']}',
        orElse: () => SyncStatus.pending,
      ),
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'id': id,
      'entity_type': entityType.toString().split('.').last,
      'entity_id': entityId,
      'operation': operation.toString().split('.').last,
      'payload': payload != null ? jsonEncode(payload) : null,
      'priority': priority,
      'retry_count': retryCount,
      'max_retries': maxRetries,
      'created_at': createdAt.millisecondsSinceEpoch,
      'scheduled_at': scheduledAt.millisecondsSinceEpoch,
      'status': status.toString().split('.').last,
    };
  }

  /// Check if can retry
  bool get canRetry => retryCount < maxRetries && status != SyncStatus.completed;

  /// Calculate next retry delay in minutes
  int get nextRetryDelay {
    final delays = [1, 5, 15, 60, 240];
    return delays[retryCount.clamp(0, delays.length - 1)];
  }

  SyncQueueItem copyWith({
    int? id,
    EntityType? entityType,
    String? entityId,
    SyncOperation? operation,
    Map<String, dynamic>? payload,
    int? priority,
    int? retryCount,
    int? maxRetries,
    DateTime? createdAt,
    DateTime? scheduledAt,
    SyncStatus? status,
  }) {
    return SyncQueueItem(
      id: id ?? this.id,
      entityType: entityType ?? this.entityType,
      entityId: entityId ?? this.entityId,
      operation: operation ?? this.operation,
      payload: payload ?? this.payload,
      priority: priority ?? this.priority,
      retryCount: retryCount ?? this.retryCount,
      maxRetries: maxRetries ?? this.maxRetries,
      createdAt: createdAt ?? this.createdAt,
      scheduledAt: scheduledAt ?? this.scheduledAt,
      status: status ?? this.status,
    );
  }
}
