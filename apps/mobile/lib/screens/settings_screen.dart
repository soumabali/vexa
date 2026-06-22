import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../services/auth_service.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  final AuthService _authService = AuthService();
  bool _isDarkMode = false;
  bool _useSystemTheme = true;
  bool _biometricEnabled = true;
  bool _notificationsEnabled = true;
  bool _autoConnect = false;
  String _terminalFontSize = 'Medium';
  String _language = 'English';
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _loadSettings();
  }

  Future<void> _loadSettings() async {
    final prefs = await SharedPreferences.getInstance();
    setState(() {
      _isDarkMode = prefs.getBool('dark_mode') ?? false;
      _useSystemTheme = prefs.getBool('use_system_theme') ?? true;
      _biometricEnabled = prefs.getBool('biometric_enabled') ?? true;
      _notificationsEnabled = prefs.getBool('notifications_enabled') ?? true;
      _autoConnect = prefs.getBool('auto_connect') ?? false;
      _terminalFontSize = prefs.getString('terminal_font_size') ?? 'Medium';
      _language = prefs.getString('language') ?? 'English';
    });
  }

  Future<void> _saveSetting(String key, dynamic value) async {
    final prefs = await SharedPreferences.getInstance();
    if (value is bool) {
      await prefs.setBool(key, value);
    } else if (value is String) {
      await prefs.setString(key, value);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Settings'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: ListView(
        children: [
          _buildSectionHeader('Appearance'),
          _buildThemeSettings(),
          _buildSectionHeader('Security'),
          _buildSecuritySettings(),
          _buildSectionHeader('Terminal'),
          _buildTerminalSettings(),
          _buildSectionHeader('Notifications'),
          _buildNotificationSettings(),
          _buildSectionHeader('Account'),
          _buildAccountSettings(),
          _buildSectionHeader('About'),
          _buildAboutSettings(),
        ],
      ),
    );
  }

  Widget _buildSectionHeader(String title) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 24, 16, 8),
      child: Text(
        title,
        style: TextStyle(
          fontSize: 14,
          fontWeight: FontWeight.bold,
          color: Theme.of(context).colorScheme.primary,
        ),
      ),
    );
  }

  Widget _buildThemeSettings() {
    return Column(
      children: [
        SwitchListTile(
          title: const Text('Use System Theme'),
          subtitle: const Text('Follow system dark/light mode'),
          value: _useSystemTheme,
          onChanged: (value) {
            setState(() => _useSystemTheme = value);
            _saveSetting('use_system_theme', value);
          },
        ),
        if (!_useSystemTheme)
          SwitchListTile(
            title: const Text('Dark Mode'),
            subtitle: const Text('Use dark theme'),
            value: _isDarkMode,
            onChanged: (value) {
              setState(() => _isDarkMode = value);
              _saveSetting('dark_mode', value);
            },
          ),
        ListTile(
          leading: const Icon(Icons.language),
          title: const Text('Language'),
          subtitle: Text(_language),
          trailing: const Icon(Icons.chevron_right),
          onTap: _showLanguagePicker,
        ),
      ],
    );
  }

  Widget _buildSecuritySettings() {
    return Column(
      children: [
        SwitchListTile(
          title: const Text('Biometric Authentication'),
          subtitle: const Text('Require fingerprint/face to unlock vault'),
          value: _biometricEnabled,
          onChanged: (value) {
            setState(() => _biometricEnabled = value);
            _saveSetting('biometric_enabled', value);
          },
        ),
        ListTile(
          leading: const Icon(Icons.key),
          title: const Text('Change Master Password'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () {
            Navigator.pushNamed(context, '/change-password');
          },
        ),
        ListTile(
          leading: const Icon(Icons.two_mp),
          title: const Text('Two-Factor Authentication'),
          subtitle: const Text('Not enabled'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () {
            Navigator.pushNamed(context, '/2fa-setup');
          },
        ),
        ListTile(
          leading: const Icon(Icons.key),
          title: const Text('API Keys'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () {
            Navigator.pushNamed(context, '/api-keys');
          },
        ),
      ],
    );
  }

  Widget _buildTerminalSettings() {
    return Column(
      children: [
        ListTile(
          leading: const Icon(Icons.font_download),
          title: const Text('Font Size'),
          subtitle: Text(_terminalFontSize),
          trailing: const Icon(Icons.chevron_right),
          onTap: _showFontSizePicker,
        ),
        ListTile(
          leading: const Icon(Icons.terminal),
          title: const Text('Default Terminal'),
          subtitle: const Text('bash'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () {
            // Show terminal picker
          },
        ),
        SwitchListTile(
          title: const Text('Auto-Connect'),
          subtitle: const Text('Connect to last session on startup'),
          value: _autoConnect,
          onChanged: (value) {
            setState(() => _autoConnect = value);
            _saveSetting('auto_connect', value);
          },
        ),
      ],
    );
  }

  Widget _buildNotificationSettings() {
    return Column(
      children: [
        SwitchListTile(
          title: const Text('Push Notifications'),
          subtitle: const Text('Receive session alerts'),
          value: _notificationsEnabled,
          onChanged: (value) {
            setState(() => _notificationsEnabled = value);
            _saveSetting('notifications_enabled', value);
          },
        ),
        ListTile(
          leading: const Icon(Icons.vibration),
          title: const Text('Vibration'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () {
            // Show vibration settings
          },
        ),
      ],
    );
  }

  Widget _buildAccountSettings() {
    return Column(
      children: [
        ListTile(
          leading: const Icon(Icons.person),
          title: const Text('Profile'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () {
            Navigator.pushNamed(context, '/profile');
          },
        ),
        ListTile(
          leading: const Icon(Icons.devices),
          title: const Text('Connected Devices'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () {
            Navigator.pushNamed(context, '/devices');
          },
        ),
        ListTile(
          leading: const Icon(Icons.logout, color: Colors.red),
          title: const Text('Sign Out', style: TextStyle(color: Colors.red)),
          onTap: _showSignOutDialog,
        ),
      ],
    );
  }

  Widget _buildAboutSettings() {
    return Column(
      children: [
        ListTile(
          leading: const Icon(Icons.info),
          title: const Text('Version'),
          subtitle: const Text('1.0.0 (Build 100)'),
        ),
        ListTile(
          leading: const Icon(Icons.update),
          title: const Text('Check for Updates'),
          trailing: const Icon(Icons.chevron_right),
          onTap: _checkForUpdates,
        ),
        ListTile(
          leading: const Icon(Icons.description),
          title: const Text('Privacy Policy'),
          trailing: const Icon(Icons.open_in_new),
          onTap: () {
            // Open privacy policy
          },
        ),
        ListTile(
          leading: const Icon(Icons.help),
          title: const Text('Help & Support'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () {
            Navigator.pushNamed(context, '/help');
          },
        ),
      ],
    );
  }

  void _showLanguagePicker() {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              title: const Text('English'),
              trailing: _language == 'English' ? const Icon(Icons.check) : null,
              onTap: () {
                setState(() => _language = 'English');
                _saveSetting('language', 'English');
                Navigator.pop(context);
              },
            ),
            ListTile(
              title: const Text('Indonesia'),
              trailing: _language == 'Indonesia' ? const Icon(Icons.check) : null,
              onTap: () {
                setState(() => _language = 'Indonesia');
                _saveSetting('language', 'Indonesia');
                Navigator.pop(context);
              },
            ),
          ],
        ),
      ),
    );
  }

  void _showFontSizePicker() {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: ['Small', 'Medium', 'Large', 'Extra Large'].map((size) {
            return ListTile(
              title: Text(size),
              trailing: _terminalFontSize == size ? const Icon(Icons.check) : null,
              onTap: () {
                setState(() => _terminalFontSize = size);
                _saveSetting('terminal_font_size', size);
                Navigator.pop(context);
              },
            );
          }).toList(),
        ),
      ),
    );
  }

  void _showSignOutDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Sign Out?'),
        content: const Text('Are you sure you want to sign out?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () async {
              Navigator.pop(context);
              setState(() => _isLoading = true);
              await _authService.signOut();
              setState(() => _isLoading = false);
              Navigator.pushNamedAndRemoveUntil(
                context,
                '/login',
                (route) => false,
              );
            },
            child: const Text('Sign Out', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
  }

  void _checkForUpdates() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Check for Updates'),
        content: const Text('You are running the latest version.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('OK'),
          ),
        ],
      ),
    );
  }
}
