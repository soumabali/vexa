# Sprint 6 — Mobile Client

> **Phase:** 5 | **Weeks:** 17–19 | **Duration:** 15 working days
> **Sprint Goal:** Build Flutter mobile app with terminal, host management, and biometric security
> **Story Points:** 55
> **Status:** Not Started
> **Depends On:** Sprint 5 (Desktop Clients)

---

## User Stories

### US-033: Flutter App Scaffold
**As a** user, **I want** a native mobile app, **so that** I can manage servers on the go.

**Acceptance Criteria:**
- [ ] Flutter project setup with proper structure
- [ ] Material Design 3 theming
- [ ] Navigation (bottom bar + drawer)
- [ ] Splash screen and app icon
- [ ] Onboarding flow (first launch)
- [ ] Responsive layouts (phone + tablet)

**Definition of Done:**
- [ ] App launches in < 2s on mid-range phone
- [ ] 60 FPS scrolling
- [ ] Works on Android 10+ and iOS 15+
- [ ] App size < 50MB

**Story Points:** 13

---

### US-034: Mobile Host Management
**As a** user, **I want** to manage hosts on mobile, **so that** I can connect quickly from my phone.

**Acceptance Criteria:**
- [ ] Host list with swipe actions
- [ ] Host creation (simplified form)
- [ ] Quick connect (tap to connect)
- [ ] Favorites (pin to top)
- [ ] Search and filter
- [ ] Host groups (expandable)

**Definition of Done:**
- [ ] List scrolls smoothly with 500 hosts
- [ ] Quick connect in < 3 taps
- [ ] Swipe actions work reliably
- [ ] Search filters in real-time

**Story Points:** 13

---

### US-035: Mobile Terminal
**As a** user, **I want** a mobile terminal, **so that** I can execute commands on my phone.

**Acceptance Criteria:**
- [ ] Touch-optimized terminal (tap to position cursor)
- [ ] Virtual keyboard integration
- [ ] Touch gestures (pinch to zoom, swipe for history)
- [ ] Terminal colors and font size
- [ ] Session persistence ( survive app background)
- [ ] Copy/paste support

**Definition of Done:**
- [ ] Terminal usable with thumb typing
- [ ] Cursor positioning accurate
- [ ] Sessions survive app background/foreground
- [ ] Copy/paste works with system clipboard

**Story Points:** 13

---

### US-036: Biometric Security
**As a** security-conscious user, **I want** biometric authentication, **so that** my credentials are protected.

**Acceptance Criteria:**
- [ ] Fingerprint unlock (Android/iOS)
- [ ] Face ID unlock (iOS)
- [ ] PIN fallback
- [ ] Auto-lock after app background
- [ ] Vault unlock with biometrics
- [ ] Credential access requires biometric

**Definition of Done:**
- [ ] Biometric prompt on app launch
- [ ] Fallback PIN works when biometric fails
- [ ] Auto-lock after 30s in background
- [ ] Credentials inaccessible without biometric

**Story Points:** 8

---

### US-037: Mobile Notifications
**As a** user, **I want** push notifications, **so that** I know when connections complete or fail.

**Acceptance Criteria:**
- [ ] Connection status notifications
- [ ] File transfer complete
- [ ] Session timeout warning
- [ ] Security alerts (unusual login)
- [ ] Configurable notification preferences
- [ ] Silent hours (Do Not Disturb)

**Definition of Done:**
- [ ] Notifications delivered within 5s
- [ ] Tap notification opens relevant screen
- [ ] Silent hours respected
- [ ] Battery impact minimal (< 1% per day)

**Story Points:** 8

---

## Dependencies

| Dependency | Source | Impact |
|------------|--------|--------|
| Sprint 5 completion | Previous sprint | Blocks mobile development |
| Flutter SDK | Framework | Blocks all mobile work |
| Biometric libraries | `local_auth` | Blocks biometric features |
| Push notification service | Firebase/APNs | Blocks notifications |
| Rust FFI bindings | `flutter_rust_bridge` | Blocks native SSH performance |

---

## Risks and Mitigations

| Risk | Probability | Impact | Mitigation | Owner |
|------|-------------|--------|------------|-------|
| Flutter performance | Medium | High | Profile early, optimize list rendering | Mobile |
| Biometric hardware差异 | Medium | Medium | Graceful degradation to PIN | Mobile |
| App store approval | Medium | High | Follow guidelines, prepare rejection plan | PM |
| Rust FFI complexity | Medium | High | Start with HTTP API, add FFI later | Mobile |

---

## Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Framework | Flutter 3.x | Single codebase, native performance |
| State management | Riverpod | Type-safe, testable |
| Terminal | `xterm` via WebView (v1) | Fastest implementation |
| Biometric | `local_auth` | Official Flutter plugin |
| Storage | `hive` (NoSQL) | Fast, encrypted |
| HTTP client | `dio` | Interceptors, retry logic |

---

## Sprint Review Checklist

- [ ] App runs on Android and iOS devices
- [ ] Host management tested with 500 hosts
- [ ] Terminal usable with touch input
- [ ] Biometric unlock works on test devices
- [ ] Notifications received correctly
- [ ] App store guidelines compliance check
- [ ] Performance: 60 FPS, < 50MB
- [ ] Sprint retrospective completed

---

## Sprint Retrospective Notes

*To be filled after sprint completion*

**What went well:**
-

**What could be improved:**
-

**Action items for next sprint:**
-
