import 'dart:async';
import 'package:flutter/material.dart';
import '../models/connection.dart';
import '../models/file_item.dart';
import '../services/sftp_service.dart';
import '../widgets/file_item_widget.dart';
import 'connection/select_connection.dart';

class FilesScreen extends StatefulWidget {
  const FilesScreen({super.key});

  @override
  State<FilesScreen> createState() => _FilesScreenState();
}

class _FilesScreenState extends State<FilesScreen> {
  final SFTPService _service = SFTPService();
  Connection? _connection;
  List<FileItem> _files = [];
  List<FileItem> _selectedFiles = [];
  bool _isLoading = true;
  String _currentPath = '/';
  String _errorMessage = '';
  bool _isSelectionMode = false;

  @override
  void initState() {
    super.initState();
    _selectConnection();
  }

  Future<void> _selectConnection() async {
    final connection = await Navigator.push<Connection>(
      context,
      MaterialPageRoute(builder: (context) => const SelectConnectionScreen()),
    );

    if (connection != null) {
      setState(() => _connection = connection);
      _loadFiles('/');
    }
  }

  Future<void> _loadFiles(String path) async {
    if (_connection == null) return;

    setState(() {
      _isLoading = true;
      _currentPath = path;
      _errorMessage = '';
      _selectedFiles = [];
      _isSelectionMode = false;
    });

    try {
      final files = await _service.listFiles(_connection!, path);
      setState(() {
        _files = files;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        _errorMessage = e.toString();
      });
    }
  }

  void _navigateToDirectory(String path) {
    _loadFiles(path);
  }

  void _navigateUp() {
    if (_currentPath == '/') return;
    final parent = _currentPath.substring(0, _currentPath.lastIndexOf('/'));
    _loadFiles(parent.isEmpty ? '/' : parent);
  }

  void _toggleSelection(FileItem file) {
    setState(() {
      if (_selectedFiles.contains(file)) {
        _selectedFiles.remove(file);
        if (_selectedFiles.isEmpty) {
          _isSelectionMode = false;
        }
      } else {
        _selectedFiles.add(file);
        _isSelectionMode = true;
      }
    });
  }

