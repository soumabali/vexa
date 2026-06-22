# vexa Mobile (Roadmap / Experimental)

Flutter mobile application for **vexa — Complete SSH Manager**.

## Status

- **Phase:** Roadmap / experimental scaffolding
- **Milestone:** MVP planned after desktop MVP reaches feature parity with the web app
- **Current contents:** Flutter project skeleton with a placeholder main screen

## Tech Stack

- Flutter
- Dart

## Roadmap

1. Login / registration flow
2. Host list and read-only credential vault
3. SSH terminal integration
4. Push notifications for audit events
5. Android APK / iOS TestFlight distribution

## Development

```bash
# Requires Flutter SDK installed
cd apps/mobile
flutter pub get
flutter analyze
flutter run
```

## Notes

- This package is not required for the web or API builds.
- Flutter/Dart toolchain must be installed to build or test this app.
