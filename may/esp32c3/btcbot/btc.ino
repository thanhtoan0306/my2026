#include <WiFi.h>
#include <WiFiClientSecure.h>
#include <HTTPClient.h>
#include <ArduinoJson.h>
#include <U8g2lib.h>
#include <Wire.h>

#if __has_include("secrets.h")
#include "secrets.h"
#endif

#ifndef TELEGRAM_BOT_TOKEN
#define TELEGRAM_BOT_TOKEN "8523312836:AAFj5tWryVxmf7b32DYQLlZHZiLgawf00mE"
#endif
#ifndef TELEGRAM_CHAT_ID
#define TELEGRAM_CHAT_ID "5642813697"
#endif

// Replace with your network credentials
const char* ssid = "INFINITY";
const char* password = "1n@Finity";

// CoinGecko API URL
const char* serverName = "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd";

// ---- OLED pins ----
#define SDA_PIN 5
#define SCL_PIN 6
U8G2_SSD1306_72X40_ER_F_HW_I2C u8g2(U8G2_R0, /* reset=*/ U8X8_PIN_NONE);
// ---- Display helper ----
void drawOLEDMessage(const char* msg) {

  u8g2.clearBuffer();
  u8g2.setDrawColor(1);

  // 1. Set your bold font
  u8g2.setFont(u8g2_font_ncenB14_tr); 
  
  // 2. Get the width of the specific message
  int stringWidth = u8g2.getStrWidth(msg);
  
  // 3. Calculate X to center it (Display width is 72)
  int x = (72 - stringWidth) / 2;
  
  // 4. Calculate Y to center it vertically (Display height is 40)
  // Note: For Y, we usually add a bit for the font's "ascent"
  int y = 28; 
  
  // 5. Draw it
  u8g2.drawStr(x, y, msg);
  
  u8g2.sendBuffer();
}

String formatNumber(long value) {
  String res = String(value);
  int insertPosition = res.length() - 3;
  
  while (insertPosition > 0) {
    res = res.substring(0, insertPosition) + "," + res.substring(insertPosition);
    insertPosition -= 3;
  }
  return res;
}

static bool telegramConfigured() {
  return strlen(TELEGRAM_BOT_TOKEN) > 0 && strlen(TELEGRAM_CHAT_ID) > 0;
}

/** Sends the same text shown on the OLED (plain price string). */
void sendTelegramMessage(const String& text) {
  if (!telegramConfigured()) {
    return;
  }

  WiFiClientSecure client;
  client.setInsecure();

  HTTPClient https;
  String url =
      String("https://api.telegram.org/bot") + TELEGRAM_BOT_TOKEN + "/sendMessage";

  if (!https.begin(client, url)) {
    Serial.println("Telegram: HTTPS begin failed");
    return;
  }

  https.addHeader("Content-Type", "application/json");

  String body = "{\"chat_id\":";
  body += TELEGRAM_CHAT_ID;
  body += ",\"text\":\"";
  body += text;
  body += "\"}";

  int httpCode = https.POST(body);

  if (httpCode >= 200 && httpCode < 300) {
    Serial.println("Telegram: message sent");
  } else {
    Serial.print("Telegram: HTTP ");
    Serial.print(httpCode);
    Serial.print(" — ");
    Serial.println(https.errorToString(httpCode));
    if (https.getSize() > 0) {
      Serial.println(https.getString());
    }
  }

  https.end();
}

void setup() {
  Wire.begin(SDA_PIN, SCL_PIN);
  u8g2.begin();

  Serial.begin(115200);
  Wire.begin(SDA_PIN, SCL_PIN);
  u8g2.begin();

  WiFi.begin(ssid, password);
  Serial.print("Connecting to WiFi");

  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("\nConnected to WiFi network");
}

void loop() {

  if (WiFi.status() == WL_CONNECTED) {
    HTTPClient http;

    // Initialize the request
    http.begin(serverName);
    
    // Send the GET request
    int httpResponseCode = http.GET();

    if (httpResponseCode > 0) {
      Serial.print("HTTP Response code: ");
      Serial.println(httpResponseCode);
      
      String payload = http.getString();
      
      // Parse the JSON response
      // Structure: {"bitcoin":{"usd":80983}}
      StaticJsonDocument<200> doc;
      DeserializationError error = deserializeJson(doc, payload);

      if (!error) {
        long btcPrice = doc["bitcoin"]["usd"];
        String formattedPrice = formatNumber(btcPrice);
        Serial.println("--- BTC Update ---");
        Serial.print("Price: $");
        Serial.println(formattedPrice);
        Serial.println("------------------");

        String ledText = formattedPrice;
        drawOLEDMessage(ledText.c_str());
        sendTelegramMessage(ledText);
      } else {
        Serial.print("deserializeJson() failed: ");
        Serial.println(error.f_str());
      }
    } else {
      Serial.print("Error code: ");
      Serial.println(httpResponseCode);
    }
    
    // Free resources
    http.end();
  } else {
    Serial.println("WiFi Disconnected");
  }

  // Update every 30 seconds (CoinGecko free tier has rate limits!)
  delay(30000);
}