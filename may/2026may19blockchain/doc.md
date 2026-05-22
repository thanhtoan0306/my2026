Để đưa dự án về trạng thái **tối thiểu hết mức có sức** (MVP - Minimum Viable Product), bạn nên chọn làm **Web App**. Với Web App, bạn không cần xây dựng một Backend Server (Node.js/Python) riêng biệt để lưu trữ database làm gì cả.

Toàn bộ logic "Backend" ở đây thực chất là **Web3 Logic** (giao tiếp trực tiếp với Blockchain mạng Testnet thông qua RPC Node như Infura/Alchemy bằng thư viện `ethers.js` hoặc `web3.js`). Toàn bộ code này có thể chạy ngay dưới Client (Trình duyệt) để tiết kiệm thời gian tối đa.

Dưới đây là bảng chia nhỏ các task Frontend (Giao diện) và Backend/Web3 Logic (Xử lý ngầm) cho 3 màn hình tối giản của bạn:

---

## TASK 1: Màn hình Tạo / Nhập ví

### Frontend (Giao diện):

* [ ] Thiết kế UI gồm 2 nút lớn: **[Tạo ví mới]** và **[Nhập ví cũ]**.
* [ ] Tạo một ô Input Text để người dùng dán chuỗi Private Key (nếu họ chọn Nhập ví).
* [ ] Tạo một khu vực hiển thị chữ (Text Box) để hiện Private Key vừa tạo mới kèm dòng cảnh báo: *"Hãy lưu chuỗi này lại, mất là mất tiền"*.
* [ ] Tạo nút **[Đăng nhập vào Ví]** để chuyển sang Màn hình chính.

### Backend / Web3 Logic:

* [ ] Viết hàm tạo khóa: Sử dụng hàm `ethers.Wallet.createRandom()` để sinh ngẫu nhiên ra 1 cặp Private Key và Public Address.
* [ ] Viết hàm nhập khóa: Sử dụng `new ethers.Wallet(privateKeyInput)` để kiểm tra xem chuỗi Private Key người dùng nhập vào có hợp lệ hay không.
* [ ] Lưu trữ tạm thời: Lưu Private Key và Address này vào biến trạng thái (State) của ứng dụng hoặc `sessionStorage` để các màn hình sau có thể sử dụng (không cần làm database).

---

## TASK 2: Màn hình Chính (Xem số dư & Copy địa chỉ)

### Frontend (Giao diện):

* [ ] Thiết kế UI hiển thị Địa chỉ ví (Dạng rút gọn ví dụ: `0x123...abc`).
* [ ] Làm 1 nút **[Copy]** ngay bên cạnh địa chỉ ví.
* [ ] Thiết kế 1 dòng chữ lớn hiển thị số dư (Mặc định ban đầu hiện `Loading...` hoặc `0.00 ETH`).
* [ ] Thiết kế 1 nút **[Gửi tiền]** để chuyển hướng sang Màn hình gửi.

### Backend / Web3 Logic:

* [ ] Kết nối mạng (Provider): Khởi tạo kết nối với một mạng Testnet (Ví dụ: Sepolia RPC URL từ Alchemy/Infura).
* [ ] Cào số dư: Viết hàm `provider.getBalance(userAddress)` để lấy số dư từ blockchain về.
* [ ] Định dạng số dư: Đổi đơn vị từ Wei (chuỗi số lớn) sang ETH bằng hàm `ethers.formatEther(balance)` rồi trả kết quả cho Frontend hiển thị.
* [ ] Logic nút Copy: Dùng hàm `navigator.clipboard.writeText(userAddress)` của trình duyệt để xử lý lệnh copy khi bấm nút.

---

## TASK 3: Màn hình Gửi tiền

### Frontend (Giao diện):

* [ ] Tạo ô Input thứ nhất: Nhập địa chỉ ví nhận (Recipient Address).
* [ ] Tạo ô Input thứ hai: Nhập số lượng coin muốn gửi (Amount).
* [ ] Tạo nút **[Xác nhận gửi]**.
* [ ] Tạo một dòng thông báo trạng thái dưới cùng (Ví dụ: Chờ xử lý... / Gửi thành công! / Lỗi!).

### Backend / Web3 Logic:

* [ ] Tạo đối tượng ký: Khởi tạo `const wallet = new ethers.Wallet(privateKey, provider)`.
* [ ] Viết hàm gửi tiền tối giản: Gọi hàm `wallet.sendTransaction({ to: recipient, value: ethers.parseEther(amount) })`.
* *Lưu ý: Bỏ qua phần tính Gas, thư viện ethers.js sẽ tự động lấy Gas chuẩn của mạng lưới tại thời điểm đó.*
* [ ] Xử lý phản hồi: Sử dụng `try/catch`. Nếu thành công, trả về TxHash (Mã giao dịch) và báo cho Frontend hiện chữ "Thành công". Nếu thất bại (do hết tiền, sai địa chỉ), báo lỗi ra màn hình.

---

### 🛠️ Gợi ý Stack công nghệ nhanh nhất:

* **Frontend:** React (hoặc thậm chí là HTML/JS thuần nếu bạn muốn nhanh tuyệt đối).
* **Web3 Library:** `ethers.js` (Phiên bản v6).
* **Mạng thử nghiệm:** Ethereum Sepolia Testnet (Bạn có thể lên các trang Faucet để xin ETH giả lập miễn phí để test luồng chuyển tiền).

Bạn có cần mình viết sẵn bộ khung code (Template) chứa cả 3 hàm core `createWallet`, `getBalance`, và `sendTransaction` bằng `ethers.js` để bạn chỉ việc gắn vào giao diện không?