# 06 · Network

這份管 remote service、Dio、DTO、timeout、interceptor、cancelation，以及 HTTP 邊界與 Repository 的分工。

---

## 一、Dio 只存在 remote service

Dio 是 remote service 的工具，不是整個 app 的資料層。Widget、ViewModel、use case、Repository interface 都不應直接看見 `Dio`。

## 二、remote service 的輸入輸出

remote service 接收明確 request 參數，回傳 DTO 或 service result。Repository 在這一層之上決定如何 map 成 domain model 與如何處理失敗。

## 三、timeout、cancelation、interceptor

每個 remote client 都應明確設定 timeout，並把 auth、logging、retry、request id 等 cross-cutting concern 放在 interceptor 或 client config，而不是散在每個 call site。

cancelation 是正常控制流。畫面離開或搜尋條件改變時取消舊請求，不把 cancelation 當成要提示使用者的錯誤。

---

### dio-outside-remote-service · MUST_NOT · new-only+on-touch · regex

Dio 只能存在 remote service 或 API client。其他層若需要資料，應透過 Repository 或純資料物件取得，不直接 new client 或保留 `Dio` 參考。

### remote-service-uses-dto-boundary · MUST · new-only+on-touch · manual

remote service 的邊界以 DTO、request object、HTTP metadata 為主；domain model 轉換與產品錯誤翻譯在 Repository 完成。

### network-timeouts-and-cancelation-are-explicit · MUST · new-only · manual

新建 remote client 時要明確定 timeout、取消與 interceptor 行為。不要依賴隱含預設，也不要把 retry 寫死在每個畫面事件裡。

---

## 主要來源

- Flutter architecture guide services section: <https://docs.flutter.dev/app-architecture/guide#services>
- `dio` package page: <https://pub.dev/packages/dio>
