import 'dart:typed_data';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class KeychainService {
  static const _storage = FlutterSecureStorage(
    aOptions: AndroidOptions(
      encryptedSharedPreferences: true,
    ),
    iOptions: IOSOptions(
      accessibility: KeychainAccessibility.first_unlock_this_device,
    ),
  );

  /// Store SSH private key
  static Future<void> storePrivateKey({
    required String credentialId,
    required String privateKey,
    String? passphrase,
  }) async {
    final key = 'pk_$credentialId';
    await _storage.write(key: key, value: privateKey);
    
    if (passphrase != null) {
      await _storage.write(
        key: 'passphrase_$credentialId',
        value: passphrase,
      );
    }
  }

  /// Get SSH private key
  static Future<String?> getPrivateKey(String credentialId) async {
    return await _storage.read(key: 'pk_$credentialId');
  }

  /// Get passphrase
  static Future<String?> getPassphrase(String credentialId) async {
    return await _storage.read(key: 'passphrase_$credentialId');
  }

  /// Store API key
  static Future<void> storeApiKey(String apiKey) async {
    await _storage.write(key: 'api_key', value: apiKey);
  }

  /// Get API key
  static Future<String?> getApiKey() async {
    return await _storage.read(key: 'api_key');
  }

  /// Store server URL
  static Future<void> storeServerUrl(String url) async {
    await _storage.write(key: 'server_url', value: url);
  }

  /// Get server URL
  static Future<String?> getServerUrl() async {
    return await _storage.read(key: 'server_url');
  }

  /// Store WebAuthn credential ID
  static Future<void> storeWebAuthnCredentialId(String credentialId) async {
    await _storage.write(
      key: 'webauthn_credential_id',
      value: credentialId,
    );
  }

  /// Get WebAuthn credential ID
  static Future<String?> getWebAuthnCredentialId() async {
    return await _storage.read(key: 'webauthn_credential_id');
  }

  /// Delete credential keys
  static Future<void> deleteCredentialKeys(String credentialId) async {
    await _storage.delete(key: 'pk_$credentialId');
    await _storage.delete(key: 'passphrase_$credentialId');
  }

  /// Clear all keychain data
  static Future<void> clearAll() async {
    await _storage.deleteAll();
  }
}
