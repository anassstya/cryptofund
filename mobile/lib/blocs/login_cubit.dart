import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:http/http.dart' as http;
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'dart:convert';

abstract class LoginState extends Equatable {
  @override
  List<Object?> get props => [];
}

class LoginInitial extends LoginState {}

class LoginLoading extends LoginState {}

class LoginSuccess extends LoginState {
  final String token;
  final String userId;
  final String email;

  LoginSuccess({
    required this.token,
    required this.userId,
    required this.email,
  });

  @override
  List<Object?> get props => [token, userId, email];
}

class LoginError extends LoginState {
  final String message;

  LoginError(this.message);

  @override
  List<Object?> get props => [message];
}

class LoginCubit extends Cubit<LoginState> {
  LoginCubit() : super(LoginInitial());

  static const _baseUrl = 'http://10.0.2.2:8080';
  static const _storage = FlutterSecureStorage();

  Future<void> login(String email, String password) async {
    emit(LoginLoading());

    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/login'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'email': email,
          'password': password,
        }),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);

        final token = data['token'] as String;
        final userId = data['user_id'] as String;
        final userEmail = data['email'] as String;

        await _storage.write(key: 'token', value: token);
        await _storage.write(key: 'user_id', value: userId);
        await _storage.write(key: 'email', value: userEmail);

        emit(LoginSuccess(
          token: token,
          userId: userId,
          email: userEmail,
        ));
      } else if (response.statusCode == 401) {
        emit(LoginError('Неверный email или пароль'));
      } else if (response.statusCode == 400) {
        final data = jsonDecode(response.body);
        emit(LoginError(data['message'] ?? 'Ошибка валидации'));
      } else {
        emit(LoginError('Сервер недоступен. Попробуйте позже.'));
      }
    } catch (e) {
      emit(LoginError('Нет соединения с сервером'));
    }
  }
}