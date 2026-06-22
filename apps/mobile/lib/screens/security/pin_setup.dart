import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class PinSetupScreen extends StatefulWidget {
  const PinSetupScreen({Key? key}) : super(key: key);

  @override
  _PinSetupScreenState createState() => _PinSetupScreenState();
}

class _PinSetupScreenState extends State<PinSetupScreen> {
  String _pin = '';
  String _confirmPin = '';
  bool _isConfirming = false;
  String? _errorMessage;
  int _pinLength = 6;

  void _onNumberPressed(String number) {
    HapticFeedback.lightImpact();
    
    if (_isConfirming) {
      if (_confirmPin.length < _pinLength) {
        setState(() {
          _confirmPin += number;
          _errorMessage = null;
        });
        
        if (_confirmPin.length == _pinLength) {
          _verifyPin();
        }
      }
    } else {
      if (_pin.length < _pinLength) {
        setState(() {
          _pin += number;
          _errorMessage = null;
        });
        
        if (_pin.length == _pinLength) {
          setState(() => _isConfirming = true);
        }
      }
    }
  }

  void _onBackspace() {
    HapticFeedback.lightImpact();
    
    if (_isConfirming && _confirmPin.isNotEmpty) {
      setState(() {
        _confirmPin = _confirmPin.substring(0, _confirmPin.length - 1);
      });
    } else if (!_isConfirming && _pin.isNotEmpty) {
      setState(() {
        _pin = _pin.substring(0, _pin.length - 1);
      });
    }
  }

  void _verifyPin() {
    if (_pin == _confirmPin) {
      // Save PIN
      _savePin();
    } else {
      HapticFeedback.heavyImpact();
      setState(() {
        _errorMessage = 'PINs do not match. Try again.';
        _confirmPin = '';
      });
    }
  }

  Future<void> _savePin() async {
    // Save PIN to secure storage
    // await SecureStorage.write(key: 'app_pin', value: _pin);
    
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('PIN set successfully'),
          backgroundColor: Colors.green,
        ),
      );
      Navigator.pop(context);
    }
  }

  void _reset() {
    setState(() {
      _pin = '';
      _confirmPin = '';
      _isConfirming = false;
      _errorMessage = null;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_isConfirming ? 'Confirm PIN' : 'Set PIN'),
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: SafeArea(
        child: Column(
          children: [
            const SizedBox(height: 40),
            // PIN dots
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: List.generate(_pinLength, (index) {
                final currentPin = _isConfirming ? _confirmPin : _pin;
                final isFilled = index < currentPin.length;
                return Container(
                  margin: const EdgeInsets.symmetric(horizontal: 8),
                  width: 16,
                  height: 16,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: isFilled ? Colors.blue : Colors.grey.shade300,
                    border: Border.all(
                      color: isFilled ? Colors.blue : Colors.grey,
                      width: 2,
                    ),
                  ),
                );
              }),
            ),
            const SizedBox(height: 24),
            // Error message
            if (_errorMessage != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 32),
                child: Text(
                  _errorMessage!,
                  style: const TextStyle(
                    color: Colors.red,
                    fontSize: 14,
                  ),
                  textAlign: TextAlign.center,
                ),
              ),
            const Spacer(),
            // Number pad
            _buildNumberPad(),
            const SizedBox(height: 16),
            // Reset button
            if (_isConfirming)
              TextButton(
                onPressed: _reset,
                child: const Text('Start Over'),
              ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _buildNumberPad() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 32),
      child: Column(
        children: [
          _buildNumberRow(['1', '2', '3']),
          const SizedBox(height: 16),
          _buildNumberRow(['4', '5', '6']),
          const SizedBox(height: 16),
          _buildNumberRow(['7', '8', '9']),
          const SizedBox(height: 16),
          _buildNumberRow(['', '0', 'backspace']),
        ],
      ),
    );
  }

  Widget _buildNumberRow(List<String> numbers) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
      children: numbers.map((number) {
        if (number.isEmpty) {
          return const SizedBox(width: 72);
        }
        
        if (number == 'backspace') {
          return SizedBox(
            width: 72,
            height: 72,
            child: IconButton(
              icon: const Icon(Icons.backspace, size: 28),
              onPressed: _onBackspace,
            ),
          );
        }

        return SizedBox(
          width: 72,
          height: 72,
          child: MaterialButton(
            shape: const CircleBorder(),
            color: Colors.grey.shade100,
            onPressed: () => _onNumberPressed(number),
            child: Text(
              number,
              style: const TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        );
      }).toList(),
    );
  }
}
