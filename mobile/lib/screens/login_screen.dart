import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../blocs/login_cubit.dart';
import '../widgets/custom_text_field.dart';
import 'register_screen.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({Key? key}) : super(key: key);

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  void _submit(BuildContext context) {
    if (!_formKey.currentState!.validate()) return;

    context.read<LoginCubit>().login(
      _emailController.text.trim(),
      _passwordController.text,
    );
  }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final textScaler = MediaQuery.textScalerOf(context);

    final horizontalPadding = screenWidth > 600 ? 48.0 : 24.0;
    final topSpacing = screenWidth > 600 ? 80.0 : 60.0;
    final sectionSpacing = screenWidth > 600 ? 72.0 : 60.0;
    final fieldSpacing = screenWidth > 600 ? 24.0 : 20.0;

    return BlocProvider(
      create: (context) => LoginCubit(),
      child: Builder(
        builder: (context) => Scaffold(
          backgroundColor: const Color(0xFF0A0E0A),
          resizeToAvoidBottomInset: true,
          body: SafeArea(
            child: Padding(
              padding: EdgeInsets.symmetric(horizontal: horizontalPadding),
              child: SingleChildScrollView(
                child: Center(
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxWidth: 450),
                    child: Form(
                      key: _formKey,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          SizedBox(height: topSpacing),

                          const Text(
                            'CRYPTOFUND',
                            style: TextStyle(
                              color: Color(0xFF6FCF97),
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              letterSpacing: 2,
                            ),
                          ),

                          const SizedBox(height: 16),

                          Text(
                            'Вход',
                            style: TextStyle(
                              color: Colors.white,
                              fontSize: textScaler.scale(36),
                              fontWeight: FontWeight.w700,
                              height: 1.15,
                            ),
                          ),

                          const SizedBox(height: 12),

                          Text(
                            'Рады видеть тебя снова.',
                            style: TextStyle(
                              color: Colors.grey[500],
                              fontSize: textScaler.scale(14),
                            ),
                          ),

                          SizedBox(height: sectionSpacing),

                          CustomTextField(
                            label: 'EMAIL',
                            hint: 'text@gmail.com',
                            keyboardType: TextInputType.emailAddress,
                            controller: _emailController,
                            validator: (value) {
                              if (value == null || value.isEmpty) return 'Введите email';
                              if (!value.contains('@')) return 'Некорректный email';
                              return null;
                            },
                          ),

                          SizedBox(height: fieldSpacing),

                          CustomTextField(
                            label: 'ПАРОЛЬ',
                            hint: '••••••••',
                            obscureText: true,
                            controller: _passwordController,
                            validator: (value) {
                              if (value == null || value.isEmpty) return 'Введите пароль';
                              if (value.length < 6) return 'Минимум 6 символов';
                              return null;
                            },
                          ),

                          Align(
                            alignment: Alignment.centerRight,
                            child: TextButton(
                              onPressed: () {

                              },
                              child: Text(
                                'Забыли пароль?',
                                style: TextStyle(
                                  color: Colors.grey[400],
                                  fontSize: textScaler.scale(13),
                                ),
                              ),
                            ),
                          ),

                          const SizedBox(height: 40),

                          BlocListener<LoginCubit, LoginState>(
                            listener: (context, state) {
                              if (state is LoginError) {
                                ScaffoldMessenger.of(context).showSnackBar(
                                  SnackBar(
                                    content: Text(state.message),
                                    backgroundColor: Colors.red[400],
                                    behavior: SnackBarBehavior.floating,
                                  ),
                                );
                              } else if (state is LoginSuccess) {
                                Navigator.pushReplacementNamed(context, '/home');
                              }
                            },
                            child: const SizedBox.shrink(),
                          ),

                          BlocBuilder<LoginCubit, LoginState>(
                            builder: (context, state) {
                              final isLoading = state is LoginLoading;
                              return SizedBox(
                                width: double.infinity,
                                height: 56,
                                child: ElevatedButton(
                                  onPressed: isLoading
                                      ? null
                                      : () => _submit(context),
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: Colors.white,
                                    foregroundColor: Colors.black,
                                    elevation: 0,
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(14),
                                    ),
                                  ),
                                  child: isLoading
                                      ? const SizedBox(
                                          height: 20,
                                          width: 20,
                                          child: CircularProgressIndicator(
                                            strokeWidth: 2,
                                            valueColor: AlwaysStoppedAnimation<Color>(Colors.black),
                                          ),
                                        )
                                      : Text(
                                          'Войти  →',
                                          style: TextStyle(
                                            color: Colors.black,
                                            fontSize: textScaler.scale(16),
                                            fontWeight: FontWeight.w600,
                                          ),
                                        ),
                                ),
                              );
                            },
                          ),

                          const SizedBox(height: 16),

                          Center(
                            child: TextButton(
                              onPressed: () {
                                Navigator.push(
                                  context,
                                  PageRouteBuilder(
                                    pageBuilder: (context, animation, secondaryAnimation) => const RegisterScreen(),
                                    transitionsBuilder: (context, animation, secondaryAnimation, child) {
                                      return FadeTransition(opacity: animation, child: child);
                                    },
                                    transitionDuration: const Duration(milliseconds: 300),
                                  ),
                                );
                              },
                              child: Text(
                                'Нет аккаунта? Зарегистрироваться',
                                style: TextStyle(
                                  color: const Color(0xFF6FCF97),
                                  fontSize: textScaler.scale(14),
                                ),
                              ),
                            ),
                          ),

                          const SizedBox(height: 40),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}