# Wolf Sprite Sheet

Mô tả sprite sheet sói pixel art dùng cho game 2D.

## Thông số kỹ thuật

| Thuộc tính | Giá trị |
|------------|---------|
| Kích thước ảnh | **500 × 500** px |
| Lưới | **5 cột × 5 hàng** |
| Số frame | **25** |
| Kích thước mỗi frame | **100 × 100** px |
| File | `assets/woft-5c5r-500x500.png` |
| Nền | Đen / trong suốt (`#000000`) |
| Hướng nhân vật | Nhìn sang phải (side view) |

### Công thức tọa độ frame

```
cột = index % 5        (0–4)
hàng = floor(index / 5) (0–4)

sx = cột × 100
sy = hàng × 100
```

Ví dụ trong code (Canvas / Phaser / Godot):

```js
const FRAME = 100;
const COLS = 5;

function frameRect(row, col) {
  return { x: col * FRAME, y: row * FRAME, w: FRAME, h: FRAME };
}
```

---

## Bản đồ hoạt ảnh (5×5)

```
        Cột 0      Cột 1      Cột 2      Cột 3      Cột 4
      ┌──────────┬──────────┬──────────┬──────────┬──────────┐
Hàng 0│ Idle 1   │ Idle 2   │ Idle 3   │ Idle 4   │ Idle 5   │
      ├──────────┼──────────┼──────────┼──────────┼──────────┤
Hàng 1│ Walk 1   │ Walk 2   │ Walk 3   │ Walk 4   │ Walk 5   │
      ├──────────┼──────────┼──────────┼──────────┼──────────┤
Hàng 2│ Run 1    │ Run 2    │ Run 3    │ Run 4    │ Run 5    │
      ├──────────┼──────────┼──────────┼──────────┼──────────┤
Hàng 3│ Attack 1 │ Attack 2 │ Attack 3 │ Attack 4 │ Attack 5 │
      ├──────────┼──────────┼──────────┼──────────┼──────────┤
Hàng 4│ Howl     │ Crouch   │ Death 1  │ Death 2  │ Death 3  │
      └──────────┴──────────┴──────────┴──────────┴──────────┘
```

---

## Chi tiết từng hàng

### Hàng 0 — Idle (đứng yên)

| Frame | Mô tả |
|-------|--------|
| 0–4 | Sói đứng yên, nhìn sang phải. Miệng hé / khép nhẹ — chu kỳ thở. |

- **Loop:** có  
- **FPS gợi ý:** 6  

---

### Hàng 1 — Walk (đi bộ)

| Frame | Mô tả |
|-------|--------|
| 0–4 | Chu kỳ đi bộ bốn chân đầy đủ. |

- **Loop:** có  
- **FPS gợi ý:** 10  

---

### Hàng 2 — Run (chạy)

| Frame | Mô tả |
|-------|--------|
| 0–4 | Chu kỳ chạy nhanh, chân động mạnh hơn Walk. |

- **Loop:** có  
- **FPS gợi ý:** 14  

---

### Hàng 3 — Attack (nhảy / tấn công)

| Frame | Mô tả |
|-------|--------|
| 0 | Đứng chuẩn bị. |
| 1–2 | Lao về phía trước. |
| 3 | Nhảy vồ — miệng há, giữa không trung. |
| 4 | Hạ cánh, về tư thế đứng. |

- **Loop:** thường **không** (chơi một lần rồi về Idle)  
- **FPS gợi ý:** 12  

---

### Hàng 4 — Howl & Death (hành động đặc biệt)

| Cột | Tên | Mô tả |
|-----|-----|--------|
| 0 | **Howl** | Ngửa đầu lên, miệng mở — tru. |
| 1 | **Crouch** | Bắt đầu gục / cúi xuống. |
| 2–4 | **Death** | Chuỗi chết: ngã sấp → nằm sấp → nằm bẹp hoàn toàn. |

- **Howl / Crouch:** 1 frame, one-shot  
- **Death:** phát tuần tự cột 2→3→4, **không lặp**  
- Có thể phát cả hàng 4 như demo (nút **Misc ×5** trong `index.html`)

---

## Chạy thử

```bash
open june/playhtm/index.html
```

Hoặc phục vụ tĩnh:

```bash
cd june/playhtm && python3 -m http.server 8080
# http://localhost:8080
```

---

## Gợi ý tích hợp game

| Trạng thái game | Hàng / animation |
|-----------------|------------------|
| Đứng yên | Idle (0) |
| Di chuyển chậm | Walk (1) |
| Di chuyển nhanh | Run (2) |
| Đánh / cắn | Attack (3), one-shot |
| Tru | Howl (4, cột 0) |
| Rình / gục | Crouch (4, cột 1) |
| Chết | Death (4, cột 2–4), one-shot |
