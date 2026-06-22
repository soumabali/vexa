import 'dart:async';
import 'dart:convert';
import 'package:sqflite/sqflite.dart';
import 'package:path/path.dart';

class OfflineStorage {
  static Database? _database;
  static const String _dbName = 'ssh_manager_offline.db';
  static const int _dbVersion = 1;

  // Table names
  static const String _tableCredentials = 'credentials';
  static const String _tableHosts = 'hosts';
  static const String _tableSessions = 'sessions';
  static const String _tableSyncQueue = 'sync_queue';
  static const String _tableKeyStore = 'key_store';

  /// Get database instance
  static Future<Database> get database async {
    _database ??= await _initDatabase();
    return _database!;
  }

  /// Initialize database
  static Future<Database> _initDatabase() async {
    final dbPath = await getDatabasesPath();
    final path = join(dbPath, _dbName);

    return await openDatabase(
      path,
      version: _dbVersion,
      onCreate: _onCreate,
      onUpgrade: _onUpgrade,
    );
  }

  /// Create tables
  static Future<void> _onCreate(Database db, int version) async {
    // Credentials table
    await db.execute('''
      CREATE TABLE $_tableCredentials (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        type TEXT NOT NULL,
        username TEXT,
        password TEXT,
        private_key TEXT,
        passphrase TEXT,
        host_ids TEXT,
        tags TEXT,
        metadata TEXT,
        is_encrypted INTEGER DEFAULT 1,
        sync_status TEXT DEFAULT 'synced',
        local_modified_at INTEGER,
        server_modified_at INTEGER,
        deleted INTEGER DEFAULT 0
      )
    ''');

    // Hosts table
    await db.execute('''
      CREATE TABLE $_tableHosts (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        hostname TEXT NOT NULL,
        port INTEGER DEFAULT 22,
        protocol TEXT DEFAULT 'ssh',
        credential_id TEXT,
        group_id TEXT,
        tags TEXT,
        metadata TEXT,
        sync_status TEXT DEFAULT 'synced',
        local_modified_at INTEGER,
        server_modified_at INTEGER,
        deleted INTEGER DEFAULT 0
      )
    ''');

    // Sessions table
    await db.execute('''
      CREATE TABLE $_tableSessions (
        id TEXT PRIMARY KEY,
        host_id TEXT NOT NULL,
        connection_type TEXT DEFAULT 'ssh',
        started_at INTEGER NOT NULL,
        ended_at INTEGER,
        duration INTEGER,
        bytes_sent INTEGER DEFAULT 0,
        bytes_received INTEGER DEFAULT 0,
        commands TEXT,
        sync_status TEXT DEFAULT 'pending',
        local_modified_at INTEGER
      )
    ''');

    // Sync queue table
    await db.execute('''
      CREATE TABLE $_tableSyncQueue (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        entity_type TEXT NOT NULL,
        entity_id TEXT NOT NULL,
        operation TEXT NOT NULL,
        payload TEXT,
        priority INTEGER DEFAULT 5,
        retry_count INTEGER DEFAULT 0,
        max_retries INTEGER DEFAULT 3,
        created_at INTEGER NOT NULL,
        scheduled_at INTEGER NOT NULL,
        status TEXT DEFAULT 'pending'
      )
    ''');

    // Key store table (for encrypted local keys)
    await db.execute('''
      CREATE TABLE $_tableKeyStore (
        id TEXT PRIMARY KEY,
        key_type TEXT NOT NULL,
        public_key TEXT,
        encrypted_private_key TEXT,
        metadata TEXT,
        created_at INTEGER NOT NULL
      )
    ''');

    // Create indexes
    await db.execute('CREATE INDEX idx_credentials_sync ON $_tableCredentials(sync_status)');
    await db.execute('CREATE INDEX idx_hosts_sync ON $_tableHosts(sync_status)');
    await db.execute('CREATE INDEX idx_queue_status ON $_tableSyncQueue(status, scheduled_at)');
    await db.execute('CREATE INDEX idx_queue_entity ON $_tableSyncQueue(entity_type, entity_id)');
  }

  /// Upgrade database
  static Future<void> _onUpgrade(Database db, int oldVersion, int newVersion) async {
    // Handle migrations in future versions
  }

  /// Close database
  static Future<void> close() async {
    if (_database != null) {
      await _database!.close();
      _database = null;
    }
  }

  // === Credential Operations ===

  static Future<void> insertCredential(Map<String, dynamic> credential) async {
    final db = await database;
    credential['local_modified_at'] = DateTime.now().millisecondsSinceEpoch;
    credential['sync_status'] = 'pending';
    await db.insert(
      _tableCredentials,
      credential,
      conflictAlgorithm: ConflictAlgorithm.replace,
    );
  }

  static Future<List<Map<String, dynamic>>> getCredentials({String? syncStatus}) async {
    final db = await database;
    if (syncStatus != null) {
      return await db.query(
        _tableCredentials,
        where: 'sync_status = ? AND deleted = 0',
        whereArgs: [syncStatus],
      );
    }
    return await db.query(_tableCredentials, where: 'deleted = 0');
  }

