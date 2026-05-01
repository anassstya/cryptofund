import 'dart:convert';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:http/http.dart' as http;

abstract class RegisterState extends Equatable {
  const RegisterState();

  @override
  List<Object?> get props => [];
}

class RegisterInitial extends RegisterState {}

class RegisterLoading extends RegisterState {}

class RegisterSuccess extends RegisterState {
  final String token;
  final String userId;

  const RegisterSuccess({required this.token, required this.userId});

  @override
  List<Object?> get props => [token, userId];
}

class RegisterError extends RegisterState {
  final String message;

  const RegisterError(this.message);

  @override
  List<Object?> get props => [message];
}

class RegisterValidator {
  static String? validateEmail(String? value) {
    if (value == null || value.isEmpty) return 'Email обязателен';
    if (!value.contains('@') || !value.contains('.')) {
      return 'Введите корректный email';
    }
    return null;
  }

  static String? validatePassword(String? value) {
    if (value == null || value.isEmpty) return 'Пароль обязателен';
    if (value.length < 6) return 'Минимум 6 символов';
    if (value.length > 64) return 'Максимум 64 символа'; // ← добавлено
    return null;
  }
}

class RegisterCubit extends Cubit<RegisterState> {
  static const _baseUrl = 'http://10.0.2.2:8080';
  // static const _baseUrl = 'http://localhost:8080'; - браузер

  RegisterCubit() : super(RegisterInitial());

  Future<void> register(String email, String password) async {
    final emailError = RegisterValidator.validateEmail(email);
    final passwordError = RegisterValidator.validatePassword(password);

    if (emailError != null || passwordError != null) {
      emit(RegisterError(emailError ?? passwordError!));
      return;
    }

    emit(RegisterLoading());

    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/register'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'email': email, 'password': password}),
      );

      if (response.statusCode == 201) {
        final data = jsonDecode(response.body);
        emit(RegisterSuccess(
          token: data['token'] ?? '',
          userId: data['user_id'] ?? '',
        ));
      } else if (response.statusCode == 409) {
        emit(const RegisterError('Пользователь уже существует'));
      } else if (response.statusCode == 400) {
        final error = jsonDecode(response.body);
        emit(RegisterError(error['message'] ?? 'Ошибка валидации'));
      } else {
        emit(const RegisterError('Ошибка сервера'));
      }
    } catch (e) {
      emit(const RegisterError('Нет соединения с сервером'));
    }
  }

  void reset() => emit(RegisterInitial());
}