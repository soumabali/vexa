import 'dart:async';
import 'package:connectivity_plus/connectivity_plus.dart';

class ConnectivityService {
  static final Connectivity _connectivity = Connectivity();
  static final StreamController<bool> _connectionController = 
    StreamController<bool>.broadcast();

  static Stream<bool> get connectionStream => _connectionController.stream;

  static bool _isOnline = true;
  static bool get isOnline => _isOnline;

  static Future<bool> get isOnline async {
    final result = await _connectivity.checkConnectivity();
    _isOnline = result != ConnectivityResult.none;
    return _isOnline;
  }

  /// Initialize connectivity monitoring
  static void initialize() {
    _connectivity.onConnectivityChanged.listen((result) {
      _isOnline = result != ConnectivityResult.none;
      _connectionController.add(_isOnline);
    });
  }

  /// Dispose
  static void dispose() {
    _connectionController.close();
  }
}
