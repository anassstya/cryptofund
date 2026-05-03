import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../blocs/register_cubit.dart';
import '../widgets/custom_text_field.dart';
import 'login_screen.dart';

class RegisterScreen extends StatefulWidget {
  const RegisterScreen({super.key});

  @override
  State<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends State<RegisterScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();

  String? _emailError;
  String? _passwordError;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  void _validateEmail(String? value) {
    setState(() {
      _emailError = RegisterValidator.validateEmail(value);
    });
  }

  void _validatePassword(String? value) {
    setState(() {
      _passwordError = RegisterValidator.validatePassword(value);
    });
  }

  void _submit(BuildContext context) {
    if (!_formKey.currentState!.validate()) return;
    context.read<RegisterCubit>().register(
      _emailController.text.trim(),
      _passwordController.text,
    );
  }

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (context) => RegisterCubit(),
      child: Builder(
        builder: (context) => Scaffold(
          backgroundColor: const Color(0xFF0A0E0A),
          resizeToAvoidBottomInset: true,
          body: SafeArea(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24.0),
              child: SingleChildScrollView(
                child: Form(
                  key: _formKey,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const SizedBox(height: 60),
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
                      const Text(
                        'Создать аккаунт',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 36,
                          fontWeight: FontWeight.w700,
                          height: 1.15,
                        ),
                      ),
                      const SizedBox(height: 12),
                      Text(
                        'Один кошелёк для всех бирж.',
                        style: TextStyle(color: Colors.grey[500], fontSize: 14),
                      ),
                      const SizedBox(height: 60),
                      CustomTextField(
                        label: 'EMAIL',
                        hint: 'text@gmail.com',
                        keyboardType: TextInputType.emailAddress,
                        controller: _emailController,
                        onChanged: _validateEmail,
                        validator: (value) => RegisterValidator.validateEmail(value),
                      ),
                      const SizedBox(height: 20),
                      CustomTextField(
                        label: 'ПАРОЛЬ',
                        hint: '••••••••',
                        obscureText: true,
                        controller: _passwordController,
                        onChanged: _validatePassword,
                        validator: (value) => RegisterValidator.validatePassword(value),
                      ),
                      BlocListener<RegisterCubit, RegisterState>(
                        listener: (context, state) {
                          if (state is RegisterError) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              SnackBar(
                                content: Text(state.message),
                                backgroundColor: Colors.red[400],
                                behavior: SnackBarBehavior.floating,
                              ),
                            );
                          } else if (state is RegisterSuccess) {
                            Navigator.pushReplacementNamed(context, '/home');
                          }
                        },
                        child: const SizedBox.shrink(),
                      ),
                      const SizedBox(height: 40),

                      BlocBuilder<RegisterCubit, RegisterState>(
                        builder: (context, state) {
                          final isLoading = state is RegisterLoading;
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
                                splashFactory: InkRipple.splashFactory,
                              ),
                              child: isLoading
                                  ? const SizedBox(
                                      height: 20,
                                      width: 20,
                                      child: CircularProgressIndicator(
                                        strokeWidth: 2,
                                        valueColor: AlwaysStoppedAnimation<Color>(
                                          Colors.black,
                                        ),
                                      ),
                                    )
                                  : const Text(
                                      'Создать аккаунт',
                                      style: TextStyle(
                                        color: Colors.black,
                                        fontSize: 16,
                                        fontWeight: FontWeight.w600,
                                      ),
                                    ),
                            ),
                          );
                        },
                      ),
                      const SizedBox(height: 0),
                      Center(
                        child: TextButton(
                          onPressed: () {
                            Navigator.push(
                              context,
                              PageRouteBuilder(
                                pageBuilder:
                                    (context, animation, secondaryAnimation) =>
                                        const LoginScreen(),
                                transitionsBuilder: (context, animation,
                                    secondaryAnimation, child) {
                                  return FadeTransition(
                                    opacity: animation,
                                    child: child,
                                  );
                                },
                                transitionDuration:
                                    const Duration(milliseconds: 300),
                              ),
                            );
                          },
                          child: Text(
                            'Уже есть аккаунт? Войти',
                            style: TextStyle(
                              color: const Color(0xFF6FCF97),
                              fontSize: 14,
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
    );
  }
}