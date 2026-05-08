// AUTO-GENERATED FILE. DO NOT EDIT.
// Source: /Users/fe.tony/Desktop/reviewcode/my2026/may/esp32c3/esp32vscode
// Re-generate by running: python3 tools/export_ino.py


// ===== BEGIN config.h =====
// ---- WiFi settings ----
// For a shareable repo, keep placeholders here and edit in the exported `.ino` if needed.
static const char* WIFI_SSID = "INFINITY";
static const char* WIFI_PASSWORD = "1n@Finity";

// ---- Pins (ESP32-C3) ----
static constexpr int LED_PIN = 8;
static constexpr int SDA_PIN = 5;
static constexpr int SCL_PIN = 6;
// ===== END config.h =====

// ===== BEGIN web_ui.h =====
// Kept as a raw literal so `sketch.ino` stays clean.
static const char kPageHtml[] PROGMEM = R"HTML(
<html>
  <head>
    <meta name='viewport' content='width=device-width, initial-scale=1'>
    <style>
      body{ text-align:center; font-family:sans-serif; background:#111; color:#eee; }
      .s{ width:80%; margin:20px; }
    </style>
  </head>
  <body>
    <h1>C3 Controller</h1>
    %LED_BLOCK%
    %OLED_BLOCK%
  </body>
</html>
)HTML";
// ===== END web_ui.h =====

// ===== BEGIN sketch.ino =====
#include <WiFi.h>
#include <U8g2lib.h>
#include <Wire.h>
// ---- OLED Object ----
U8G2_SSD1306_72X40_ER_F_HW_I2C u8g2(U8G2_R0, /* reset=*/ U8X8_PIN_NONE);

// ---- Timing Variables ----
int ledDelay = 500;   // LED speed
int oledDelay = 300;  // Alphabet speed
unsigned long lastLedMs = 0;
unsigned long lastOledMs = 0;

bool ledState = LOW;

// 35 common symbols (Unicode code points) rendered via u8g2's symbols font.
// If you want different icons, replace items in this list.
static const uint16_t kSymbols[35] = {
    0x2600,  // ☀
    0x2601,  // ☁
    0x2602,  // ☂
    0x2603,  // ☃
    0x2605,  // ★
    0x260E,  // ☎
    0x2611,  // ☑
    0x2615,  // ☕
    0x2620,  // ☠
    0x2622,  // ☢
    0x2623,  // ☣
    0x262F,  // ☯
    0x263A,  // ☺
    0x2660,  // ♠
    0x2663,  // ♣
    0x2665,  // ♥
    0x2666,  // ♦
    0x266A,  // ♪
    0x266B,  // ♫
    0x2691,  // ⚑
    0x26A0,  // ⚠
    0x26A1,  // ⚡
    0x26BD,  // ⚽
    0x2702,  // ✂
    0x2708,  // ✈
    0x2709,  // ✉
    0x2713,  // ✓
    0x2717,  // ✗
    0x2728,  // ✨
    0x2733,  // ✳
    0x2764,  // ❤
    0x2795,  // ➕
    0x2796,  // ➖
    0x2797,  // ➗
    0x27A1,  // ➡
};
static uint8_t symbolIndex = 0;

WiFiServer server(80);

static void writeHttpOkHtml(WiFiClient& client) {
  client.println("HTTP/1.1 200 OK\nContent-type:text/html\n");
}

static void writePage(WiFiClient& client) {
  writeHttpOkHtml(client);

  // Stream HTML (small page; String ops are acceptable here)
  String html = FPSTR(kPageHtml);
  html.replace(
      "%LED_BLOCK%",
      "<p>LED Speed: " + String(ledDelay) + "ms</p>"
      "<input type='range' min='50' max='2000' value='" + String(ledDelay) +
          "' class='s' onchange='window.location.href=\"/?led=\"+this.value'>");
  html.replace(
      "%OLED_BLOCK%",
      "<p>OLED Speed: " + String(oledDelay) + "ms</p>"
      "<input type='range' min='50' max='2000' value='" + String(oledDelay) +
          "' class='s' onchange='window.location.href=\"/?oled=\"+this.value'>");

  client.print(html);
}

void setup() {
  Serial.begin(115200);
  pinMode(LED_PIN, OUTPUT);

  // Start OLED
  Wire.begin(SDA_PIN, SCL_PIN);
  u8g2.begin();
  u8g2.setFont(u8g2_font_unifont_t_symbols);

  // Start WiFi
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("\nConnected! IP: ");
  Serial.println(WiFi.localIP());
  server.begin();
}

void loop() {
  // 1. Handle LED Blinking (Non-blocking)
  if (millis() - lastLedMs >= (unsigned long)ledDelay) {
    lastLedMs = millis();
    ledState = !ledState;
    digitalWrite(LED_PIN, ledState);
  }

  // 2. Handle OLED Symbols (Non-blocking)
  if (millis() - lastOledMs >= (unsigned long)oledDelay) {
    lastOledMs = millis();
    u8g2.clearBuffer();

    // 72x40 display: draw one symbol roughly centered.
    // drawGlyph uses baseline coordinates.
    u8g2.drawGlyph(28, 30, kSymbols[symbolIndex]);
    u8g2.sendBuffer();

    symbolIndex++;
    if (symbolIndex >= (uint8_t)(sizeof(kSymbols) / sizeof(kSymbols[0]))) {
      symbolIndex = 0;
    }
  }

  // 3. Handle Web Server
  WiFiClient client = server.available();
  if (!client) return;

  String req = client.readStringUntil('\r');

  // Parse LED speed
  if (req.indexOf("GET /?led=") != -1) {
    int pos = req.indexOf("led=") + 4;
    ledDelay = req.substring(pos, req.indexOf(" ", pos)).toInt();
  }
  // Parse OLED speed
  if (req.indexOf("GET /?oled=") != -1) {
    int pos = req.indexOf("oled=") + 5;
    oledDelay = req.substring(pos, req.indexOf(" ", pos)).toInt();
  }

  writePage(client);
  client.stop();
}
// ===== END sketch.ino =====
