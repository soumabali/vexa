import 'dart:async';
import 'package:flutter/material.dart';
import '../models/credential.dart';
import '../services/vault_service.dart';
import '../widgets/vault_item_widget.dart';
import 'vault/add_credential.dart';
import 'vault/share_credential.dart';

class VaultScreen extends StatefulWidget {
  const VaultScreen({super.key});

  @override
  State<VaultScreen> createState() => _VaultScreenState();
}

class _VaultScreenState extends State<VaultScreen> {
  final VaultService _service = VaultService();
  List<Credential> _credentials = [];
  List<Credential> _filteredCredentials = [];
  bool _isLoading = true;
  String _searchQuery = '';
  bool _isSelectionMode = false;
  List<Credential> _selectedCredentials = [];
  String? _biometricError;

  @override
  void initState() {
    super.initState();
    _authenticateAndLoad();
  }

  Future<void> _authenticateAndLoad() async {
    setState(() => _isLoading = true);
    try {
      final authenticated = await _service.authenticate();
      if (authenticated) {
        await _loadCredentials();
      } else {
        setState(() {
          _biometricError = 'Authentication required';
          _isLoading = false;
        });
      }
    } catch (e) {
      setState(() {
        _biometricError = e.toString();
        _isLoading = false;
      });
    }
  }

  Future<void> _loadCredentials() async {
    try {
      final credentials = await _service.getCredentials();
      setState(() {
        _credentials = credentials;
        _filteredCredentials = credentials;
        _isLoading = false;
      });
    } catch (e) {
      setState(() => _isLoading = false);
      _showError('Failed to load credentials');
    }
  }

  void _filterCredentials(String query) {
    setState(() {
      _searchQuery = query;
      _filteredCredentials = _credentials
          .where((c) =>
              c.name.toLowerCase().contains(query.toLowerCase()) ||
              c.username.toLowerCase().contains(query.toLowerCase()) ||
              c.host.toLowerCase().contains(query.toLowerCase()))
          .toList();
    });
  }

  void _toggleSelection(Credential credential) {
    setState(() {
      if (_selectedCredentials.contains(credential)) {
        _selectedCredentials.remove(credential);
        if (_selectedCredentials.isEmpty) {
          _isSelectionMode = false;
        }
      } else {
        _selectedCredentials.add(credential);
        _isSelectionMode = true;
      }
    });
  }

