# 📘 Hướng dẫn Sử Dung Apidog cho VN Administrative API

Tài liệu này hướng dẫn cách thiết lập và sử dụng Apidog một cách chuyên nghiệp cho dự án này, từ bước cài đặt đến kiểm thử tự động.

## 1. Cài đặt & Chuẩn bị
1.  **Tải Apidog**: [https://apidog.com/download/](https://apidog.com/download/)
2.  **Khởi động Server**:
    Đảm bảo API của bạn đang chạy tại localhost.
    ```bash
    go run cmd/server/main.go
    ```

## 2. Import API (Nhanh & Chuẩn)
Thay vì tạo từng request thủ công, hãy dùng file `openapi.yaml` mình đã chuẩn bị sẵn.

1.  Mở Apidog -> Vào Project của bạn.
2.  Chọn **Settings** (hoặc icon +) -> **Import Data**.
3.  Chọn tab **OpenAPI/Swagger**.
4.  Kéo thả file `openapi.yaml` từ thư mục dự án vào.
5.  Nhấn **Confirm**.
    -> *Toàn bộ 3 endpoint (`/provinces`, `/units`, `/search`) sẽ hiển thị đầy đủ với document và ví dụ.*

## 3. Thiết lập Môi trường (Environments)
Để không phải sửa URL thủ công khi đổi môi trường (Local -> Prod).

1.  Nhìn góc trên bên phải, nhấn vào nút quản lý môi trường (thường là "No Environment" hoặc icon ⚙️).
2.  Tab **Global Variable** hoặc **Environment**:
    *   Tạo môi trường **Local Env**:
        *   `Service URL`: `http://localhost:8080`
    *   Tạo môi trường **Production Env** (ví dụ sau này):
        *   `Service URL`: `https://api.myapp.com`
3.  Lưu lại.
4.  Khi test, chỉ cần chọn **Local Env** ở dropdown góc phải. URL trong request sẽ tự động hiểu là `{{baseUrl}}/api/v1/...`.

## 4. Kiểm thử Tự động (Assertions)
Giúp bạn validate API đúng sai mà không cần nhìn bằng mắt.

1.  Mở request `GET /api/v1/search`.
2.  Điền Params `q` = `Hà Nội`.
3.  Chuyển sang tab **Post-processors**.
4.  Chọn **Add Post-processor** -> **Assertion**.
5.  Thiết lập rule:
    *   `JSONPath`: `$.data[0].tenhc` (Kiểm tra tên phần tử đầu tiên)
    *   `Assetion`: `Contains`
    *   `Value`: `Hà Nội`
6.  Nhấn **Send**. Nếu kết quả trả về đúng, bạn sẽ thấy thông báo **Pass** màu xanh.

## 5. Tạo Kịch bản Test (Scenario)
Test luồng người dùng thực tế: **Lấy danh sách Tỉnh -> Lấy chi tiết Quận/Huyện**.

1.  Vào menu **Testing** -> **Test Scenarios** -> **New Scenario**.
2.  Nhấn **Import from API** -> Chọn `GET /provinces` và `GET /units` theo thứ tự.
3.  **Bước 1 (Provinces)**:
    *   Vào tab **Post-processors** -> **Extract Variable**.
    *   `JSONPath`: `$.data[0].id` (Lấy ID của tỉnh đầu tiên).
    *   `Variable Name`: `province_id`.
    *   `Scope`: `Scenario Variable`.
4.  **Bước 2 (Units)**:
    *   Sửa URL hoặc Params: Thay giá trị cứng `1` thành `{{province_id}}`.
5.  Nhấn **Run**.
    -> *Apidog sẽ chạy Bước 1, lấy ID, truyền vào Bước 2 và báo cáo kết quả.*

## 6. Xuất Báo cáo
Sau khi chạy xong Scenario:
1.  Nhấn **Export Report**.
2.  Chọn format HTML.
3.  Gửi file này cho sếp/khách hàng để demo tính ổn định của hệ thống!
