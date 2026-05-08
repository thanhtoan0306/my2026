## Goal

Develop in VS Code with a clean project structure, but **export a single `.ino`** that you can open in **Arduino IDE** and upload.

This folder is that VS Code project. The exported sketch is written to `dist/slideLEDandOLED.ino`.

## Layout

- `src/`
  - `sketch.ino`: the main sketch entry (`setup()`/`loop()`)
  - `config.h`: pins, constants, WiFi placeholders, etc
  - `web_ui.h`: HTML response as a raw string (keeps `sketch.ino` readable)
- `tools/export_ino.py`: concatenates the sources into **one** `.ino`
- `dist/slideLEDandOLED.ino`: generated output (open this in Arduino IDE)

## Export to one `.ino`

From this folder (`may/esp32c3/esp32vscode`):

```bash
python3 tools/export_ino.py
```

Then open `dist/slideLEDandOLED.ino` in Arduino IDE.

## Optional: build from CLI (no upload)

If you use `arduino-cli`, you can compile the exported `.ino`:

```bash
arduino-cli compile --fqbn esp32:esp32:esp32c3 --libraries "$HOME/Documents/Arduino/libraries" dist/slideLEDandOLED.ino
```

Notes:
- You must have the ESP32 core installed in Arduino IDE / arduino-cli.
- Libraries (like `U8g2`) must be installed where Arduino can find them.

