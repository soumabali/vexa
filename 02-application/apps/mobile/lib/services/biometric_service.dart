import 'package:flutter/services.dart';
import 'package:local_auth/local_auth.dart';
import 'package:local_auth/error_codes.dart' as auth_error;

class BiometricService {
  static final LocalAuthentication _localAuth = LocalAuthentication();

  /// Check if device supports biometrics
  static Future<bool> get isAvailable async {
    try {
      return await _localAuth.canCheckBiometrics;
    } on PlatformException catch (_) {
      return false;
    }
  }

  /// Get list of enrolled biometric types
  static Future<List<BiometricType>> get enrolledTypes async {
    try {
      return await _localAuth.getAvailableBiometrics();
    } on PlatformException catch (_) {
      return [];
    }
  }

  /// Authenticate with biometrics
  static Future<bool> authenticate({
    String localizedReason = 'Authenticate to access your credentials',
    bool useErrorDialogs = true,
    bool stickyAuth = false,
    bool sensitiveTransaction = true,
  }) async {
    try {
      return await _localAuth.authenticate(
        localizedReason: localizedReason,
        authMessages: const [
          AndroidAuthMessages(
            signInTitle: 'Biometric Authentication',
            cancelButton: 'Cancel',
            biometricHint: 'Verify your identity',
            biometricNotRecognized: 'Not recognized, try again',
            biometricRequiredTitle: 'Biometric authentication required',
            deviceCredentialsRequiredTitle: 'Device credentials required',
            deviceCredentialsSetupDescription: 'Please set up device credentials',
            goToSettingsButton: 'Go to Settings',
            goToSettingsDescription: 'Please set up biometric authentication in Settings',
          ),
          IOSAuthMessages(
            cancelButton: 'Cancel',
            goToSettingsButton: 'Go to Settings',
            goToSettingsDescription: 'Please set up biometric authentication in Settings',
            lockOut: 'Please reenable biometric authentication',
          ),
        ],
        options: AuthenticationOptions(
          useErrorDialogs: useErrorDialogs,
          stickyAuth: stickyAuth,
          sensitiveTransaction: sensitiveTransaction,
          biometricOnly: false,
        ),
      );
    } on PlatformException catch (e) {
      if (e.code == auth_error.notEnrolled) {
        throw BiometricNotEnrolledException();
      } else if (e.code == auth_error.lockedOut) {
        throw BiometricLockedOutException();
      } else if (e.code == auth_error.permanentlyLockedOut) {
        throw BiometricPermanentlyLockedOutException();
      }
      return false;
    }
  }

  /// Authenticate with biometrics only (no device credentials fallback)
  static Future<bool> authenticateBiometricOnly({
    String localizedReason = 'Authenticate with biometrics',
  }) async {
    try {
      return await _localAuth.authenticate(
        localizedReason: localizedReason,
        options: const AuthenticationOptions(
          useErrorDialogs: true,
          stickyAuth: false,
          sensitiveTransaction: true,
          biometricOnly: true,
        ),
      );
    } on PlatformException catch (e) {
      if (e.code == auth_error.notEnrolled) {
        throw BiometricNotEnrolledException();
      } else if (e.code == auth_error.lockedOut) {
        throw BiometricLockedOutException();
      } else if (e.code == auth_error.permanentlyLockedOut) {
        throw BiometricPermanentlyLockedOutException();
      }
      return false;
    }
  }

  /// Stop authentication
  static Future<bool> stopAuthentication() async {
    try {
      return await _localAuth.stopAuthentication();
    } on PlatformException catch (_) {
      return false;
    }
  }
}

class BiometricNotEnrolledException implements Exception {
  final String message = 'Biometric authentication is not enrolled on this device';
  @override
  String toString() => message;
}

class BiometricLockedOutException implements Exception {
  final String message = 'Biometric authentication is locked out. Try again later.';
  @override
  String toString() => message;
}

class BiometricPermanentlyLockedOutException implements Exception {
  final String message = 'Biometric authentication is permanently locked out. Please re-enroll.';
  @override
  String toString() => message;
}
