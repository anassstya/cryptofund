import 'package:flutter/material.dart';
import '../services/exchange_api_service.dart';

class AddExchangeScreen extends StatefulWidget {
  const AddExchangeScreen({Key? key}) : super(key: key);

  @override
  State<AddExchangeScreen> createState() => _AddExchangeScreenState();
}

class _AddExchangeScreenState extends State<AddExchangeScreen> {
  final _formKey = GlobalKey<FormState>();
  final _apiKeyController = TextEditingController();
  final _apiSecretController = TextEditingController();
  final _exchangeApiService = ExchangeApiService();

  final LayerLink _exchangeLayerLink = LayerLink();
  final GlobalKey _exchangeFieldKey = GlobalKey();

  OverlayEntry? _exchangeOverlayEntry;

  String? _selectedExchange;
  bool _isLoading = false;
  bool _isExchangeDropdownOpen = false;

  final List<String> _exchanges = [
    'Binance',
    'Bybit',
    'Bitget',
    'Gate',
    'Mexc',
  ];

  @override
  void dispose() {
    _hideExchangeDropdown();
    _apiKeyController.dispose();
    _apiSecretController.dispose();
    super.dispose();
  }

  String? _exchangeHint() {
    switch (_selectedExchange) {
      case 'Bitget':
        return 'Bitget пока работает только с demo/mock ключами.';
      case 'Gate':
        return 'Для Gate используйте API v4 Key с правами Read-only.';
      case 'Bybit':
        return 'Для Bybit включите Read-only и Unified Trading.';
      case 'Binance':
        return 'Для Binance нужен Spot Read-only API ключ.';
      case 'Mexc':
        return 'Для MEXC нужен Spot Account Read-only API ключ.';
      default:
        return null;
    }
  }

  void _toggleExchangeDropdown() {
    FocusScope.of(context).unfocus();

    if (_isExchangeDropdownOpen) {
      _hideExchangeDropdown();
    } else {
      _showExchangeDropdown();
    }
  }

