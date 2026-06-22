import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../services/biometric_service.dart';

class BiometricPrompt extends StatefulWidget {
  final String localizedReason;
  final VoidCallback? onSuccess;
  final VoidCallback? onFailure;
  final VoidCallback? onCancel;

  const BiometricPrompt({
    Key? key,
    this.localizedReason = 'Authenticate to continue',
    this.onSuccess,
    this.onFailure,
    this.onCancel,
  }) : super(key: key);

  static Future<bool> show(
    BuildContext context, {
    String localizedReason = 'Authenticate to continue',
  }) async {
    final result = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (context) => BiometricPrompt(
        localizedReason: localizedReason,
      ),
    );
    return result ?? false;
  }

  @override
  _BiometricPromptState createState() => _BiometricPromptState();
}

class _BiometricPromptState extends State<BiometricPrompt> 
    with SingleTickerProviderStateMixin {
  bool _isAuthenticating = false;
  String? _errorMessage;
  late AnimationController _animationController;

  @override
  void initState() {
    super.initState();
    _animationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat();
    _authenticate();
  }

  @override
  void dispose() {
    _animationController.dispose();
    super.dispose();
  }

  Future<void> _authenticate() async {
    setState(() {
      _isAuthenticating = true;
      _errorMessage = null;
    });

    try {
      final success = await BiometricService.authenticate(
        localizedReason: widget.localizedReason,
        useErrorDialogs: false,
        stickyAuth: true,
      );

      if (mounted) {
        if (success) {
          widget.onSuccess?.call();
          Navigator.of(context).pop(true);
        } else {
          setState(() {
            _isAuthenticating = false;
            _errorMessage = 'Authentication failed. Please try again.';
          });
          widget.onFailure?.call();
        }
      }
    } on BiometricNotEnrolledException catch (_) {
      setState(() {
        _isAuthenticating = false;
        _errorMessage = 'Biometric not set up. Please use PIN.';
      });
    } on BiometricLockedOutException catch (_) {
      setState(() {
        _isAuthenticating = false;
        _errorMessage = 'Biometric locked. Use PIN.';
      });
    } catch (e) {
      setState(() {
        _isAuthenticating = false;
        _errorMessage = 'Error: ${e.toString()}';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
      ),
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Animated icon
            AnimatedBuilder(
              animation: _animationController,
              builder: (context, child) {
                return Container(
                  width: 80,
                  height: 80,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    border: Border.all(
                      color: Colors.blue.withOpacity(
                        0.5 + 0.5 * _animationController.value,
                      ),
                      width: 2,
                    ),
                  ),
                  child: Icon(
                    Icons.fingerprint,
                    size: 40,
                    color: _isAuthenticating ? Colors.blue : Colors.grey,
                  ),
                );
              },
            ),
            const SizedBox(height: 24),
            Text(
              widget.localizedReason,
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w500,
              ),
              textAlign: TextAlign.center,
            ),
            if (_errorMessage != null) ...[
              const SizedBox(height: 16),
              Text(
                _errorMessage!,
                style: const TextStyle(
                  color: Colors.red,
                  fontSize: 14,
                ),
                textAlign: TextAlign.center,
              ),
            ],
            const SizedBox(height: 24),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              children: [
                TextButton(
                  onPressed: () {
                    widget.onCancel?.call();
                    Navigator.of(context).pop(false);
                  },
                  child: const Text('Cancel'),
                ),
                if (!_isAuthenticating) ...[
                  ElevatedButton(
                    onPressed: _authenticate,
                    child: const Text('Retry'),
                  ),
                ] else ...[
                  const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  ),
                ],
              ],
            ),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () {
                // Fallback to PIN
                Navigator.of(context).pop(false);
              },
              child: const Text('Use PIN Instead'),
            ),
          ],
        ),
      ),
    );
  }
}
