import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'dart:convert';
import '../models/exchange.dart';

class ExchangeStorage {
  static const _storage = FlutterSecureStorage();
  static const _key = 'user_exchanges';

  static Future<List<Exchange>> getExchanges() async {
    final data = await _storage.read(key: _key);

    if (data == null || data.isEmpty) {
      return [];
    }

    final List<dynamic> jsonList = jsonDecode(data);

    return jsonList
        .map((e) => Exchange.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  static Future<void> addExchange(Exchange exchange) async {
    final current = await getExchanges();
    current.add(exchange);

    final encoded = jsonEncode(
      current.map((e) => e.toJson()).toList(),
    );

    await _storage.write(key: _key, value: encoded);
  }

  static Future<void> deleteExchange(String id) async {
    final current = await getExchanges();
    current.removeWhere((e) => e.id == id);

    final encoded = jsonEncode(
      current.map((e) => e.toJson()).toList(),
    );

    await _storage.write(key: _key, value: encoded);
  }

  static Future<void> clearExchanges() async {
    await _storage.delete(key: _key);
  }
}