  void _showExchangeDropdown() {
    final overlay = Overlay.of(context);
    final renderBox =
        _exchangeFieldKey.currentContext?.findRenderObject() as RenderBox?;

    if (renderBox == null) return;

    final width = renderBox.size.width;

    setState(() => _isExchangeDropdownOpen = true);

    _exchangeOverlayEntry = OverlayEntry(
      builder: (context) {
        return Stack(
          children: [
            Positioned.fill(
              child: GestureDetector(
                behavior: HitTestBehavior.translucent,
                onTap: _hideExchangeDropdown,
                child: const SizedBox.expand(),
              ),
            ),
            CompositedTransformFollower(
              link: _exchangeLayerLink,
              showWhenUnlinked: false,
              offset: const Offset(0, 64),
              child: Material(
                color: Colors.transparent,
                child: SizedBox(
                  width: width,
                  child: Container(
                    decoration: BoxDecoration(
                      color: const Color(0xFF101510),
                      borderRadius: BorderRadius.circular(18),
                      border: Border.all(
                        color: const Color(0xFF6FCF97).withOpacity(0.24),
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.45),
                          blurRadius: 28,
                          offset: const Offset(0, 14),
                        ),
                      ],
                    ),
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(18),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: _exchanges
                            .map((exchange) => _buildExchangeDropdownItem(exchange))
                            .toList(),
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ],
        );
      },
    );

    overlay.insert(_exchangeOverlayEntry!);
  }

  void _hideExchangeDropdown() {
    _exchangeOverlayEntry?.remove();
    _exchangeOverlayEntry = null;

    if (mounted && _isExchangeDropdownOpen) {
      setState(() => _isExchangeDropdownOpen = false);
    }
  }

  Widget _buildExchangeDropdownItem(String exchange) {
    final isSelected = _selectedExchange == exchange;

    return InkWell(
      onTap: () {
        setState(() => _selectedExchange = exchange);
        _hideExchangeDropdown();
      },
      child: Container(
        height: 58,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        decoration: BoxDecoration(
          color: isSelected
              ? const Color(0xFF6FCF97).withOpacity(0.14)
              : Colors.transparent,
        ),
        child: Row(
          children: [
            Expanded(
              child: Text(
                exchange,
                style: TextStyle(
                  color: isSelected ? const Color(0xFF6FCF97) : Colors.white,
                  fontSize: 16,
                  fontWeight: isSelected ? FontWeight.w700 : FontWeight.w500,
                ),
              ),
            ),
            if (isSelected)
              const Icon(
                Icons.check_rounded,
                color: Color(0xFF6FCF97),
                size: 22,
              ),
          ],
        ),
      ),
    );
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate() || _selectedExchange == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Выберите биржу и заполните поля'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    _hideExchangeDropdown();

    setState(() => _isLoading = true);

    try {
      await _exchangeApiService.addExchange(
        name: _selectedExchange!,
        apiKey: _apiKeyController.text.trim(),
        apiSecret: _apiSecretController.text.trim(),
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Биржа успешно добавлена!'),
            backgroundColor: Color(0xFF6FCF97),
          ),
        );

        Navigator.pop(context, true);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(e.toString()),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final horizontalPadding = screenWidth > 600 ? 48.0 : 24.0;
    final isTablet = screenWidth > 600;
    final hint = _exchangeHint();

    return Scaffold(
      backgroundColor: const Color(0xFF0A0E0A),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.white),
          onPressed: () {
            _hideExchangeDropdown();
            Navigator.pop(context);
          },
        ),
        title: const Text(
          'Добавить биржу',
          style: TextStyle(
            color: Colors.white,
            fontSize: 20,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      body: SafeArea(
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: horizontalPadding),
          child: Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 600),
              child: SingleChildScrollView(
                keyboardDismissBehavior: ScrollViewKeyboardDismissBehavior.onDrag,
                child: Form(
                  key: _formKey,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const SizedBox(height: 24),
                      Text(
                        'Подключите API ключи',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: isTablet ? 28 : 24,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'Выберите биржу и введите API данные',
                        style: TextStyle(
                          color: Colors.grey[500],
                          fontSize: 14,
                        ),
                      ),
                      const SizedBox(height: 40),

                      const Text(
                        'БИРЖА',
                        style: TextStyle(
                          color: Color(0xFF6FCF97),
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                          letterSpacing: 1,
                        ),
                      ),
                      const SizedBox(height: 12),

                      CompositedTransformTarget(
                        link: _exchangeLayerLink,
                        child: GestureDetector(
                          key: _exchangeFieldKey,
                          onTap: _toggleExchangeDropdown,
                          child: Container(
                            height: 56,
                            padding: const EdgeInsets.symmetric(horizontal: 16),
                            decoration: BoxDecoration(
                              color: const Color(0xFF141814),
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(
                                color: _isExchangeDropdownOpen
                                    ? const Color(0xFF6FCF97)
                                    : Colors.grey[800]!,
                              ),
                            ),
                            child: Row(
                              children: [
                                Expanded(
                                  child: Text(
                                    _selectedExchange ?? 'Выберите биржу',
                                    style: TextStyle(
                                      color: _selectedExchange == null
                                          ? Colors.grey
                                          : Colors.white,
                                      fontSize: 16,
                                    ),
                                  ),
                                ),
                                AnimatedRotation(
                                  turns: _isExchangeDropdownOpen ? 0.5 : 0,
                                  duration: const Duration(milliseconds: 180),
                                  child: const Icon(
                                    Icons.keyboard_arrow_down_rounded,
                                    color: Colors.grey,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),

                      SizedBox(
                        height: 42,
                        child: Align(
                          alignment: Alignment.centerLeft,
                          child: AnimatedSwitcher(
                            duration: const Duration(milliseconds: 180),
                            child: hint == null
                                ? const SizedBox.shrink()
                                : Text(
                                    hint,
                                    key: ValueKey(hint),
                                    style: TextStyle(
                                      color: Colors.grey[500],
                                      fontSize: 12,
                                      height: 1.3,
                                    ),
                                  ),
                          ),
                        ),
                      ),

                      const Text(
                        'API KEY',
                        style: TextStyle(
                          color: Color(0xFF6FCF97),
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                          letterSpacing: 1,
                        ),
                      ),
                      const SizedBox(height: 12),
                      TextFormField(
                        controller: _apiKeyController,
                        style: const TextStyle(color: Colors.white),
                        decoration: _inputDecoration('Введите API ключ'),
                        validator: (v) => v == null || v.trim().isEmpty
                            ? 'Введите API ключ'
                            : null,
                      ),

                      const SizedBox(height: 24),
                      const Text(
                        'API SECRET',
                        style: TextStyle(
                          color: Color(0xFF6FCF97),
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                          letterSpacing: 1,
                        ),
                      ),
                      const SizedBox(height: 12),
                      TextFormField(
                        controller: _apiSecretController,
                        obscureText: true,
                        style: const TextStyle(color: Colors.white),
                        decoration: _inputDecoration('Введите секретный ключ'),
                        validator: (v) => v == null || v.trim().isEmpty
                            ? 'Введите API Secret'
                            : null,
                      ),

                      const SizedBox(height: 40),
                      SizedBox(
                        width: double.infinity,
                        height: 56,
                        child: ElevatedButton(
                          onPressed: _isLoading ? null : _submit,
                          style: ElevatedButton.styleFrom(
                            backgroundColor: const Color(0xFF6FCF97),
                            foregroundColor: Colors.black,
                            elevation: 0,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(14),
                            ),
                          ),
                          child: _isLoading
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
                                  'Добавить биржу',
                                  style: TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.w600,
                                  ),
                                ),
                        ),
                      ),

                      const SizedBox(height: 40),
                      Container(
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          color: Colors.grey[900],
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Row(
                          children: [
                            Icon(
                              Icons.info_outline,
                              color: Colors.grey[400],
                              size: 20,
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: Text(
                                'API ключи отправляются на сервер и хранятся в зашифрованном виде. Никогда не делитесь ими!',
                                style: TextStyle(
                                  color: Colors.grey[400],
                                  fontSize: 12,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 24),
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

  InputDecoration _inputDecoration(String hint) => InputDecoration(
        hintText: hint,
        hintStyle: TextStyle(color: Colors.grey[600]),
        filled: true,
        fillColor: const Color(0xFF141814),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.grey[800]!),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.grey[800]!),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFF6FCF97)),
        ),
      );
}