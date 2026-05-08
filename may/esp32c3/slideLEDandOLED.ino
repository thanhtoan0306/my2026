#include <WiFi.h>
#include <U8g2lib.h>
#include <Wire.h>

// ---- WiFi settings ----
const char* ssid = "INFINITY";
const char* password = "1n@Finity";

// ---- Pins ----
#define LED_PIN 8
#define SDA_PIN 5
#define SCL_PIN 6

// ---- OLED Object ----
U8G2_SSD1306_72X40_ER_F_HW_I2C u8g2(U8G2_R0, /* reset=*/ U8X8_PIN_NONE);

// ---- Timing Variables ----
int ledDelay = 500;        // LED speed
int oledDelay = 300;       // Alphabet speed
unsigned long lastLedMs = 0;
unsigned long lastOledMs = 0;

bool ledState = LOW;
char currentLetter = 'A';

WiFiServer server(80);

void setup() {
  Serial.begin(115200);
  pinMode(LED_PIN, OUTPUT);
  
  // Start OLED
  Wire.begin(SDA_PIN, SCL_PIN);
  u8g2.begin();
  u8g2.setFont(u8g2_font_logisoso24_tr);

  // Start WiFi
  WiFi.begin(ssid, password);
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
  if (millis() - lastLedMs >= ledDelay) {
    lastLedMs = millis();
    ledState = !ledState;
    digitalWrite(LED_PIN, ledState);
  }

  // 2. Handle OLED Alphabet (Non-blocking)
  if (millis() - lastOledMs >= oledDelay) {
    lastOledMs = millis();
    u8g2.clearBuffer();
    char buf[2] = {currentLetter, '\0'};
    u8g2.drawStr(24, 34, buf); 
    u8g2.sendBuffer();
    currentLetter++;
    if (currentLetter > 'Z') currentLetter = 'A';
  }

  // 3. Handle Web Server
  WiFiClient client = server.available();
  if (client) {
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

    client.println("HTTP/1.1 200 OK\nContent-type:text/html\n");
    client.println("<html><head><meta name='viewport' content='width=device-width, initial-scale=1'>");
    client.println("<style>body{text-align:center; font-family:sans-serif; background:#111; color:#eee;} .s{width:80%; margin:20px;}</style></head><body>");
    
    client.println("<h1>C3 Controller</h1>");
    
    client.print("<p>LED Speed: "); client.print(ledDelay); client.println("ms</p>");
    client.println("<input type='range' min='50' max='2000' value='"+String(ledDelay)+"' class='s' onchange='window.location.href=\"/?led=\"+this.value'>");
    
    client.print("<p>OLED Speed: "); client.print(oledDelay); client.println("ms</p>");
    client.println("<input type='range' min='50' max='2000' value='"+String(oledDelay)+"' class='s' onchange='window.location.href=\"/?oled=\"+this.value'>");
    
    client.println("</body></html>");
    client.stop();
  }
}