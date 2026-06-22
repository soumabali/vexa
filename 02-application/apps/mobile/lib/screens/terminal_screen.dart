import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/connection.dart';
import '../services/terminal_service.dart';
import '../widgets/terminal_view.dart';

class TerminalScreen extends StatefulWidget {
  final Connection connection;

  const TerminalScreen({super.key, required this.connection});

  @override
  State<TerminalScreen> createState() => _TerminalScreenState();
}

class _TerminalScreenState extends State<TerminalScreen>
    with TickerProviderStateMixin {
  final TerminalService _service = TerminalService();
  final TextEditingController _inputController = TextEditingController();
  final FocusNode _focusNode = FocusNode();
  final ScrollController _scrollController = ScrollController();

  bool _isConnected = false;
  bool _isConnecting = true;
  String _errorMessage = '';
  List<String> _outputLines = [];
  String _currentPrompt = '';
  StreamSubscription? _outputSubscription;

  // Terminal tabs
  List<TerminalSession> _sessions = [];
  int _activeSessionIndex = 0;

  // Terminal settings
  double _fontSize = 14;
  bool _isDark = true;
  String _fontFamily = 'JetBrainsMono';

  @override
  void initState() {
    super.initState();
    _connect();
  }

  Future<void> _connect() async {
    setState(() {
      _isConnecting = true;
      _errorMessage = '';
    });

    try {
      await _service.connect(widget.connection);
      _outputSubscription = _service.outputStream.listen(
        (data) => _onOutput(data),
        onError: (error) => _onError(error.toString()),
        onDone: () => _onDisconnected(),
      );

      setState(() {
        _isConnected = true;
        _isConnecting = false;
        _sessions.add(TerminalSession(
          id: DateTime.now().millisecondsSinceEpoch.toString(),
          name: widget.connection.name,
          output: [],
        ));
      });
    } catch (e) {
      setState(() {
        _isConnecting = false;
        _errorMessage = e.toString();
      });
    }
  }

  void _onOutput(String data) {
    setState(() {
      if (_sessions.isNotEmpty) {
        _sessions[_activeSessionIndex].output.add(data);
      }
      _outputLines.add(data);
    });
    _scrollToBottom();
  }

  void _onError(String error) {
    setState(() => _errorMessage = error);
  }

  void _onDisconnected() {
    setState(() => _isConnected = false);
  }

  void _scrollToBottom() {
    if (_scrollController.hasClients) {
      _scrollController.animateTo(
        _scrollController.position.maxScrollExtent,
        duration: const Duration(milliseconds: 100),
        curve: Curves.easeOut,
      );
    }
  }

  void _sendCommand(String command) {
    if (command.isEmpty || !_isConnected) return;
    _service.sendCommand(command);
    _inputController.clear();
    _focusNode.requestFocus();
  }

  void _sendKeyEvent(RawKeyEvent event) {
    if (!_isConnected) return;

    if (event is RawKeyDownEvent) {
      String? sequence;

      if (event.logicalKey == LogicalKeyboardKey.enter) {
        sequence = '\r';
      } else if (event.logicalKey == LogicalKeyboardKey.backspace) {
        sequence = '\b';
      } else if (event.logicalKey == LogicalKeyboardKey.delete) {
        sequence = '\x7f';
      } else if (event.logicalKey == LogicalKeyboardKey.tab) {
        sequence = '\t';
      } else if (event.logicalKey == LogicalKeyboardKey.escape) {
        sequence = '\x1b';
      } else if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
        sequence = '\x1b[A';
      } else if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
        sequence = '\x1b[B';
      } else if (event.logicalKey == LogicalKeyboardKey.arrowRight) {
        sequence = '\x1b[C';
      } else if (event.logicalKey == LogicalKeyboardKey.arrowLeft) {
        sequence = '\x1b[D';
      } else if (event.logicalKey == LogicalKeyboardKey.home) {
        sequence = '\x1b[H';
      } else if (event.logicalKey == LogicalKeyboardKey.end) {
        sequence = '\x1b[F';
      } else if (event.isControlPressed) {
        // Ctrl+A-Z
        final keyLabel = event.logicalKey.keyLabel;
        if (keyLabel.length == 1) {
          final code = keyLabel.toUpperCase().codeUnitAt(0);
          if (code >= 65 && code <= 90) {
            sequence = String.fromCharCode(code - 64);
          }
        }
      } else {
        sequence = event.character;
      }

      if (sequence != null) {
        _service.sendRawData(Uint8List.fromList(utf8.encode(sequence)));
      }
    }
  }

  void _addNewTab() {
    setState(() {
      _sessions.add(TerminalSession(
        id: DateTime.now().millisecondsSinceEpoch.toString(),
        name: 'Tab ${_sessions.length + 1}',
        output: [],
      ));
      _activeSessionIndex = _sessions.length - 1;
    });
  }

  void _closeTab(int index) {
    if (_sessions.length <= 1) return;
    setState(() {
      _sessions.removeAt(index);
      if (_activeSessionIndex >= _sessions.length) {
        _activeSessionIndex = _sessions.length - 1;
      }
    });
  }

  void _showTerminalMenu() {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.add),
              title: const Text('New Tab'),
              onTap: () {
                Navigator.pop(context);
                _addNewTab();
              },
            ),
            ListTile(
              leading: const Icon(Icons.font_download),
              title: const Text('Font Size'),
              onTap: () {
                Navigator.pop(context);
                _showFontSizeDialog();
              },
            ),
            ListTile(
              leading: const Icon(Icons.color_lens),
              title: const Text('Theme'),
              onTap: () {
                Navigator.pop(context);
                _showThemeDialog();
              },
            ),
            ListTile(
              leading: const Icon(Icons.copy),
              title: const Text('Copy'),
              onTap: () {
                Navigator.pop(context);
                _copyToClipboard();
              },
            ),
            ListTile(
              leading: const Icon(Icons.paste),
              title: const Text('Paste'),
              onTap: () {
                Navigator.pop(context);
                _pasteFromClipboard();
              },
            ),
            ListTile(
              leading: const Icon(Icons.settings),
              title: const Text('Settings'),
              onTap: () {
                Navigator.pop(context);
                Navigator.pushNamed(context, '/settings');
              },
            ),
            ListTile(
              leading: const Icon(Icons.close, color: Colors.red),
              title: const Text('Disconnect', style: TextStyle(color: Colors.red)),
              onTap: () {
                Navigator.pop(context);
                _disconnect();
              },
            ),
          ],
        ),
      ),
    );
  }

  void _showFontSizeDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Font Size'),
        content: StatefulBuilder(
          builder: (context, setDialogState) => Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Slider(
                value: _fontSize,
                min: 10,
                max: 24,
                divisions: 14,
                label: _fontSize.toStringAsFixed(0),
                onChanged: (value) {
                  setDialogState(() => _fontSize = value);
                  setState(() {});
                },
              ),
              Text('${_fontSize.toStringAsFixed(0)}pt'),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Done'),
          ),
        ],
      ),
    );
  }

  void _showThemeDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Terminal Theme'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.dark_mode),
              title: const Text('Dark'),
              trailing: _isDark ? const Icon(Icons.check) : null,
              onTap: () {
                setState(() => _isDark = true);
                Navigator.pop(context);
              },
            ),
            ListTile(
              leading: const Icon(Icons.light_mode),
              title: const Text('Light'),
              trailing: !_isDark ? const Icon(Icons.check) : null,
              onTap: () {
                setState(() => _isDark = false);
                Navigator.pop(context);
              },
            ),
          ],
        ),
      ),
    );
  }

  void _copyToClipboard() {
    // TODO: Implement copy selection
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Copied to clipboard')),
    );
  }

  void _pasteFromClipboard() async {
    final data = await Clipboard.getData(Clipboard.kTextPlain);
    if (data?.text != null) {
      _service.sendCommand(data!.text!);
    }
  }

  void _disconnect() {
    _service.disconnect();
    _outputSubscription?.cancel();
    Navigator.pop(context);
  }

  @override
  void dispose() {
    _service.disconnect();
    _outputSubscription?.cancel();
    _inputController.dispose();
    _focusNode.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(widget.connection.name),
            Text(
              '${widget.connection.user}@${widget.connection.host}',
              style: const TextStyle(fontSize: 12),
            ),
          ],
        ),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: _disconnect,
        ),
        actions: [
          if (_isConnected)
            Container(
              margin: const EdgeInsets.only(right: 8),
              width: 8,
              height: 8,
              decoration: const BoxDecoration(
                color: Colors.green,
                shape: BoxShape.circle,
              ),
            )
          else
            Container(
              margin: const EdgeInsets.only(right: 8),
              width: 8,
              height: 8,
              decoration: const BoxDecoration(
                color: Colors.red,
                shape: BoxShape.circle,
              ),
            ),
          IconButton(
            icon: const Icon(Icons.more_vert),
            onPressed: _showTerminalMenu,
          ),
        ],
        bottom: _sessions.length > 1
            ? PreferredSize(
                preferredSize: const Size.fromHeight(40),
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: _sessions.asMap().entries.map((entry) {
                      final index = entry.key;
                      final session = entry.value;
                      return Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 4),
                        child: ChoiceChip(
                          label: Text(session.name),
                          selected: index == _activeSessionIndex,
                          onSelected: (_) => setState(() => _activeSessionIndex = index),
                          deleteIcon: const Icon(Icons.close, size: 16),
                          onDeleted: _sessions.length > 1
                              ? () => _closeTab(index)
                              : null,
                        ),
                      );
                    }).toList(),
                  ),
                ),
              )
            : null,
      ),
      body: _isConnecting
          ? const Center(child: CircularProgressIndicator())
          : _errorMessage.isNotEmpty
              ? _buildErrorState()
              : _buildTerminal(),
    );
  }

  Widget _buildErrorState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.error_outline, size: 64, color: Colors.red),
          const SizedBox(height: 16),
          Text(
            'Connection Failed',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 8),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Text(
              _errorMessage,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
            ),
          ),
          const SizedBox(height: 24),
          ElevatedButton.icon(
            onPressed: _connect,
            icon: const Icon(Icons.refresh),
            label: const Text('Retry'),
          ),
        ],
      ),
    );
  }

  Widget _buildTerminal() {
    return Column(
      children: [
        Expanded(
          child: Container(
            color: _isDark ? Colors.black : Colors.white,
            child: TerminalView(
              output: _sessions.isNotEmpty
                  ? _sessions[_activeSessionIndex].output
                  : _outputLines,
              fontSize: _fontSize,
              isDark: _isDark,
              fontFamily: _fontFamily,
              scrollController: _scrollController,
            ),
          ),
        ),
        Container(
          decoration: BoxDecoration(
            color: Theme.of(context).colorScheme.surface,
            border: Border(
              top: BorderSide(
                color: Theme.of(context).dividerColor,
              ),
            ),
          ),
          child: SafeArea(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              child: Row(
                children: [
                  IconButton(
                    icon: const Icon(Icons.keyboard),
                    onPressed: () => _focusNode.requestFocus(),
                  ),
                  Expanded(
                    child: TextField(
                      controller: _inputController,
                      focusNode: _focusNode,
                      decoration: const InputDecoration(
                        hintText: 'Enter command...',
                        border: InputBorder.none,
                        contentPadding: EdgeInsets.symmetric(horizontal: 8),
                      ),
                      onSubmitted: _sendCommand,
                      textInputAction: TextInputAction.send,
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.send),
                    onPressed: () => _sendCommand(_inputController.text),
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class TerminalSession {
  final String id;
  String name;
  final List<String> output;

  TerminalSession({
    required this.id,
    required this.name,
    required this.output,
  });
}