  void _showFileActions(FileItem file) {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.open_in_new),
              title: const Text('Open'),
              onTap: () {
                Navigator.pop(context);
                _openFile(file);
              },
            ),
            ListTile(
              leading: const Icon(Icons.download),
              title: const Text('Download'),
              onTap: () {
                Navigator.pop(context);
                _downloadFile(file);
              },
            ),
            ListTile(
              leading: const Icon(Icons.share),
              title: const Text('Share'),
              onTap: () {
                Navigator.pop(context);
                _shareFile(file);
              },
            ),
            ListTile(
              leading: const Icon(Icons.edit),
              title: const Text('Rename'),
              onTap: () {
                Navigator.pop(context);
                _renameFile(file);
              },
            ),
            ListTile(
              leading: const Icon(Icons.info),
              title: const Text('Properties'),
              onTap: () {
                Navigator.pop(context);
                _showFileProperties(file);
              },
            ),
            ListTile(
              leading: const Icon(Icons.delete, color: Colors.red),
              title: const Text('Delete', style: TextStyle(color: Colors.red)),
              onTap: () {
                Navigator.pop(context);
                _deleteFile(file);
              },
            ),
          ],
        ),
      ),
    );
  }

  void _openFile(FileItem file) {
    if (file.isDirectory) {
      _navigateToDirectory(file.path);
    } else {
      // Open file viewer
      Navigator.pushNamed(context, '/file-viewer', arguments: file);
    }
  }

  Future<void> _downloadFile(FileItem file) async {
    try {
      await _service.downloadFile(_connection!, file.path, file.name);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Downloaded ${file.name}')),
      );
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Download failed: $e')),
      );
    }
  }

  void _shareFile(FileItem file) {
    // Share file
  }

  void _renameFile(FileItem file) {
    showDialog(
      context: context,
      builder: (context) {
        final controller = TextEditingController(text: file.name);
        return AlertDialog(
          title: const Text('Rename'),
          content: TextField(
            controller: controller,
            decoration: const InputDecoration(labelText: 'New name'),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Cancel'),
            ),
            TextButton(
              onPressed: () async {
                Navigator.pop(context);
                try {
                  await _service.renameFile(
                    _connection!,
                    file.path,
                    '${file.parentPath}/${controller.text}',
                  );
                  _loadFiles(_currentPath);
                } catch (e) {
                  _showError('Rename failed: $e');
                }
              },
              child: const Text('Rename'),
            ),
          ],
        );
      },
    );
  }

  void _showFileProperties(FileItem file) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(file.name),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildPropertyRow('Type', file.isDirectory ? 'Directory' : 'File'),
            _buildPropertyRow('Size', _formatFileSize(file.size)),
            _buildPropertyRow('Modified', _formatDate(file.modifiedTime)),
            _buildPropertyRow('Permissions', file.permissions),
            _buildPropertyRow('Owner', file.owner),
            _buildPropertyRow('Group', file.group),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Close'),
          ),
        ],
      ),
    );
  }

  Widget _buildPropertyRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(fontWeight: FontWeight.bold)),
          Text(value),
        ],
      ),
    );
  }

  void _deleteFile(FileItem file) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete?'),
        content: Text('Are you sure you want to delete ${file.name}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () async {
              Navigator.pop(context);
              try {
                await _service.deleteFile(_connection!, file.path);
                _loadFiles(_currentPath);
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

  void _uploadFile() {
    // Show file picker and upload
  }

  void _createDirectory() {
    showDialog(
      context: context,
      builder: (context) {
        final controller = TextEditingController();
        return AlertDialog(
          title: const Text('New Directory'),
          content: TextField(
            controller: controller,
            decoration: const InputDecoration(labelText: 'Directory name'),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Cancel'),
            ),
            TextButton(
              onPressed: () async {
                Navigator.pop(context);
                try {
                  await _service.createDirectory(
                    _connection!,
                    '$_currentPath/${controller.text}',
                  );
                  _loadFiles(_currentPath);
                } catch (e) {
                  _showError('Create directory failed: $e');
                }
              },
              child: const Text('Create'),
            ),
          ],
        );
      },
    );
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message), backgroundColor: Colors.red),
    );
  }

  String _formatFileSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(1)} GB';
  }

  String _formatDate(DateTime date) {
    return '${date.day}/${date.month}/${date.year} ${date.hour}:${date.minute}';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: _isSelectionMode
            ? Text('${_selectedFiles.length} selected')
            : Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Files'),
                  if (_connection != null)
                    Text(
                      '${_connection!.name}: $_currentPath',
                      style: const TextStyle(fontSize: 12),
                    ),
                ],
              ),
        leading: _isSelectionMode
            ? IconButton(
                icon: const Icon(Icons.close),
                onPressed: () => setState(() {
                  _isSelectionMode = false;
                  _selectedFiles = [];
                }),
              )
            : IconButton(
                icon: const Icon(Icons.arrow_back),
                onPressed: () => Navigator.pop(context),
              ),
        actions: [
          if (_isSelectionMode) ...[
            IconButton(
              icon: const Icon(Icons.download),
              onPressed: () {
                for (final file in _selectedFiles) {
                  _downloadFile(file);
                }
              },
            ),
            IconButton(
              icon: const Icon(Icons.delete),
              onPressed: () {
                for (final file in _selectedFiles) {
                  _deleteFile(file);
                }
              },
            ),
          ] else ...[
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: () => _loadFiles(_currentPath),
            ),
            IconButton(
              icon: const Icon(Icons.more_vert),
              onPressed: _showFilesMenu,
            ),
          ],
        ],
      ),
      body: _connection == null
          ? _buildNoConnection()
          : _isLoading
              ? _buildSkeleton()
              : _errorMessage.isNotEmpty
                  ? _buildError()
                  : RefreshIndicator(
                      onRefresh: () => _loadFiles(_currentPath),
                      child: ListView.builder(
                        itemCount: _files.length,
                        itemBuilder: (context, index) {
                          final file = _files[index];
                          return FileItemWidget(
                            file: file,
                            isSelected: _selectedFiles.contains(file),
                            isSelectionMode: _isSelectionMode,
                            onTap: () => _isSelectionMode
                                ? _toggleSelection(file)
                                : file.isDirectory
                                    ? _navigateToDirectory(file.path)
                                    : _showFileActions(file),
                            onLongPress: () => _toggleSelection(file),
                          );
                        },
                      ),
                    ),
      floatingActionButton: _connection != null
          ? FloatingActionButton(
              onPressed: _uploadFile,
              child: const Icon(Icons.upload_file),
            )
          : null,
      bottomNavigationBar: _connection != null
          ? BottomAppBar(
              child: Row(
                children: [
                  IconButton(
                    icon: const Icon(Icons.arrow_upward),
                    onPressed: _navigateUp,
                  ),
                  Expanded(
                    child: Text(
                      _currentPath,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(fontSize: 12),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.create_new_folder),
                    onPressed: _createDirectory,
                  ),
                ],
              ),
            )
          : null,
    );
  }

  Widget _buildNoConnection() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.computer, size: 64, color: Colors.grey[400]),
          const SizedBox(height: 16),
          const Text('No connection selected'),
          const SizedBox(height: 8),
          ElevatedButton(
            onPressed: _selectConnection,
            child: const Text('Select Connection'),
          ),
        ],
      ),
    );
  }

  Widget _buildSkeleton() {
    return ListView.builder(
      itemCount: 10,
      itemBuilder: (context, index) => const FileItemSkeleton(),
    );
  }

  Widget _buildError() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.error, size: 64, color: Colors.red),
          const SizedBox(height: 16),
          Text(_errorMessage),
          const SizedBox(height: 16),
          ElevatedButton(
            onPressed: () => _loadFiles(_currentPath),
            child: const Text('Retry'),
          ),
        ],
      ),
    );
  }

  void _showFilesMenu() {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.folder_open),
              title: const Text('Change Connection'),
              onTap: () {
                Navigator.pop(context);
                _selectConnection();
              },
            ),
            ListTile(
              leading: const Icon(Icons.create_new_folder),
              title: const Text('New Directory'),
              onTap: () {
                Navigator.pop(context);
                _createDirectory();
              },
            ),
            ListTile(
              leading: const Icon(Icons.upload_file),
              title: const Text('Upload File'),
              onTap: () {
                Navigator.pop(context);
                _uploadFile();
              },
            ),
          ],
        ),
      ),
    );
  }
}
