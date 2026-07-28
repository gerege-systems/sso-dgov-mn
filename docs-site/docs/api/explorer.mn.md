# API Explorer

Доорх интерактив OpenAPI explorer нь backend-ийн үүсгэсэн spec-ээс
(`backend/docs/swagger.json`) рендер хийгддэг. Эрх бүхий, кодоор баталгаажсан
endpoint жагсаалтын хувьд [REST API](index.md) лавлахыг ашиглана уу — үүсгэсэн
spec нь мэдэгдэж буй дутагдал болон хуучирсан бичлэгтэй (тэнд тэмдэглэсэн).

!!! warning "Spec-ийн анхааруулга"
    - Зарим бүлэг (RBAC, Provider, admin-users) нь spec-д байхгүй.
    - Spec дэх зарим password/OTP/бүртгэлийн auth зам нь кодоос **устгагдсан** (eID
      бол цорын ганц нэвтрэх арга).
    - Spec-ийн `host` нь хөгжүүлэлтийн өгөгдмөл `localhost:8080`.

!!swagger ../assets/swagger.json!!
