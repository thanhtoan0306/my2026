#pragma once

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

