import 'package:flutter/material.dart';
import '../models/file_item.dart';

class FileItemWidget extends StatelessWidget {
  final FileItem file;
  final bool isSelected;
  final bool isSelectionMode;
  final VoidCallback onTap;
  final VoidCallback onLongPress;

  const FileItemWidget({
    super.key,
    required this.file,
    required this.isSelected,
    required this.isSelectionMode,
    required this.onTap,
    required this.onLongPress,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return ListTile(
      leading: _buildIcon(),
      title: Text(
        file.name,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      subtitle: Text(
        _formatFileSize(file.size),
        style: TextStyle(
          fontSize: 12,
          color: colorScheme.onSurfaceVariant,
        ),
      ),
      trailing: isSelectionMode
          ? Checkbox(
              value: isSelected,
              onChanged: (_) => onTap(),
            )
          : const Icon(Icons.chevron_right),
      selected: isSelected,
      onTap: onTap,
      onLongPress: onLongPress,
    );
  }

  Widget _buildIcon() {
    if (file.isDirectory) {
      return const Icon(Icons.folder, color: Colors.amber);
    }

    final extension = file.name.split('.').last.toLowerCase();
    switch (extension) {
      case 'txt':
      case 'md':
      case 'log':
        return const Icon(Icons.description, color: Colors.blue);
      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'gif':
        return const Icon(Icons.image, color: Colors.purple);
      case 'mp3':
      case 'wav':
      case 'ogg':
        return const Icon(Icons.audio_file, color: Colors.orange);
      case 'mp4':
      case 'avi':
      case 'mkv':
        return const Icon(Icons.video_file, color: Colors.red);
      case 'zip':
      case 'tar':
      case 'gz':
        return const Icon(Icons.folder_zip, color: Colors.brown);
      case 'pdf':
        return const Icon(Icons.picture_as_pdf, color: Colors.red);
      case 'go':
      case 'js':
      case 'ts':
      case 'dart':
      case 'py':
      case 'rs':
        return const Icon(Icons.code, color: Colors.green);
      default:
        return const Icon(Icons.insert_drive_file);
    }
  }

  String _formatFileSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(1)} GB';
  }
}

class FileItemSkeleton extends StatelessWidget {
  const FileItemSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return ListTile(
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
        width: 80,
        height: 12,
        decoration: BoxDecoration(
          color: isDark ? Colors.grey[800] : Colors.grey[200],
          borderRadius: BorderRadius.circular(4),
        ),
      ),
    );
  }
}
