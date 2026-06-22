import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../services/biometric_service.dart';

class BiometricSetupScreen extends StatefulWidget {
  const BiometricSetupScreen({Key? key}) : super(key: key);

  @override
  _BiometricSetupScreenState createState() => _BiometricSetupScreenState();
}

class _BiometricSetupScreenState extends State<BiometricSetupScreen> {
  bool _isLoading = true;
  bool _isAvailable = false;
  List<BiometricType> _enrolledTypes = [];
  bool _useBiometric = false;
  bool _useForUnlock = false;
  bool _useForCredentials = false;

  @override
  void initState() {
    super.initState();
    _checkBiometricStatus();
  }

  Future<void> _checkBiometricStatus() async {
    final available = await BiometricService.isAvailable;
    final types = await BiometricService.enrolledTypes;

    setState(() {
      _isAvailable = available;
      _enrolledTypes = types;
      _isLoading = false;
    });
  }

  Future<void> _testAuthentication() async {
    try {
      final success = await BiometricService.authenticate(
        localizedReason: 'Test biometric authentication',
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(success ? 'Authentication successful!' : 'Authentication failed'),
            backgroundColor: success ? Colors.green : Colors.red,
          ),
        );
      }
    } on BiometricNotEnrolledException catch (_) {
      _showSetupDialog();
    } on BiometricLockedOutException catch (_) {
      _showLockedOutDialog();
    }
  }

  void _showSetupDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Biometric Not Set Up'),
        content: const Text(
          'Please set up biometric authentication in your device settings.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () {
              // Open device settings
              Navigator.pop(context);
            },
            child: const Text('Open Settings'),
          ),
        ],
      ),
    );
  }

  void _showLockedOutDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Biometric Locked'),
        content: const Text(
          'Too many failed attempts. Please use your PIN/password.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('OK'),
          ),
        ],
      ),
    );
  }

  String _getBiometricName(BiometricType type) {
    switch (type) {
      case BiometricType.face:
        return 'Face ID';
      case BiometricType.fingerprint:
        return 'Fingerprint';
      case BiometricType.iris:
        return 'Iris';
      default:
        return 'Biometric';
    }
  }

  IconData _getBiometricIcon(BiometricType type) {
    switch (type) {
      case BiometricType.face:
        return Icons.face;
      case BiometricType.fingerprint:
        return Icons.fingerprint;
      case BiometricType.iris:
        return Icons.visibility;
      default:
        return Icons.security;
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Biometric Authentication'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (!_isAvailable) ...[
              _buildUnavailableCard(),
            ] else ...[
              _buildStatusCard(),
              const SizedBox(height: 24),
              _buildSettingsCard(),
              const SizedBox(height: 24),
              _buildTestCard(),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildUnavailableCard() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            const Icon(Icons.security, size: 48, color: Colors.grey),
            const SizedBox(height: 16),
            const Text(
              'Biometric Not Available',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            const Text(
              'This device does not support biometric authentication or it is not enrolled.',
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: () => Navigator.pushNamed(context, '/security/pin-setup'),
              child: const Text('Set Up PIN Instead'),
            ),
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
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Enrolled Biometrics',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 16),
            ..._enrolledTypes.map((type) => ListTile(
              leading: Icon(_getBiometricIcon(type)),
              title: Text(_getBiometricName(type)),
              trailing: const Icon(Icons.check_circle, color: Colors.green),
            )),
          ],
        ),
      ),
    );
  }

  Widget _buildSettingsCard() {
    return Card(
      child: Column(
        children: [
          SwitchListTile(
            title: const Text('Enable Biometric Unlock'),
            subtitle: const Text('Use biometrics to unlock the app'),
            value: _useBiometric,
            onChanged: (value) => setState(() => _useBiometric = value),
          ),
          const Divider(height: 1),
          SwitchListTile(
            title: const Text('Unlock App'),
            subtitle: const Text('Require biometric to open the app'),
            value: _useForUnlock,
            onChanged: _useBiometric
                ? (value) => setState(() => _useForUnlock = value)
                : null,
          ),
          const Divider(height: 1),
          SwitchListTile(
            title: const Text('Access Credentials'),
            subtitle: const Text('Require biometric to view credentials'),
            value: _useForCredentials,
            onChanged: _useBiometric
                ? (value) => setState(() => _useForCredentials = value)
                : null,
          ),
        ],
      ),
    );
  }

  Widget _buildTestCard() {
    return Card(
      child: ListTile(
        leading: const Icon(Icons.fingerprint, color: Colors.blue),
        title: const Text('Test Biometric'),
        subtitle: const Text('Verify your biometric works'),
        trailing: ElevatedButton(
          onPressed: _testAuthentication,
          child: const Text('Test'),
        ),
      ),
    );
  }
}
