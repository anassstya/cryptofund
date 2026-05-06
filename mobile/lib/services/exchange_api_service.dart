import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class ExchangeApiService {
  static const String baseUrl = 'http://10.0.2.2:8080';
  static const FlutterSecureStorage _storage = FlutterSecureStorage();

  Future<Map<String, dynamic>> addExchange({
    required String name,
    required String apiKey,
    required String apiSecret,
  }) async {
    final token = await _storage.read(key: 'token');

    print('TOKEN FROM STORAGE: $token');

    if (token == null || token.isEmpty) {
      throw Exception('Пользователь не авторизован');
    }

    final response = await http.post(
      Uri.parse('$baseUrl/exchange'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode({
        'name': name,
        'api_key': apiKey,
        'api_secret': apiSecret,
      }),
    );

    if (response.statusCode != 201) {
      throw Exception('Ошибка сервера: ${response.statusCode} ${response.body}');
    }

    return jsonDecode(response.body) as Map<String, dynamic>;
  }
}