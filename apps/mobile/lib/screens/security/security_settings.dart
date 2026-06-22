import 'package:flutter/material.dart';
import '../../services/biometric_service.dart';
import 'biometric_setup.dart';
import 'pin_setup.dart';

class SecuritySettingsScreen extends StatefulWidget {
  const SecuritySettingsScreen({Key? key}) : super(key: key);

  @override
  _SecuritySettingsScreenState createState() => _SecuritySettingsScreenState();
}

class _SecuritySettingsScreenState extends State<SecuritySettingsScreen> {
  bool _biometricAvailable = false;
  bool _useBiometric = false;
  bool _usePin = false;
  bool _autoLock = true;
  int _autoLockTimeout = 5; // minutes
  bool _screenshotProtection = true;
  bool _clipboardClear = true;

  @override
  void initState() {
    super.initState();
    _checkBiometric();
    _loadSettings();
  }

  Future<void> _checkBiometric() async {
    final available = await BiometricService.isAvailable;
    setState(() => _biometricAvailable = available);
  }

  Future<void> _loadSettings() async {
    // Load settings from secure storage
    setState(() {
      // _useBiometric = ...
      // _usePin = ...
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Security'),
      ),
      body: ListView(
        children: [
          // Authentication section
          _buildSectionHeader('Authentication'),
          if (_biometricAvailable)
            ListTile(
              leading: const Icon(Icons.fingerprint),
              title: const Text('Biometric'),
              subtitle: const Text('Use Face ID / Fingerprint'),
              trailing: Switch(
                value: _useBiometric,
                onChanged: (value) {
                  if (value) {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (_) => const BiometricSetupScreen(),
                      ),
                    );
                  }
                  setState(() => _useBiometric = value);
                },
              ),
              onTap: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => const BiometricSetupScreen(),
                  ),
                );
              },
            ),
          ListTile(
            leading: const Icon(Icons.pin),
            title: const Text('PIN Code'),
            subtitle: const Text('Set up a PIN for fallback'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => const PinSetupScreen(),
                ),
              );
            },
          ),
          
          // App Lock section
          _buildSectionHeader('App Lock'),
          SwitchListTile(
            title: const Text('Auto-Lock'),
            subtitle: const Text('Lock app when idle'),
            value: _autoLock,
            onChanged: (value) => setState(() => _autoLock = value),
          ),
          if (_autoLock)
            ListTile(
              leading: const Icon(Icons.timer),
              title: const Text('Auto-Lock Timeout'),
              trailing: DropdownButton<int>(
                value: _autoLockTimeout,
                onChanged: (value) {
                  if (value != null) {
                    setState(() => _autoLockTimeout = value);
                  }
                },
                items: [1, 2, 5, 10, 15, 30].map((minutes) {
                  return DropdownMenuItem(
                    value: minutes,
                    child: Text('$minutes min'),
                  );
                }).toList(),
              ),
            ),

          // Privacy section
          _buildSectionHeader('Privacy'),
          SwitchListTile(
            title: const Text('Screenshot Protection'),
            subtitle: const Text('Prevent screenshots'),
            value: _screenshotProtection,
            onChanged: (value) => setState(() => _screenshotProtection = value),
          ),
          SwitchListTile(
            title: const Text('Clear Clipboard'),
            subtitle: const Text('Auto-clear after 30 seconds'),
            value: _clipboardClear,
            onChanged: (value) => setState(() => _clipboardClear = value),
          ),

          // Advanced section
          _buildSectionHeader('Advanced'),
          ListTile(
            leading: const Icon(Icons.key),
            title: const Text('Manage Keys'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () {
              // Navigate to key management
            },
          ),
          ListTile(
            leading: const Icon(Icons.history),
            title: const Text('Security Audit Log'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () {
              // Navigate to audit log
            },
          ),
          const Divider(),
          ListTile(
            leading: const Icon(Icons.delete_forever, color: Colors.red),
            title: const Text(
              'Clear All Data',
              style: TextStyle(color: Colors.red),
            ),
            onTap: _showClearDataDialog,
          ),
        ],
      ),
    );
  }

  Widget _buildSectionHeader(String title) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 24, 16, 8),
      child: Text(
        title.toUpperCase(),
        style: TextStyle(
          fontSize: 13,
          fontWeight: FontWeight.w600,
          color: Colors.grey.shade600,
          letterSpacing: 0.5,
        ),
      ),
    );
  }

  void _showClearDataDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Clear All Data?'),
        content: const Text(
          'This will delete all local data including credentials, hosts, and settings. '
          'Data synced to the server will not be affected.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () {
              // Clear all data
              Navigator.pop(context);
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                  content: Text('All local data cleared'),
                  backgroundColor: Colors.green,
                ),
              );
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red,
            ),
            child: const Text('Clear'),
          ),
        ],
      ),
    );
  }
}
