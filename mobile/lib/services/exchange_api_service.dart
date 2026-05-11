import 'dart:convert';

import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:http/http.dart' as http;

class ApiException implements Exception {
  final String message;

  ApiException(this.message);

  @override
  String toString() => message;
}

class ExchangeApiService {
  static const String baseUrl = 'http://10.0.2.2:8080';
  static const FlutterSecureStorage _storage = FlutterSecureStorage();

  Future<String> _getToken() async {
    final token = await _storage.read(key: 'token');

    if (token == null || token.isEmpty) {
      throw ApiException('Пользователь не авторизован');
    }

    return token;
  }

  String _extractErrorMessage(http.Response response) {
    try {
      final decoded = jsonDecode(response.body);

      if (decoded is Map<String, dynamic>) {
        final message = decoded['message']?.toString();
        final error = decoded['error']?.toString();

        if (message != null && message.isNotEmpty) {
          return message;
        }

        switch (error) {
          case 'exchange_already_exists':
            return 'Такая биржа уже добавлена';

          case 'invalid_credentials':
            return 'Неверные API ключи или нет доступа к балансу';

          case 'unsupported_exchange':
            return 'Эта биржа пока не поддерживается';

          case 'missing_required_fields':
            return 'Заполните все обязательные поля';

          case 'unauthorized':
            return 'Пользователь не авторизован';

          case 'internal_server_error':
            return 'Не удалось добавить биржу. Проверьте API ключи и права доступа';

          default:
            if (error != null && error.isNotEmpty) {
              return error;
            }
        }
      }
    } catch (_) {
      // Если body не JSON — пойдём в fallback ниже.
    }

    switch (response.statusCode) {
      case 400:
        return 'Некорректные данные';
      case 401:
        return 'Пользователь не авторизован';
      case 403:
        return 'Нет доступа. Проверьте права API ключа';
      case 404:
        return 'Запрос не найден';
      case 409:
        return 'Такая биржа уже добавлена';
      case 500:
        return 'Ошибка сервера. Проверьте API ключи и попробуйте ещё раз';
      default:
        return 'Ошибка сервера: ${response.statusCode}';
    }
  }

  Future<Map<String, dynamic>> addExchange({
    required String name,
    required String apiKey,
    required String apiSecret,
  }) async {
    final token = await _getToken();

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

    if (response.statusCode == 201) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }

    throw ApiException(_extractErrorMessage(response));
  }

  Future<List<Map<String, dynamic>>> getExchanges() async {
    final token = await _getToken();

    final response = await http.get(
      Uri.parse('$baseUrl/exchange'),
      headers: {
        'Authorization': 'Bearer $token',
      },
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body) as List<dynamic>;

      return data.map((item) => item as Map<String, dynamic>).toList();
    }

    throw ApiException(_extractErrorMessage(response));
  }
}