import 'dart:convert';
import 'dart:typed_data';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:encrypt/encrypt.dart' as encrypt;
import 'package:crypto/crypto.dart';

class SecureStorage {
  static const _storage = FlutterSecureStorage(
    aOptions: AndroidOptions(
      encryptedSharedPreferences: true,
      keyCipherAlgorithm: KeyCipherAlgorithm.RSA_ECB_PKCS1Padding,
      storageCipherAlgorithm: StorageCipherAlgorithm.AES_GCM_NoPadding,
    ),
    iOptions: IOSOptions(
      accessibility: KeychainAccessibility.first_unlock_this_device,
      accountName: 'ssh_manager_secure_storage',
    ),
  );

  static const _keyStorageKey = 'encryption_key_v1';
  static encrypt.Key? _cachedKey;

  /// Initialize and get encryption key
  static Future<encrypt.Key> _getKey() async {
    if (_cachedKey != null) return _cachedKey!;
    
    String? keyBase64 = await _storage.read(key: _keyStorageKey);
    
    if (keyBase64 == null) {
      // Generate new key
      final key = encrypt.Key.fromSecureRandom(32);
      keyBase64 = base64Encode(key.bytes);
      await _storage.write(key: _keyStorageKey, value: keyBase64);
    }
    
    _cachedKey = encrypt.Key.fromBase64(keyBase64);
    return _cachedKey!;
  }

  /// Encrypt data with AES-256-GCM
  static Future<String> _encrypt(String plainText) async {
    final key = await _getKey();
    final iv = encrypt.IV.fromSecureRandom(12); // GCM recommends 12 bytes
    final encrypter = encrypt.Encrypter(
      encrypt.AES(key, mode: encrypt.AESMode.gcm, padding: null),
    );
    
    final encrypted = encrypter.encrypt(plainText, iv: iv);
    
    // Store IV + ciphertext + auth tag
    final combined = Uint8List.fromList([
      ...iv.bytes,
      ...encrypted.bytes,
    ]);
    
    return base64Encode(combined);
  }

  /// Decrypt data
  static Future<String?> _decrypt(String encryptedBase64) async {
    try {
      final combined = base64Decode(encryptedBase64);
      final iv = encrypt.IV(Uint8List.fromList(combined.sublist(0, 12)));
      final ciphertext = Uint8List.fromList(combined.sublist(12));
      
      final key = await _getKey();
      final encrypter = encrypt.Encrypter(
        encrypt.AES(key, mode: encrypt.AESMode.gcm, padding: null),
      );
      
      final decrypted = encrypter.decrypt64(
        base64Encode(ciphertext),
        iv: iv,
      );
      
      return decrypted;
    } catch (e) {
      return null;
    }
  }

  /// Write encrypted value
  static Future<void> write({required String key, required String value}) async {
    final encrypted = await _encrypt(value);
    await _storage.write(key: key, value: encrypted);
  }

  /// Read and decrypt value
  static Future<String?> read({required String key}) async {
    final encrypted = await _storage.read(key: key);
    if (encrypted == null) return null;
    return _decrypt(encrypted);
  }

  /// Delete value
  static Future<void> delete({required String key}) async {
    await _storage.delete(key: key);
  }

  /// Delete all values
  static Future<void> deleteAll() async {
    await _storage.deleteAll();
    _cachedKey = null;
  }

  /// Check if key exists
  static Future<bool> contains({required String key}) async {
    return await _storage.containsKey(key: key);
  }

  /// Get all keys
  static Future<Set<String>> getAllKeys() async {
    final all = await _storage.readAll();
    return all.keys.toSet();
  }

  /// Store sensitive credential with additional key derivation
  static Future<void> writeCredential({
    required String credentialId,
    required Map<String, dynamic> data,
    String? biometricKey,
  }) async {
    String jsonData = jsonEncode(data);
    
    if (biometricKey != null) {
      // Additional key mixing with biometric-derived key
      final key = await _getKey();
      final combined = Uint8List.fromList([
        ...key.bytes,
        ...utf8.encode(biometricKey),
      ]);
      final derived = sha256.convert(combined);
      _cachedKey = encrypt.Key(Uint8List.fromList(derived.bytes));
    }
    
    await write(key: 'credential_$credentialId', value: jsonData);
  }

  /// Read credential
  static Future<Map<String, dynamic>?> readCredential({required String credentialId}) async {
    final data = await read(key: 'credential_$credentialId');
    if (data == null) return null;
    try {
      return jsonDecode(data) as Map<String, dynamic>;
    } catch (_) {
      return null;
    }
  }
}
