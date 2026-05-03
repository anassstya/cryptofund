import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'screens/register_screen.dart';
import 'screens/login_screen.dart';
import 'screens/home_screen.dart';
import 'blocs/register_cubit.dart';

void main() {
  runApp(const CryptoFundApp());
}

class CryptoFundApp extends StatelessWidget {
  const CryptoFundApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'CryptoFund',
      theme: ThemeData.dark().copyWith(
        scaffoldBackgroundColor: const Color(0xFF0A0E0A),
        primaryColor: const Color(0xFF6FCF97),
      ),

      home: BlocProvider(
        create: (_) => RegisterCubit(),
        child: const RegisterScreen(),
      ),

      routes: {
        '/home': (context) => const HomeScreen(),
        '/login': (context) => const LoginScreen(),
      },
      debugShowCheckedModeBanner: false,
    );
  }
}