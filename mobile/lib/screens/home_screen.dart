import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'add_exchange_screen.dart';
import '../services/exchange_api_service.dart';
import 'exchange_detail_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({Key? key}) : super(key: key);

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final ExchangeApiService _exchangeApiService = ExchangeApiService();

  List<Map<String, dynamic>> _exchanges = [];
  bool _isLoading = true;
  bool _isRefreshingSilently = false;

  Timer? _refreshTimer;

  @override
  void initState() {
    super.initState();

    _loadExchanges();

    _refreshTimer = Timer.periodic(
      const Duration(minutes: 3),
      (_) {
        if (mounted) {
          _refreshExchangesSilently();
        }
      },
    );
  }

  @override
  void dispose() {
    _refreshTimer?.cancel();
    super.dispose();
  }

  Future<void> _addExchange() async {
    final added = await Navigator.push<bool>(
      context,
      PageRouteBuilder(
        pageBuilder: (context, animation, secondaryAnimation) =>
            const AddExchangeScreen(),
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(opacity: animation, child: child),
        transitionDuration: const Duration(milliseconds: 300),
      ),
    );

    if (mounted && added == true) {
      await _loadExchanges();
    }
  }

  Future<void> _loadExchanges({bool showLoader = true}) async {
    if (showLoader) {
      setState(() => _isLoading = true);
    }

    try {
      final exchanges = await _exchangeApiService.getExchanges();

      if (!mounted) return;

      setState(() {
        _exchanges = exchanges;
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;

      setState(() => _isLoading = false);

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Ошибка загрузки бирж: $e'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  Future<void> _refreshExchangesSilently() async {
    if (_isRefreshingSilently || _isLoading) return;

    _isRefreshingSilently = true;

    try {
      final exchanges = await _exchangeApiService.getExchanges();

      if (!mounted) return;

      setState(() {
        _exchanges = exchanges;
      });
    } catch (_) {
    } finally {
      _isRefreshingSilently = false;
    }
  }

  double _toDouble(dynamic value) {
    if (value == null) return 0;
    if (value is int) return value.toDouble();
    if (value is double) return value;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0;
    return 0;
  }

  int _toInt(dynamic value) {
    if (value == null) return 0;
    if (value is int) return value;
    if (value is double) return value.toInt();
    if (value is num) return value.toInt();
    if (value is String) return int.tryParse(value) ?? 0;
    return 0;
  }

  String _formatAssetsCount(dynamic value) {
    final count = _toInt(value);

    final mod100 = count % 100;
    final mod10 = count % 10;

    if (mod100 >= 11 && mod100 <= 14) {
      return '$count активов';
    }

    if (mod10 == 1) {
      return '$count актив';
    }

    if (mod10 >= 2 && mod10 <= 4) {
      return '$count актива';
    }

    return '$count активов';
  }

  double _totalBalance() {
    return _exchanges.fold<double>(
      0,
      (sum, exchange) => sum + _toDouble(exchange['total_balance']),
    );
  }

  String _formatBalance(dynamic value) {
    final balance = _toDouble(value);
    return '\$${balance.toStringAsFixed(2)}';
  }

  String _formatChange(dynamic value) {
    final change = _toDouble(value);
    final sign = change >= 0 ? '+' : '';
    return '$sign${change.toStringAsFixed(2)}%';
  }

  Color _changeColor(dynamic value) {
    final change = _toDouble(value);
    return change >= 0 ? const Color(0xFF6FCF97) : const Color(0xFFFF6B6B);
  }

  String _shortName(String name) {
    if (name.length <= 3) return name.toUpperCase();
    return name.substring(0, 3).toUpperCase();
  }

  String _exchangeIconPath(String name) {
    switch (name.toLowerCase()) {
      case 'binance':
        return 'assets/binance.png';
      case 'bybit':
        return 'assets/bybit.png';
      case 'bitget':
        return 'assets/bitget.png';
      case 'gate':
        return 'assets/gate.png';
      case 'mexc':
        return 'assets/mexc.png';
      default:
        return '';
    }
  }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final horizontalPadding = screenWidth > 600 ? 48.0 : 24.0;

    return Scaffold(
      backgroundColor: const Color(0xFF0A0E0A),
      body: SafeArea(
        child: Stack(
          children: [
            RefreshIndicator(
              onRefresh: () => _loadExchanges(showLoader: false),
              color: const Color(0xFF6FCF97),
              backgroundColor: const Color(0xFF141814),
              child: CustomScrollView(
                physics: const AlwaysScrollableScrollPhysics(),
                slivers: [
                  SliverPadding(
                    padding: EdgeInsets.fromLTRB(
                      horizontalPadding,
                      20,
                      horizontalPadding,
                      0,
                    ),
                    sliver: SliverToBoxAdapter(
                      child: _buildHeader(),
                    ),
                  ),
                  SliverPadding(
                    padding: EdgeInsets.fromLTRB(
                      horizontalPadding,
                      20,
                      horizontalPadding,
                      0,
                    ),
                    sliver: SliverToBoxAdapter(
                      child: _buildPortfolioCard(),
                    ),
                  ),
                  SliverPadding(
                    padding: EdgeInsets.fromLTRB(
                      horizontalPadding,
                      28,
                      horizontalPadding,
                      12,
                    ),
                    sliver: SliverToBoxAdapter(
                      child: _buildSectionTitle(),
                    ),
                  ),
                  if (_isLoading)
                    const SliverFillRemaining(
                      hasScrollBody: false,
                      child: Center(
                        child: CircularProgressIndicator(
                          color: Color(0xFF6FCF97),
                        ),
                      ),
                    )
                  else if (_exchanges.isEmpty)
                    SliverFillRemaining(
                      hasScrollBody: false,
                      child: Padding(
                        padding: EdgeInsets.symmetric(
                          horizontal: horizontalPadding,
                        ),
                        child: _buildEmptyState(),
                      ),
                    )
                  else
                    SliverPadding(
                      padding: EdgeInsets.fromLTRB(
                        horizontalPadding,
                        0,
                        horizontalPadding,
                        120,
                      ),
                      sliver: SliverList(
                        delegate: SliverChildBuilderDelegate(
                          (context, index) {
                            final ex = _exchanges[index];
                            return _buildExchangeCard(ex);
                          },
                          childCount: _exchanges.length,
                        ),
                      ),
                    ),
                ],
              ),
            ),
            Positioned(
              left: horizontalPadding,
              right: horizontalPadding,
              bottom: 18,
              child: _buildFloatingAddButton(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        const Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'CRYPTOFUND',
              style: TextStyle(
                color: Color(0xFF6FCF97),
                fontSize: 13,
                fontWeight: FontWeight.w700,
                letterSpacing: 2.2,
              ),
            ),
            SizedBox(height: 8),
            Text(
              'Портфель',
              style: TextStyle(
                color: Colors.white,
                fontSize: 30,
                fontWeight: FontWeight.w800,
                height: 1.1,
              ),
            ),
          ],
        ),
        Container(
          decoration: BoxDecoration(
            color: const Color(0xFF141814),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Colors.white.withOpacity(0.06)),
          ),
          child: IconButton(
            icon: const Icon(Icons.logout_rounded, color: Colors.grey),
            onPressed: () => Navigator.pushReplacementNamed(context, '/login'),
          ),
        ),
      ],
    );
  }

  Widget _buildPortfolioCard() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(22),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(26),
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            Color(0xFF182A1E),
            Color(0xFF101510),
            Color(0xFF0D120D),
          ],
        ),
        border: Border.all(color: const Color(0xFF6FCF97).withOpacity(0.18)),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF6FCF97).withOpacity(0.08),
            blurRadius: 28,
            offset: const Offset(0, 14),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Общий баланс',
            style: TextStyle(
              color: Colors.grey[400],
              fontSize: 13,
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            _formatBalance(_totalBalance()),
            style: const TextStyle(
              color: Colors.white,
              fontSize: 34,
              fontWeight: FontWeight.w800,
              height: 1,
            ),
          ),
          const SizedBox(height: 18),
          Row(
            children: [
              _buildSmallStat(
                icon: Icons.account_balance_wallet_outlined,
                label: 'Бирж',
                value: '${_exchanges.length}',
              ),
              const SizedBox(width: 12),
              _buildSmallStat(
                icon: Icons.sync_rounded,
                label: 'Обновление',
                value: '3 мин',
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildSmallStat({
    required IconData icon,
    required String label,
    required String value,
  }) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
        decoration: BoxDecoration(
          color: Colors.black.withOpacity(0.22),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: Colors.white.withOpacity(0.05)),
        ),
        child: Row(
          children: [
            Icon(icon, color: const Color(0xFF6FCF97), size: 18),
            const SizedBox(width: 8),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    label,
                    style: TextStyle(
                      color: Colors.grey[500],
                      fontSize: 11,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    value,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 13,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSectionTitle() {
    return const Text(
      'Подключённые биржи',
      style: TextStyle(
        color: Colors.white,
        fontSize: 18,
        fontWeight: FontWeight.w700,
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.only(bottom: 90),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 96,
              height: 96,
              decoration: BoxDecoration(
                color: const Color(0xFF141814),
                borderRadius: BorderRadius.circular(28),
                border: Border.all(color: Colors.white.withOpacity(0.06)),
              ),
              child: const Icon(
                Icons.account_balance_wallet_outlined,
                size: 48,
                color: Color(0xFF6FCF97),
              ),
            ),
            const SizedBox(height: 22),
            const Text(
              'Нет добавленных бирж',
              style: TextStyle(
                color: Colors.white,
                fontSize: 21,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'Добавьте первую биржу, чтобы увидеть баланс и аналитику.',
              style: TextStyle(
                color: Colors.grey[500],
                fontSize: 14,
                height: 1.4,
              ),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFloatingAddButton() {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(18),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF6FCF97).withOpacity(0.28),
            blurRadius: 24,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: ElevatedButton.icon(
        onPressed: _addExchange,
        icon: const Icon(Icons.add_rounded, color: Colors.black, size: 24),
        label: const Text(
          'Добавить биржу',
          style: TextStyle(
            color: Colors.black,
            fontSize: 16,
            fontWeight: FontWeight.w800,
          ),
        ),
        style: ElevatedButton.styleFrom(
          minimumSize: const Size(double.infinity, 58),
          backgroundColor: const Color(0xFF6FCF97),
          foregroundColor: Colors.black,
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(18),
          ),
        ),
      ),
    );
  }

  Widget _buildExchangeCard(Map<String, dynamic> ex) {
    final name = ex['name']?.toString() ?? 'Unknown';
    final balance = ex['total_balance'];
    final changePercent = ex['change_percent'];
    final assetsText = _formatAssetsCount(ex['assets_count']);
    final iconPath = _exchangeIconPath(name);

    return AnimatedSwitcher(
      duration: const Duration(milliseconds: 250),
      child: GestureDetector(
        key: ValueKey('${name}_${balance}_$changePercent'),
        onTap: () {
          Navigator.push(
            context,
            PageRouteBuilder(
              pageBuilder: (context, animation, secondaryAnimation) =>
                  ExchangeDetailScreen(exchange: ex),
              transitionsBuilder: (context, animation, secondaryAnimation, child) =>
                  FadeTransition(opacity: animation, child: child),
              transitionDuration: const Duration(milliseconds: 300),
            ),
          );
        },
        child: Container(
          margin: const EdgeInsets.only(bottom: 14),
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: const Color(0xFF141814),
            borderRadius: BorderRadius.circular(22),
            border: Border.all(color: Colors.white.withOpacity(0.05)),
          ),
          child: Row(
            children: [
              WiggleExchangeIcon(
                iconPath: iconPath,
                fallbackText: _shortName(name),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      name,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 17,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const SizedBox(height: 6),
                    Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: const Color(0xFF6FCF97).withOpacity(0.12),
                            borderRadius: BorderRadius.circular(999),
                          ),
                          child: Text(
                            assetsText,
                            style: const TextStyle(
                              color: Color(0xFF6FCF97),
                              fontSize: 10,
                              fontWeight: FontWeight.w800,
                              letterSpacing: 0.4,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Text(
                    _formatBalance(balance),
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 17,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const SizedBox(height: 6),
                  Text(
                    _formatChange(changePercent),
                    style: TextStyle(
                      color: _changeColor(changePercent),
                      fontSize: 13,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class WiggleExchangeIcon extends StatefulWidget {
  final String iconPath;
  final String fallbackText;

  const WiggleExchangeIcon({
    Key? key,
    required this.iconPath,
    required this.fallbackText,
  }) : super(key: key);

  @override
  State<WiggleExchangeIcon> createState() => _WiggleExchangeIconState();
}

class _WiggleExchangeIconState extends State<WiggleExchangeIcon>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _animation;

  @override
  void initState() {
    super.initState();

    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 450),
    );

    _animation = TweenSequence<double>([
      TweenSequenceItem(tween: Tween(begin: 0, end: -0.12), weight: 1),
      TweenSequenceItem(tween: Tween(begin: -0.12, end: 0.12), weight: 2),
      TweenSequenceItem(tween: Tween(begin: 0.12, end: -0.08), weight: 2),
      TweenSequenceItem(tween: Tween(begin: -0.08, end: 0.06), weight: 2),
      TweenSequenceItem(tween: Tween(begin: 0.06, end: 0), weight: 2),
    ]).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
  }

  void _wiggle() {
    if (_controller.isAnimating) return;
    _controller.forward(from: 0);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Widget _buildFallback() {
    return Center(
      child: Text(
        widget.fallbackText,
        style: const TextStyle(
          color: Color(0xFF6FCF97),
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: _wiggle,
      child: AnimatedBuilder(
        animation: _animation,
        builder: (context, child) {
          return Transform.rotate(
            angle: _animation.value * math.pi,
            child: child,
          );
        },
        child: Container(
          width: 52,
          height: 52,
          decoration: BoxDecoration(
            color: Colors.grey[900],
            borderRadius: BorderRadius.circular(16),
          ),
          child: ClipRRect(
            borderRadius: BorderRadius.circular(16),
            child: widget.iconPath.isEmpty
                ? _buildFallback()
                : Image.asset(
                    widget.iconPath,
                    fit: BoxFit.cover,
                    errorBuilder: (context, error, stackTrace) {
                      return _buildFallback();
                    },
                  ),
          ),
        ),
      ),
    );
  }
}