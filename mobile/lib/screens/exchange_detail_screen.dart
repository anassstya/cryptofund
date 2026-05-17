import 'dart:math' as math;
import 'package:flutter/material.dart';

class ExchangeDetailScreen extends StatefulWidget {
  final Map<String, dynamic> exchange;

  const ExchangeDetailScreen({Key? key, required this.exchange}) : super(key: key);

  @override
  State<ExchangeDetailScreen> createState() => _ExchangeDetailScreenState();
}

class _ExchangeDetailScreenState extends State<ExchangeDetailScreen> {
  double _toDouble(dynamic value) {
    if (value == null) return 0;
    if (value is int) return value.toDouble();
    if (value is double) return value;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0;
    return 0;
  }

  String _formatBalance(dynamic v) {
    final balance = _toDouble(v);
    if (balance == 0) return '\$0';
    if (balance.abs() >= 1) return '\$${balance.toStringAsFixed(2)}';
    return '\$${balance.toStringAsFixed(4).replaceFirst(RegExp(r'\.?0+$'), '')}';
  }

  String _formatPercent(dynamic v) => '${_toDouble(v) >= 0 ? '+' : ''}${_toDouble(v).toStringAsFixed(2)}%';

  String _formatAmount(dynamic v) {
    final a = _toDouble(v);
    if (a == 0) return '0';
    return a.abs() >= 1 ? a.toStringAsFixed(4).replaceFirst(RegExp(r'\.?0+$'), '') : a.toStringAsFixed(8).replaceFirst(RegExp(r'\.?0+$'), '');
  }

  Color _changeColor(dynamic v) => _toDouble(v) >= 0 ? const Color(0xFF6FCF97) : const Color(0xFFFF6B6B);

  String _shortName(String n) => n.length <= 3 ? n.toUpperCase() : n.substring(0, 3).toUpperCase();

  List<_AssetInfo> _parseAssets() {
    final rawAssets = widget.exchange['pairs'] ?? widget.exchange['assets'] ?? [];
    if (rawAssets is! List) return [];

    final assets = <_AssetInfo>[];
    for (final item in rawAssets) {
      if (item is! Map) continue;
      final symbol = (item['symbol'] ?? item['name'] ?? '').toString().toUpperCase().trim();
      if (symbol.isEmpty) continue;

      final amount = _toDouble(item['amount'] ?? item['balance']);
      final price = _toDouble(item['price_usdt'] ?? item['price']);
      var value = _toDouble(item['value_usdt'] ?? item['value']);
      if (value <= 0 && amount > 0 && price > 0) value = amount * price;
      final metric = value > 0 ? value : amount;
      if (amount <= 0 && metric <= 0) continue;

      assets.add(_AssetInfo(symbol: symbol, amount: amount, priceUSDT: price, valueUSDT: value, metric: metric));
    }

    final total = assets.fold<double>(0, (s, a) => s + a.metric);
    if (total <= 0) return assets;
    return assets.map((a) => a.copyWith(sharePercent: a.metric / total * 100)).toList()
      ..sort((a, b) => b.metric.compareTo(a.metric));
  }

