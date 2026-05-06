class Exchange {
  final String id;
  final String name;
  final String apiKey;
  final String apiSecret;

  Exchange({
    required this.id,
    required this.name,
    required this.apiKey,
    required this.apiSecret,
  });

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'apiKey': apiKey,
        'apiSecret': apiSecret,
      };

  factory Exchange.fromJson(Map<String, dynamic> json) => Exchange(
        id: json['id'],
        name: json['name'],
        apiKey: json['apiKey'],
        apiSecret: json['apiSecret'],
      );
}