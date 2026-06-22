import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/connection.dart';
import '../services/connection_service.dart';
import '../widgets/connection_card.dart';
import '../widgets/bottom_nav.dart';
import '../widgets/drawer.dart';
import 'connection/add_connection.dart';
import 'terminal_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final ConnectionService _service = ConnectionService();
  List<Connection> _connections = [];
  List<Connection> _filteredConnections = [];
  bool _isLoading = true;
  String _searchQuery = '';
  int _selectedIndex = 0;

  @override
  void initState() {
    super.initState();
    _loadConnections();
  }

  Future<void> _loadConnections() async {
    setState(() => _isLoading = true);
    try {
      final connections = await _service.getConnections();
      setState(() {
        _connections = connections;
        _filteredConnections = connections;
        _isLoading = false;
      });
    } catch (e) {
      setState(() => _isLoading = false);
      _showError('Failed to load connections');
    }
  }

  void _filterConnections(String query) {
    setState(() {
      _searchQuery = query;
      _filteredConnections = _connections
          .where((c) =>
              c.name.toLowerCase().contains(query.toLowerCase()) ||
              c.host.toLowerCase().contains(query.toLowerCase()))
          .toList();
    });
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message), backgroundColor: Colors.red),
    );
  }

  void _onConnectionTap(Connection connection) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => TerminalScreen(connection: connection),
      ),
    );
  }

  void _onConnectionLongPress(Connection connection) {
    _showConnectionActions(connection);
  }

  void _showConnectionActions(Connection connection) {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.terminal),
              title: const Text('Connect'),
              onTap: () {
                Navigator.pop(context);
                _onConnectionTap(connection);
              },
            ),
            ListTile(
              leading: const Icon(Icons.edit),
              title: const Text('Edit'),
              onTap: () {
                Navigator.pop(context);
                _editConnection(connection);
              },
            ),
            ListTile(
              leading: const Icon(Icons.share),
              title: const Text('Share'),
              onTap: () {
                Navigator.pop(context);
                _shareConnection(connection);
              },
            ),
            ListTile(
              leading: const Icon(Icons.delete, color: Colors.red),
              title: const Text('Delete', style: TextStyle(color: Colors.red)),
              onTap: () {
                Navigator.pop(context);
                _deleteConnection(connection);
              },
            ),
          ],
        ),
      ),
    );
  }

  void _editConnection(Connection connection) {
    // Navigate to edit screen
  }

  void _shareConnection(Connection connection) {
    // Share via E2E
  }

  void _deleteConnection(Connection connection) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Connection?'),
        content: Text('Are you sure you want to delete ${connection.name}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              _service.deleteConnection(connection.id);
              _loadConnections();
            },
            child: const Text('Delete', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
  }

  void _onBottomNavTap(int index) {
    setState(() => _selectedIndex = index);
    switch (index) {
      case 0:
        // Home - already here
        break;
      case 1:
        Navigator.pushNamed(context, '/files');
        break;
      case 2:
        Navigator.pushNamed(context, '/vault');
        break;
      case 3:
        Navigator.pushNamed(context, '/settings');
        break;
    }
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      appBar: AppBar(
        title: const Text('SSH Manager'),
        centerTitle: true,
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.search),
            onPressed: () {
              showSearch(
                context: context,
                delegate: ConnectionSearchDelegate(
                  connections: _connections,
                  onTap: _onConnectionTap,
                ),
              );
            },
          ),
          IconButton(
            icon: const Icon(Icons.add),
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (context) => const AddConnectionScreen()),
              );
            },
          ),
        ],
        bottom: _searchQuery.isNotEmpty
            ? PreferredSize(
                preferredSize: const Size.fromHeight(48),
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  child: TextField(
                    onChanged: _filterConnections,
                    decoration: InputDecoration(
                      hintText: 'Search connections...',
                      prefixIcon: const Icon(Icons.search),
                      suffixIcon: _searchQuery.isNotEmpty
                          ? IconButton(
                              icon: const Icon(Icons.clear),
                              onPressed: () {
                                setState(() {
                                  _searchQuery = '';
                                  _filteredConnections = _connections;
                                });
                              },
                            )
                          : null,
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      filled: true,
                      fillColor: isDark ? Colors.grey[800] : Colors.grey[100],
                    ),
                  ),
                ),
              )
            : null,
      ),
      drawer: const AppDrawer(),
      body: RefreshIndicator(
        onRefresh: _loadConnections,
        child: _isLoading
            ? _buildSkeleton()
            : _filteredConnections.isEmpty
                ? _buildEmptyState()
                : ListView.builder(
                    padding: const EdgeInsets.all(16),
                    itemCount: _filteredConnections.length,
                    itemBuilder: (context, index) {
                      final connection = _filteredConnections[index];
                      return Dismissible(
                        key: Key(connection.id),
                        background: Container(
                          color: Colors.red,
                          alignment: Alignment.centerRight,
                          padding: const EdgeInsets.only(right: 20),
                          child: const Icon(Icons.delete, color: Colors.white),
                        ),
                        direction: DismissDirection.endToStart,
                        onDismissed: (_) => _deleteConnection(connection),
                        child: ConnectionCard(
                          connection: connection,
                          onTap: () => _onConnectionTap(connection),
                          onLongPress: () => _onConnectionLongPress(connection),
                        ),
                      );
                    },
                  ),
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () {
          Navigator.push(
            context,
            MaterialPageRoute(builder: (context) => const AddConnectionScreen()),
          );
        },
        icon: const Icon(Icons.add),
        label: const Text('New Connection'),
      ),
      bottomNavigationBar: BottomNav(
        currentIndex: _selectedIndex,
        onTap: _onBottomNavTap,
      ),
    );
  }

  Widget _buildSkeleton() {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: 5,
      itemBuilder: (context, index) => const ConnectionCardSkeleton(),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.computer, size: 64, color: Colors.grey[400]),
          const SizedBox(height: 16),
          Text(
            'No connections yet',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 8),
          Text(
            'Tap + to add your first SSH connection',
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
                  builder: (context) => const AddConnectionScreen(),
                ),
              );
            },
            icon: const Icon(Icons.add),
            label: const Text('Add Connection'),
          ),
        ],
      ),
    );
  }
}

class ConnectionSearchDelegate extends SearchDelegate<Connection?> {
  final List<Connection> connections;
  final Function(Connection) onTap;

  ConnectionSearchDelegate({
    required this.connections,
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
    final filtered = connections
        .where((c) =>
            c.name.toLowerCase().contains(query.toLowerCase()) ||
            c.host.toLowerCase().contains(query.toLowerCase()))
        .toList();

    return ListView.builder(
      itemCount: filtered.length,
      itemBuilder: (context, index) {
        final connection = filtered[index];
        return ListTile(
          leading: const Icon(Icons.terminal),
          title: Text(connection.name),
          subtitle: Text('${connection.user}@${connection.host}:${connection.port}'),
          onTap: () {
            close(context, connection);
            onTap(connection);
          },
        );
      },
    );
  }
}