  void _showCredentialActions(Credential credential) {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.visibility),
              title: const Text('View'),
              onTap: () {
                Navigator.pop(context);
                _viewCredential(credential);
              },
            ),
            ListTile(
              leading: const Icon(Icons.edit),
              title: const Text('Edit'),
              onTap: () {
                Navigator.pop(context);
                _editCredential(credential);
              },
            ),
            ListTile(
              leading: const Icon(Icons.copy),
              title: const Text('Copy Password'),
              onTap: () {
                Navigator.pop(context);
                _copyPassword(credential);
              },
            ),
            ListTile(
              leading: const Icon(Icons.share),
              title: const Text('Share'),
              onTap: () {
                Navigator.pop(context);
                _shareCredential(credential);
              },
            ),
            ListTile(
              leading: const Icon(Icons.delete, color: Colors.red),
              title: const Text('Delete', style: TextStyle(color: Colors.red)),
              onTap: () {
                Navigator.pop(context);
                _deleteCredential(credential);
              },
            ),
          ],
        ),
      ),
    );
  }

  void _viewCredential(Credential credential) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(credential.name),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildCredentialField('Host', credential.host),
            _buildCredentialField('Username', credential.username),
            _buildCredentialField('Port', credential.port.toString()),
            if (credential.notes != null)
              _buildCredentialField('Notes', credential.notes!),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Close'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              _copyPassword(credential);
            },
            child: const Text('Copy Password'),
          ),
        ],
      ),
    );
  }

  Widget _buildCredentialField(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
          const SizedBox(height: 4),
          Text(value, style: const TextStyle(fontSize: 16)),
        ],
      ),
    );
  }

  void _editCredential(Credential credential) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => AddCredentialScreen(credential: credential),
      ),
    ).then((_) => _loadCredentials());
  }

  void _copyPassword(Credential credential) async {
    try {
      final password = await _service.getPassword(credential.id);
      // Copy to clipboard
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Password copied to clipboard')),
      );
    } catch (e) {
      _showError('Failed to copy password');
    }
  }

  void _shareCredential(Credential credential) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => ShareCredentialScreen(credential: credential),
      ),
    );
  }

  void _deleteCredential(Credential credential) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Credential?'),
        content: Text('Are you sure you want to delete ${credential.name}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () async {
              Navigator.pop(context);
              try {
                await _service.deleteCredential(credential.id);
                _loadCredentials();
              } catch (e) {
                _showError('Delete failed: $e');
              }
            },
            child: const Text('Delete', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message), backgroundColor: Colors.red),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: _isSelectionMode
            ? Text('${_selectedCredentials.length} selected')
            : const Text('Credential Vault'),
        leading: _isSelectionMode
            ? IconButton(
                icon: const Icon(Icons.close),
                onPressed: () => setState(() {
                  _isSelectionMode = false;
                  _selectedCredentials = [];
                }),
              )
            : null,
        actions: [
          if (_isSelectionMode) ...[
            IconButton(
              icon: const Icon(Icons.share),
              onPressed: () {
                for (final cred in _selectedCredentials) {
                  _shareCredential(cred);
                }
              },
            ),
            IconButton(
              icon: const Icon(Icons.delete),
              onPressed: () {
                for (final cred in _selectedCredentials) {
                  _deleteCredential(cred);
                }
              },
            ),
          ] else ...[
            IconButton(
              icon: const Icon(Icons.search),
              onPressed: () {
                showSearch(
                  context: context,
                  delegate: CredentialSearchDelegate(
                    credentials: _credentials,
                    onTap: (c) => _showCredentialActions(c),
                  ),
                );
              },
            ),
            IconButton(
              icon: const Icon(Icons.more_vert),
              onPressed: _showVaultMenu,
            ),
          ],
        ],
      ),
      body: _isLoading
          ? _buildSkeleton()
          : _biometricError != null
              ? _buildBiometricError()
              : _filteredCredentials.isEmpty
                  ? _buildEmptyState()
                  : RefreshIndicator(
                      onRefresh: _loadCredentials,
                      child: ListView.builder(
                        padding: const EdgeInsets.all(16),
                        itemCount: _filteredCredentials.length,
                        itemBuilder: (context, index) {
                          final credential = _filteredCredentials[index];
                          return Dismissible(
                            key: Key(credential.id),
                            background: Container(
                              color: Colors.red,
                              alignment: Alignment.centerRight,
                              padding: const EdgeInsets.only(right: 20),
                              child: const Icon(Icons.delete, color: Colors.white),
                            ),
                            direction: DismissDirection.endToStart,
                            onDismissed: (_) => _deleteCredential(credential),
                            child: VaultItemWidget(
                              credential: credential,
                              isSelected: _selectedCredentials.contains(credential),
                              isSelectionMode: _isSelectionMode,
                              onTap: () => _isSelectionMode
                                  ? _toggleSelection(credential)
                                  : _showCredentialActions(credential),
                              onLongPress: () => _toggleSelection(credential),
                            ),
                          );
                        },
                      ),
                    ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () {
          Navigator.push(
            context,
            MaterialPageRoute(builder: (context) => const AddCredentialScreen()),
          ).then((_) => _loadCredentials());
        },
        icon: const Icon(Icons.add),
        label: const Text('Add Credential'),
      ),
    );
  }

  Widget _buildSkeleton() {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: 5,
      itemBuilder: (context, index) => const VaultItemSkeleton(),
    );
  }

  Widget _buildBiometricError() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.fingerprint, size: 64, color: Colors.grey[400]),
          const SizedBox(height: 16),
          Text(
            _biometricError!,
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 16),
          ElevatedButton.icon(
            onPressed: _authenticateAndLoad,
            icon: const Icon(Icons.fingerprint),
            label: const Text('Authenticate'),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.key, size: 64, color: Colors.grey[400]),
          const SizedBox(height: 16),
          Text(
            'No credentials yet',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 8),
          Text(
            'Store passwords securely in the vault',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Colors.grey[600],
                ),
          ),
          const SizedBox(height: 24),
          ElevatedButton.icon(
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (context) => const AddCredentialScreen(),
                ),
              ).then((_) => _loadCredentials());
            },
            icon: const Icon(Icons.add),
            label: const Text('Add Credential'),
          ),
        ],
      ),
    );
  }

  void _showVaultMenu() {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.import_export),
              title: const Text('Import'),
              onTap: () {
                Navigator.pop(context);
                // Import credentials
              },
            ),
            ListTile(
              leading: const Icon(Icons.backup),
              title: const Text('Backup'),
              onTap: () {
                Navigator.pop(context);
                // Backup vault
              },
            ),
            ListTile(
              leading: const Icon(Icons.settings),
              title: const Text('Vault Settings'),
              onTap: () {
                Navigator.pop(context);
                Navigator.pushNamed(context, '/vault-settings');
              },
            ),
          ],
        ),
      ),
    );
  }
}

class CredentialSearchDelegate extends SearchDelegate<Credential?> {
  final List<Credential> credentials;
  final Function(Credential) onTap;

  CredentialSearchDelegate({
    required this.credentials,
    required this.onTap,
  });

  @override
  List<Widget> buildActions(BuildContext context) => [
        IconButton(
          icon: const Icon(Icons.clear),
          onPressed: () => query = '',
        ),
      ];

  @override
  Widget buildLeading(BuildContext context) => IconButton(
        icon: const Icon(Icons.arrow_back),
        onPressed: () => close(context, null),
      );

  @override
  Widget buildResults(BuildContext context) => _buildList();

  @override
  Widget buildSuggestions(BuildContext context) => _buildList();

  Widget _buildList() {
    final filtered = credentials
        .where((c) =>
            c.name.toLowerCase().contains(query.toLowerCase()) ||
            c.username.toLowerCase().contains(query.toLowerCase()) ||
            c.host.toLowerCase().contains(query.toLowerCase()))
        .toList();

    return ListView.builder(
      itemCount: filtered.length,
      itemBuilder: (context, index) {
        final credential = filtered[index];
        return ListTile(
          leading: const Icon(Icons.key),
          title: Text(credential.name),
          subtitle: Text('${credential.username}@${credential.host}'),
          onTap: () {
            close(context, credential);
            onTap(credential);
          },
        );
      },
    );
  }
}
