import 'package:flutter/material.dart';
import '../models/credential.dart';

class VaultItemWidget extends StatelessWidget {
  final Credential credential;
  final bool isSelected;
  final bool isSelectionMode;
  final VoidCallback onTap;
  final VoidCallback onLongPress;

  const VaultItemWidget({
    super.key,
    required this.credential,
    required this.isSelected,
    required this.isSelectionMode,
    required this.onTap,
    required this.onLongPress,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(
          color: isSelected
              ? colorScheme.primary
              : colorScheme.outline.withOpacity(0.2),
        ),
      ),
      child: ListTile(
        leading: _buildIcon(),
        title: Text(
          credential.name,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              '${credential.username}@${credential.host}',
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                fontSize: 12,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            if (credential.lastUsed != null)
              Text(
                'Last used: ${_formatDate(credential.lastUsed!)}',
                style: TextStyle(
                  fontSize: 11,
                  color: colorScheme.onSurfaceVariant.withOpacity(0.7),
                ),
              ),
          ],
        ),
        trailing: isSelectionMode
            ? Checkbox(
                value: isSelected,
                onChanged: (_) => onTap(),
              )
            : IconButton(
                icon: const Icon(Icons.more_vert),
                onPressed: () {
                  // Show actions
                },
              ),
        selected: isSelected,
        onTap: onTap,
        onLongPress: onLongPress,
      ),
    );
  }

  Widget _buildIcon() {
    if (credential.type == 'ssh') {
      return const Icon(Icons.terminal, color: Colors.blue);
    } else if (credential.type == 'rdp') {
      return const Icon(Icons.desktop_windows, color: Colors.purple);
    } else if (credential.type == 'vnc') {
      return const Icon(Icons.monitor, color: Colors.orange);
    }
    return const Icon(Icons.key, color: Colors.green);
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final diff = now.difference(date);

    if (diff.inDays > 365) {
      return '${(diff.inDays / 365).floor()}y ago';
    } else if (diff.inDays > 30) {
      return '${(diff.inDays / 30).floor()}mo ago';
    } else if (diff.inDays > 0) {
      return '${diff.inDays}d ago';
    } else if (diff.inHours > 0) {
      return '${diff.inHours}h ago';
    } else if (diff.inMinutes > 0) {
      return '${diff.inMinutes}m ago';
    }
    return 'Just now';
  }
}

class VaultItemSkeleton extends StatelessWidget {
  const VaultItemSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      elevation: 0,
      child: ListTile(
        leading: Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            color: isDark ? Colors.grey[800] : Colors.grey[200],
            shape: BoxShape.circle,
          ),
        ),
        title: Container(
          width: 150,
          height: 16,
          decoration: BoxDecoration(
            color: isDark ? Colors.grey[800] : Colors.grey[200],
            borderRadius: BorderRadius.circular(4),
          ),
        ),
        subtitle: Container(
          width: 200,
          height: 12,
          decoration: BoxDecoration(
            color: isDark ? Colors.grey[800] : Colors.grey[200],
            borderRadius: BorderRadius.circular(4),
          ),
        ),
      ),
    );
  }
}
