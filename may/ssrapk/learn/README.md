# SSR Apk vs SSR Desktop

Same **v1.0.0 hello world**: local HTTP, SSR template, name form, dark UI.

| | Mac [`ssrdesktop`](../../ssrdesktop) | Android [`ssrapk`](../README.md) |
|---|--------------------------------------|-------------------------------------|
| Package | `.app` | `.apk` |
| Window | macOS WebKit | Android WebView |
| Server language | Go | Kotlin (NanoHTTPD) |
| Template engine | `html/template` | `{{Name}}` replace (same fields) |
| TV box | ❌ Mac only | ✅ Install APK or `adb install` |

**Why not Go inside APK?** Standard Android apps use Kotlin/Java. Go on device = Termux (`tvboxtermux`), not a launcher APK. This project mirrors the **architecture**, not the Go runtime.