  static Future<Map<String, dynamic>?> getCredential(String id) async {
    final db = await database;
    final results = await db.query(
      _tableCredentials,
      where: 'id = ? AND deleted = 0',
      whereArgs: [id],
    );
    return results.isNotEmpty ? results.first : null;
  }

  static Future<void> updateCredential(String id, Map<String, dynamic> data) async {
    final db = await database;
    data['local_modified_at'] = DateTime.now().millisecondsSinceEpoch;
    data['sync_status'] = 'pending';
    await db.update(
      _tableCredentials,
      data,
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  static Future<void> deleteCredential(String id) async {
    final db = await database;
    await db.update(
      _tableCredentials,
      {
        'deleted': 1,
        'sync_status': 'pending',
        'local_modified_at': DateTime.now().millisecondsSinceEpoch,
      },
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  // === Host Operations ===

  static Future<void> insertHost(Map<String, dynamic> host) async {
    final db = await database;
    host['local_modified_at'] = DateTime.now().millisecondsSinceEpoch;
    host['sync_status'] = 'pending';
    await db.insert(
      _tableHosts,
      host,
      conflictAlgorithm: ConflictAlgorithm.replace,
    );
  }

  static Future<List<Map<String, dynamic>>> getHosts({String? groupId}) async {
    final db = await database;
    if (groupId != null) {
      return await db.query(
        _tableHosts,
        where: 'group_id = ? AND deleted = 0',
        whereArgs: [groupId],
      );
    }
    return await db.query(_tableHosts, where: 'deleted = 0');
  }

  // === Sync Queue Operations ===

  static Future<void> enqueue({
    required String entityType,
    required String entityId,
    required String operation,
    Map<String, dynamic>? payload,
    int priority = 5,
    int maxRetries = 3,
  }) async {
    final db = await database;
    final now = DateTime.now().millisecondsSinceEpoch;
    
    await db.insert(_tableSyncQueue, {
      'entity_type': entityType,
      'entity_id': entityId,
      'operation': operation,
      'payload': payload != null ? jsonEncode(payload) : null,
      'priority': priority,
      'max_retries': maxRetries,
      'created_at': now,
      'scheduled_at': now,
      'status': 'pending',
    });
  }

  static Future<List<Map<String, dynamic>>> getPendingQueue({int limit = 100}) async {
    final db = await database;
    final now = DateTime.now().millisecondsSinceEpoch;
    
    return await db.query(
      _tableSyncQueue,
      where: 'status = ? AND scheduled_at <= ?',
      whereArgs: ['pending', now],
      orderBy: 'priority ASC, created_at ASC',
      limit: limit,
    );
  }

  static Future<void> markQueueItemProcessed(int id, {String status = 'completed'}) async {
    final db = await database;
    await db.update(
      _tableSyncQueue,
      {'status': status},
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  static Future<void> retryQueueItem(int id, {int delayMinutes = 5}) async {
    final db = await database;
    final now = DateTime.now().millisecondsSinceEpoch;
    final scheduledAt = now + (delayMinutes * 60 * 1000);
    
    await db.rawUpdate('''
      UPDATE $_tableSyncQueue 
      SET retry_count = retry_count + 1,
          scheduled_at = ?,
          status = CASE WHEN retry_count + 1 >= max_retries THEN 'failed' ELSE 'pending' END
      WHERE id = ?
    ''', [scheduledAt, id]);
  }

  // === Session Operations ===

  static Future<void> insertSession(Map<String, dynamic> session) async {
    final db = await database;
    session['local_modified_at'] = DateTime.now().millisecondsSinceEpoch;
    await db.insert(
      _tableSessions,
      session,
      conflictAlgorithm: ConflictAlgorithm.replace,
    );
  }

  static Future<List<Map<String, dynamic>>> getPendingSessions() async {
    final db = await database;
    return await db.query(
      _tableSessions,
      where: 'sync_status = ?',
      whereArgs: ['pending'],
    );
  }

  // === Statistics ===

  static Future<Map<String, int>> getStats() async {
    final db = await database;
    
    final credentialCount = Sqflite.firstIntValue(
      await db.rawQuery('SELECT COUNT(*) FROM $_tableCredentials WHERE deleted = 0'),
    ) ?? 0;
    
    final hostCount = Sqflite.firstIntValue(
      await db.rawQuery('SELECT COUNT(*) FROM $_tableHosts WHERE deleted = 0'),
    ) ?? 0;
    
    final pendingCount = Sqflite.firstIntValue(
      await db.rawQuery('SELECT COUNT(*) FROM $_tableSyncQueue WHERE status = ?', ['pending']),
    ) ?? 0;
    
    final failedCount = Sqflite.firstIntValue(
      await db.rawQuery('SELECT COUNT(*) FROM $_tableSyncQueue WHERE status = ?', ['failed']),
    ) ?? 0;

    return {
      'credentials': credentialCount,
      'hosts': hostCount,
      'pending_sync': pendingCount,
      'failed_sync': failedCount,
    };
  }
}
