import 'package:flutter/material.dart';
import 'add_exchange_screen.dart';
import '../models/exchange.dart';
import '../services/exchange_storage.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({Key? key}) : super(key: key);

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  List<Exchange> _exchanges = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadExchanges();
  }

 Future<void> _addExchange() async {
   final added = await Navigator.push<bool>(
     context,
     PageRouteBuilder(
       pageBuilder: (context, animation, secondaryAnimation) => const AddExchangeScreen(),
       transitionsBuilder: (context, animation, secondaryAnimation, child) =>
           FadeTransition(opacity: animation, child: child),
       transitionDuration: const Duration(milliseconds: 300),
     ),
   );

   if (mounted && added == true) {
     _loadExchanges();
   }
 }

 Future<void> _loadExchanges() async {
   setState(() => _isLoading = true);
   _exchanges = await ExchangeStorage.getExchanges();
   setState(() => _isLoading = false);
 }

 Map<String, dynamic> _getMockData(String name) {
   final hash = name.hashCode.abs();
   return {
     'balance': '\$${(hash % 50000) + 10000}.${(hash % 99).toString().padLeft(2, '0')}',
     'change': '${hash % 2 == 0 ? '+' : '-'}${(hash % 5) + 0.1}%',
     'changeColor': hash % 2 == 0 ? const Color(0xFF6FCF97) : const Color(0xFFFF6B6B),
     'assets': (hash % 20) + 1,
   };
 }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final horizontalPadding = screenWidth > 600 ? 48.0 : 24.0;

    return Scaffold(
      backgroundColor: const Color(0xFF0A0E0A),
      body: SafeArea(
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: horizontalPadding),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 20),
              const Text('CRYPTOFUND', style: TextStyle(color: Color(0xFF6FCF97), fontSize: 14, fontWeight: FontWeight.w600, letterSpacing: 2)),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Биржи — ${_exchanges.length}', style: const TextStyle(color: Colors.white, fontSize: 28, fontWeight: FontWeight.w700)),
                  IconButton(
                    icon: const Icon(Icons.logout, color: Colors.grey),
                    onPressed: () => Navigator.pushReplacementNamed(context, '/login'),
                  ),
                ],
              ),
              const SizedBox(height: 24),

              _isLoading
                  ? const Expanded(child: Center(child: CircularProgressIndicator(color: Color(0xFF6FCF97))))
                  : _exchanges.isEmpty
                      ? _buildEmptyState()
                      : _buildExchangeList(),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Expanded(
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.all(32),
              decoration: BoxDecoration(color: const Color(0xFF141814), borderRadius: BorderRadius.circular(24)),
              child: const Icon(Icons.account_balance_wallet_outlined, size: 64, color: Color(0xFF6FCF97)),
            ),
            const SizedBox(height: 24),
            const Text('Нет добавленных бирж', style: TextStyle(color: Colors.white, fontSize: 20, fontWeight: FontWeight.w600)),
            SizedBox(height: 8),
            Text('Нажмите + чтобы добавить первую биржу', style: TextStyle(color: Colors.grey[500], fontSize: 14), textAlign: TextAlign.center),
            const SizedBox(height: 32),
            GestureDetector(
              onTap: _addExchange,
              child: Container(
                width: 80, height: 80,
                decoration: const BoxDecoration(color: Color(0xFF6FCF97), shape: BoxShape.circle),
                child: const Icon(Icons.add, size: 40, color: Colors.black),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildExchangeList() {
    return Expanded(
      child: Column(
        children: [
          Expanded(
            child: ListView.builder(
              itemCount: _exchanges.length,
              padding: const EdgeInsets.only(top: 8),
              itemBuilder: (context, index) {
                final ex = _exchanges[index];
                final mock = _getMockData(ex.name);
                return Dismissible(
                  key: Key(ex.id),
                  direction: DismissDirection.endToStart,
                  background: Container(
                    alignment: Alignment.centerRight,
                    padding: const EdgeInsets.only(right: 24),
                    color: Colors.red[800],
                    child: const Icon(Icons.delete, color: Colors.white),
                  ),
                  confirmDismiss: (direction) async {
                    return await showDialog(
                      context: context,
                      builder: (ctx) => AlertDialog(
                        backgroundColor: const Color(0xFF141814),
                        title: const Text('Удалить биржу?', style: TextStyle(color: Colors.white)),
                        actions: [
                          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена', style: TextStyle(color: Colors.grey))),
                          ElevatedButton(onPressed: () => Navigator.pop(ctx, true), style: ElevatedButton.styleFrom(backgroundColor: Colors.red), child: const Text('Удалить')),
                        ],
                      ),
                    );
                  },
                  onDismissed: (direction) async {
                    await ExchangeStorage.deleteExchange(ex.id);
                    _loadExchanges();
                    if (mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text('Биржа удалена'), backgroundColor: Color(0xFF6FCF97)),
                      );
                    }
                  },
                  child: _buildExchangeCard(ex, mock),
                );
              },
            ),
          ),
          const SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            height: 56,
            child: ElevatedButton.icon(
              onPressed: _addExchange,
              icon: const Icon(Icons.add, color: Colors.black),
              label: const Text('Добавить биржу', style: TextStyle(color: Colors.black, fontSize: 16, fontWeight: FontWeight.w600)),
              style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF6FCF97), foregroundColor: Colors.black, elevation: 0, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14))),
            ),
          ),
          const SizedBox(height: 20),
        ],
      ),
    );
  }

  Widget _buildExchangeCard(Exchange ex, Map<String, dynamic> mock) {
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(color: const Color(0xFF141814), borderRadius: BorderRadius.circular(16)),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 40, height: 40,
                decoration: BoxDecoration(color: Colors.grey[800], borderRadius: BorderRadius.circular(8)),
                child: Center(child: Text(ex.name.substring(0, 3).toUpperCase(), style: const TextStyle(color: Color(0xFF6FCF97), fontWeight: FontWeight.bold))),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(ex.name, style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.w600)),
                    Text('${mock['assets']} активов', style: TextStyle(color: Colors.grey[500], fontSize: 12)),
                  ],
                ),
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Text(mock['balance'] as String, style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold)),
                  Text(mock['change'] as String, style: TextStyle(color: mock['changeColor'] as Color, fontSize: 12, fontWeight: FontWeight.w500)),
                ],
              ),
            ],
          ),
          const SizedBox(height: 12),
          const Divider(color: Colors.grey, height: 1),
          const SizedBox(height: 8),
          GestureDetector(
            onTap: () {
              // TODO: Навигация на аналитику биржи
            },
            child: Text('Открыть аналитику →', style: TextStyle(color: Colors.grey[500], fontSize: 12)),
          ),
        ],
      ),
    );
  }
}