  @override
  Widget build(BuildContext context) {
    final name = widget.exchange['name']?.toString() ?? 'Unknown';
    final balance = _toDouble(widget.exchange['total_balance'] ?? widget.exchange['balance']);
    final changePercent = _toDouble(widget.exchange['change_percent'] ?? 0);
    final source = widget.exchange['source']?.toString() ?? 'unknown';

    final assets = _parseAssets();
    final pad = EdgeInsets.fromLTRB(24, 20, 24, 120);

    return Scaffold(
      backgroundColor: const Color(0xFF0A0E0A),
      body: SafeArea(
        child: CustomScrollView(
          slivers: [
            SliverPadding(padding: pad.copyWith(top: 20, bottom: 0), sliver: SliverToBoxAdapter(child: _buildHeader(name))),
            SliverPadding(padding: pad.copyWith(top: 24, bottom: 0), sliver: SliverToBoxAdapter(child: _buildBalanceCard(name, balance, changePercent, source))),
            SliverPadding(padding: pad.copyWith(top: 28, bottom: 0), sliver: SliverToBoxAdapter(child: _buildSectionTitle('Состав активов'))),
            SliverPadding(padding: pad.copyWith(top: 14, bottom: 0), sliver: SliverToBoxAdapter(child: _buildAssetsWheelCard(assets))),
            SliverPadding(padding: pad, sliver: SliverToBoxAdapter(child: _buildAssetsList(assets))),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader(String name) => Row(
    children: [
      Container(
        decoration: BoxDecoration(color: const Color(0xFF141814), borderRadius: BorderRadius.circular(16), border: Border.all(color: Colors.white.withOpacity(0.06))),
        child: IconButton(icon: const Icon(Icons.arrow_back_rounded, color: Colors.white), onPressed: () => Navigator.pop(context)),
      ),
      const SizedBox(width: 14),
      Container(
        width: 54, height: 54, decoration: BoxDecoration(color: Colors.grey[900], borderRadius: BorderRadius.circular(17), border: Border.all(color: Colors.white.withOpacity(0.05))),
        child: Center(child: Text(_shortName(name), style: const TextStyle(color: Color(0xFF6FCF97), fontWeight: FontWeight.bold))),
      ),
      const SizedBox(width: 14),
      Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        const Text('БИРЖА', style: TextStyle(color: Color(0xFF6FCF97), fontSize: 12, fontWeight: FontWeight.w800, letterSpacing: 1.8)),
        const SizedBox(height: 5),
        Text(name, style: const TextStyle(color: Colors.white, fontSize: 28, fontWeight: FontWeight.w800, height: 1.1)),
      ])),
    ],
  );

  Widget _buildBalanceCard(String name, double balance, double change, String source) => Container(
    width: double.infinity, padding: const EdgeInsets.all(22),
    decoration: BoxDecoration(
      borderRadius: BorderRadius.circular(26),
      gradient: const LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: [Color(0xFF182A1E), Color(0xFF101510), Color(0xFF0D120D)]),
      border: Border.all(color: const Color(0xFF6FCF97).withOpacity(0.18)),
      boxShadow: [BoxShadow(color: const Color(0xFF6FCF97).withOpacity(0.08), blurRadius: 28, offset: const Offset(0, 14))],
    ),
    child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Text('Общий баланс на $name', style: TextStyle(color: const Color(0xFF9E9E9E), fontSize: 13, fontWeight: FontWeight.w500)),
      const SizedBox(height: 10),
      Text(_formatBalance(balance), style: const TextStyle(color: Colors.white, fontSize: 36, fontWeight: FontWeight.w800, height: 1)),
      const SizedBox(height: 18),
      Wrap(spacing: 10, runSpacing: 10, children: [
        _buildChip(text: '${_formatPercent(change)} за 1 час', color: _changeColor(change)),
        _buildChip(text: source, color: const Color(0xFF6FCF97)),
      ]),
    ]),
  );

  Widget _buildChip({required String text, required Color color}) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 9),
    decoration: BoxDecoration(
      color: color.withOpacity(0.12),
      borderRadius: BorderRadius.circular(999),
      border: Border.all(color: color.withOpacity(0.22)),
    ),
    child: Text(text, style: TextStyle(color: color, fontSize: 12, fontWeight: FontWeight.w800)),
  );

  Widget _buildSectionTitle(String t) => Text(t, style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.w800));

  Widget _buildAssetsWheelCard(List<_AssetInfo> assets) {
    if (assets.isEmpty) {
      return Container(
        padding: const EdgeInsets.all(18),
        decoration: BoxDecoration(color: const Color(0xFF141814), borderRadius: BorderRadius.circular(22)),
        child: const Center(child: Text('Нет данных об активах', style: TextStyle(color: Color(0xFF757575)))),
      );
    }
    return Container(
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(color: const Color(0xFF141814), borderRadius: BorderRadius.circular(22)),
      child: Column(
        children: [
          SizedBox(
            width: 190, height: 190,
            child: CustomPaint(
              painter: _AssetsPiePainter(assets: assets),
              child: Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Text('Активы', style: TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.w800)),
                    Text('${assets.length} монет', style: TextStyle(color: const Color(0xFF757575))),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(height: 18),
          Column(
            children: assets.take(5).map((a) {
              final c = _AssetsPiePainter.colors[assets.indexOf(a) % _AssetsPiePainter.colors.length];
              return Padding(
                padding: const EdgeInsets.only(bottom: 9),
                child: Row(
                  children: [
                    Container(width: 10, height: 10, decoration: BoxDecoration(color: c, borderRadius: BorderRadius.circular(99))),
                    const SizedBox(width: 9),
                    Expanded(child: Text(a.symbol, style: const TextStyle(color: Colors.white, fontSize: 13, fontWeight: FontWeight.w700))),
                    Text('${a.sharePercent.toStringAsFixed(1)}%', style: TextStyle(color: const Color(0xFF9E9E9E), fontSize: 13, fontWeight: FontWeight.w700)),
                  ],
                ),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildAssetsList(List<_AssetInfo> assets) {
    if (assets.isEmpty) return const SizedBox.shrink();
    return Column(
      children: assets.map((a) => Container(
        margin: const EdgeInsets.only(bottom: 12), padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(color: const Color(0xFF141814), borderRadius: BorderRadius.circular(22), border: Border.all(color: Colors.white.withOpacity(0.05))),
        child: Row(
          children: [
            Container(
              width: 44, height: 44, decoration: BoxDecoration(color: const Color(0xFF6FCF97).withOpacity(0.12), borderRadius: BorderRadius.circular(15)),
              child: Center(child: Text(a.symbol.length > 3 ? a.symbol.substring(0, 3) : a.symbol, style: const TextStyle(color: Color(0xFF6FCF97), fontSize: 13, fontWeight: FontWeight.w900))),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Text(a.symbol, style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.w800)),
                const SizedBox(height: 5),
                Text('${_formatAmount(a.amount)} ${a.symbol}', style: TextStyle(color: const Color(0xFF757575), fontSize: 12, fontWeight: FontWeight.w600)),
              ]),
            ),
            const SizedBox(width: 10),
            Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(a.valueUSDT > 0 ? _formatBalance(a.valueUSDT) : '${a.sharePercent.toStringAsFixed(1)}%', style: const TextStyle(color: Colors.white, fontSize: 15, fontWeight: FontWeight.w800)),
                const SizedBox(height: 5),
                Text(a.priceUSDT > 0 ? 'цена ${_formatBalance(a.priceUSDT)}' : 'доля ${a.sharePercent.toStringAsFixed(1)}%', style: const TextStyle(color: Color(0xFF6FCF97), fontSize: 11, fontWeight: FontWeight.w700)),
              ],
            ),
          ],
        ),
      )).toList(),
    );
  }
}

class _AssetInfo {
  final String symbol;
  final double amount;
  final double priceUSDT;
  final double valueUSDT;
  final double metric;
  final double sharePercent;
  const _AssetInfo({required this.symbol, required this.amount, required this.priceUSDT, required this.valueUSDT, required this.metric, this.sharePercent = 0});
  _AssetInfo copyWith({double? sharePercent}) => _AssetInfo(symbol: symbol, amount: amount, priceUSDT: priceUSDT, valueUSDT: valueUSDT, metric: metric, sharePercent: sharePercent ?? this.sharePercent);
}

class _AssetsPiePainter extends CustomPainter {
  final List<_AssetInfo> assets;
  static const colors = [Color(0xFF6FCF97), Color(0xFF56CCF2), Color(0xFFF2C94C), Color(0xFFBB6BD9), Color(0xFFFF6B6B), Color(0xFF2D9CDB), Color(0xFFF2994A)];
  _AssetsPiePainter({required this.assets});

  @override
  void paint(Canvas canvas, Size size) {
    if (assets.isEmpty) return;
    final total = assets.fold<double>(0, (s, a) => s + a.metric);
    if (total <= 0) return;
    final rect = Rect.fromLTWH(0, 0, size.width, size.height);
    double angle = -math.pi / 2;
    for (int i = 0; i < assets.length; i++) {
      final sweep = assets[i].metric / total * math.pi * 2;
      canvas.drawArc(rect.deflate(12), angle, sweep, false, Paint()..color = colors[i % colors.length]..style = PaintingStyle.stroke..strokeWidth = 24..strokeCap = StrokeCap.round);
      angle += sweep;
    }
  }

  @override
  bool shouldRepaint(covariant _AssetsPiePainter old) => old.assets.length != assets.length;
}