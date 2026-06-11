Tính năng render ra file HTML tĩnh một lần duy nhất rồi vứt lên CDN **không phải là tính năng mặc định của mọi công nghệ SSR, mà nó là một cơ chế nâng cao (Optional)**.

Trong thế giới lập trình, tính năng này có một cái tên học thuật riêng là **SSG (Static Site Generation)** hoặc phiên bản nâng cao của nó là **ISR (Incremental Static Regeneration)**.

Để hiểu rõ tại sao nó là tùy chọn (Optional) và các loại SSR hỗ trợ nó ra sao, chúng ta cần phân biệt 2 nhóm công nghệ SSR:

---

## 1. Nhóm SSR Truyền thống (Không tự động hỗ trợ)

Các công nghệ SSR đời đầu hoặc các backend thuần túy như **PHP (Laravel/WordPress thuần), Java (Spring Boot), C# (.NET), hay Node.js thuần (Express)** không tự động làm việc này.

* **Bản chất:** Theo mặc định, cứ có request là các công nghệ này sẽ "bật máy" lên để render lại từ đầu. Nó là một cỗ máy sinh code động (Dynamic).
* **Muốn lưu thành file tĩnh lên CDN?** Lập trình viên phải **tự cấu hình thêm các công cụ bên ngoài**. Tòa soạn phải mua thêm dịch vụ CDN (như Cloudflare, Fastly) hoặc cài thêm các lớp phần mềm đệm (như Varnish Cache, Nginx Cache) đứng trước Server SSR để hứng lấy file HTML vừa tạo, đóng băng nó lại rồi đi phân phối.

---

## 2. Nhóm SSR Hiện đại (Hỗ trợ tận răng - Bật/Tắt bằng 1 dòng code)

Các Framework SSR hiện đại (ra đời từ sau năm 2016) như **Next.js, Nuxt.js, Astro, Remix** được thiết kế riêng cho thời đại điện toán đám mây, nên chúng **tích hợp sẵn** tính năng này bên trong lõi.

Tuy nhiên, nó vẫn là **Optional (Tùy chọn)**. Bạn thích trang nào chạy SSR động, trang nào biến thành file tĩnh lên CDN là do bạn quyết định bằng cách cấu hình:

### Ví dụ trong Next.js / Astro:

* **Lựa chọn 1 (SSR thuần túy - Dynamic):** Bạn cấu hình trang web ở chế độ *Force Dynamic*. Cứ mỗi giây người dùng bấm vào, server lại chạy code để lấy giá vàng, tỷ giá mới nhất.
* **Lựa chọn 2 (Biến thành file tĩnh - Static/ISR):** Bạn chỉ cần thêm một dòng lệnh cấu hình thời gian hết hạn (ví dụ: `revalidate = 3600` - tức là 1 tiếng). Next.js/Astro sẽ tự hiểu: *"À, bài báo này tôi chỉ render đúng 1 lần đầu tiên thôi. Trong vòng 1 tiếng tiếp theo, ai vào đọc thì cứ lấy file HTML cũ đó từ bộ nhớ đệm ra mà dùng, cấm làm phiền Server!"*.

---

## Tại sao các trang báo lớn bắt buộc phải bật "Tùy chọn" này?

Dù dùng công nghệ cũ (phải tự cài thêm app Cache) hay công nghệ mới (Framework hỗ trợ sẵn), các trang báo lớn đều **bắt buộc** phải kích hoạt tính năng "Đóng băng thành file tĩnh" này vì 2 lý do cực kỳ thực tế:

1. **Sống sót qua các đợt bão Traffic:** Khi có một sự kiện chấn động (ví dụ: Chung kết bóng đá, Thiên tai, Tin giật gân), lượng truy cập có thể tăng đột ngột gấp 500 lần bình thường. Nếu để SSR chạy động thuần túy, Server sẽ sập sau 3 giây. Biến nó thành file tĩnh nằm trên CDN là cách duy nhất để trang web sống sót.
2. **Tiết kiệm tiền:** Tiền thuê năng lượng CPU để Server chạy code SSR đắt gấp hàng chục lần so với tiền thuê ổ đĩa CDN để lưu trữ và phân phối file HTML tĩnh.

> **Tóm lại:** Tính năng này là **Optional (Tùy chọn)**. Bạn có quyền bật hoặc tắt. Nhưng riêng đối với ngành làm báo điện tử, đây là một tùy chọn **"Bắt buộc phải bật"** nếu không muốn sập web và phá sản vì tiền thuê server.