import 'package:flutter/material.dart';

void main() {
  runApp(const SSHManagerApp());
}

class SSHManagerApp extends StatelessWidget {
  const SSHManagerApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'SSH Manager',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      home: const Scaffold(
        body: Center(child: Text('SSH Manager Mobile - Sprint 0 Placeholder')),
      ),
    );
  }
}
