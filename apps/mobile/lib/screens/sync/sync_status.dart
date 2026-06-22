import 'dart:async';
import 'package:flutter/material.dart';
import '../../services/sync_service.dart';
import '../../services/offline_storage.dart';

class SyncStatusScreen extends StatefulWidget {
  const SyncStatusScreen({Key? key}) : super(key: key);

  @override
  _SyncStatusScreenState createState() => _SyncStatusScreenState();
}

class _SyncStatusScreenState extends State<SyncStatusScreen> {
  SyncResult? _lastResult;
  Map<String, int> _stats = {};
  bool _isSyncing = false;
  StreamSubscription? _syncSubscription;

  @override
  void initState() {
    super.initState();
    _loadStats();
    _listenToSync();
  }

  @override
  void dispose() {
    _syncSubscription?.cancel();
    super.dispose();
  }

  void _listenToSync() {
    _syncSubscription = SyncService().syncStream.listen((event) {
      if (mounted) {
        setState(() {
          _isSyncing = event.type == SyncEventType.started;
          if (event.type == SyncEventType.completed) {
            _lastResult = event.result;
          }
        });
        _loadStats();
      }
    });
  }

  Future<void> _loadStats() async {
    final stats = await OfflineStorage.getStats();
    if (mounted) {
      setState(() => _stats = stats);
    }
  }

  Future<void> _triggerSync() async {
    setState(() => _isSyncing = true);
    final result = await SyncService().sync();
    setState(() {
      _isSyncing = false;
      _lastResult = result;
    });
    _loadStats();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Sync Status'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _isSyncing ? null : _triggerSync,
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _loadStats,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            _buildStatusCard(),
            const SizedBox(height: 16),
            _buildStatsCard(),
            const SizedBox(height: 16),
            if (_lastResult != null) _buildLastSyncCard(),
            const SizedBox(height: 16),
            _buildSettingsCard(),
          ],
        ),
      ),
    );
  }

  Widget _buildStatusCard() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            Row(
              children: [
                Container(
                  width: 12,
                  height: 12,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: _isSyncing ? Colors.blue : Colors.green,
                  ),
                ),
                const SizedBox(width: 8),
                Text(
                  _isSyncing ? 'Syncing...' : 'Up to Date',
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
            if (_isSyncing) ...[
              const SizedBox(height: 16),
              const LinearProgressIndicator(),
            ],
            const SizedBox(height: 16),
            ElevatedButton.icon(
              onPressed: _isSyncing ? null : _triggerSync,
              icon: _isSyncing
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        valueColor: AlwaysStoppedAnimation(Colors.white),
                      ),
                    )
                  : const Icon(Icons.sync),
              label: Text(_isSyncing ? 'Syncing...' : 'Sync Now'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatsCard() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Local Data',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            _buildStatRow('Credentials', _stats['credentials'] ?? 0, Icons.key),
            const Divider(),
            _buildStatRow('Hosts', _stats['hosts'] ?? 0, Icons.computer),
            const Divider(),
            _buildStatRow(
              'Pending Sync',
              _stats['pending_sync'] ?? 0,
              Icons.sync,
              color: Colors.orange,
            ),
            if ((_stats['failed_sync'] ?? 0) > 0) ...[
              const Divider(),
              _buildStatRow(
                'Failed',
                _stats['failed_sync'] ?? 0,
                Icons.error,
                color: Colors.red,
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildStatRow(String label, int value, IconData icon, {Color? color}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Icon(icon, color: color),
          const SizedBox(width: 16),
          Expanded(
            child: Text(label),
          ),
          Text(
            value.toString(),
            style: TextStyle(
              fontWeight: FontWeight.bold,
              color: color,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLastSyncCard() {
    final result = _lastResult!;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Last Sync Result',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            if (result.success) ...[
              _buildResultRow('Uploaded', result.uploaded, Colors.green),
              _buildResultRow('Downloaded', result.downloaded, Colors.blue),
              if (result.conflicts > 0)
                _buildResultRow('Conflicts', result.conflicts, Colors.orange),
              if (result.errors > 0)
                _buildResultRow('Errors', result.errors, Colors.red),
            ] else ...[
              Text(
                'Error: ${result.errorMessage}',
                style: const TextStyle(color: Colors.red),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildResultRow(String label, int value, Color color) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: color,
            ),
          ),
          const SizedBox(width: 8),
          Expanded(child: Text(label)),
          Text(
            value.toString(),
            style: TextStyle(
              fontWeight: FontWeight.bold,
              color: color,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSettingsCard() {
    return Card(
      child: Column(
        children: [
          ListTile(
            leading: const Icon(Icons.schedule),
            title: const Text('Auto Sync'),
            subtitle: const Text('Every 15 minutes'),
            trailing: Switch(
              value: true,
              onChanged: (value) {
                // Toggle auto sync
              },
            ),
          ),
          const Divider(height: 1),
          ListTile(
            leading: const Icon(Icons.merge_type),
            title: const Text('Conflict Resolution'),
            subtitle: const Text('Server wins'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () {
              // Open conflict settings
            },
          ),
          const Divider(height: 1),
          ListTile(
            leading: const Icon(Icons.wifi_off),
            title: const Text('Offline Mode'),
            subtitle: const Text('Work without internet'),
            trailing: Switch(
              value: false,
              onChanged: (value) {
                // Toggle offline mode
              },
            ),
          ),
        ],
      ),
    );
  }
}
