# Clean Architecture

Backend нь **хамаарал зөвхөн дотогшоо чиглэсэн** дөрвөн цагирагт Clean
Architecture-ыг дагадаг: `handler → usecase → repository → domain`.

## Хамаарлын дүрэм

| Цагираг | Package root | Import хийж болно | Import хийж болохгүй |
|------|--------------|-----------|-----------------|
| HTTP (delivery) | `internal/http/` | usecase interface-үүд (`auth.Usecase`, `users.Usecase`, …) | — |
| Usecase | `internal/business/usecases/` | `repositories/interface` + domain | postgres adapter-ууд, chi/net-http |
| Repository | `internal/datasources/repositories/postgres/` | domain, `records`, `rls` | — |
| Domain | `internal/business/domain/` | stdlib + `golang.org/x/crypto/bcrypt` | `internal/` эсвэл `pkg/` доторх юу ч |

Domain package нь өөрийн хязгаарлалтыг
`internal/business/domain/domain_users.go` дотор баримтжуулсан — энэ нь зөвхөн
стандарт сан болон `bcrypt`-ыг import хийдэг.

### Зөвшөөрөгдсөн ганц үл хамаарах зүйл

"Domain нь internal-аас юу ч import хийхгүй" гэдэгт зориудаар гаргасан ганц үл
хамаарах зүйл бий: leaf package `internal/datasources/rls/rls.go`. Энэ нь зөвхөн
stdlib `context`-оос хамаардаг бөгөөд гурван гадна цагираг бүрт хуваалцагддаг тул
import cycle үүсгэлгүйгээр хүсэлт тус бүрийн RLS identity-г зөөж чадна.

## `_interface` package

Gateway abstraction-ууд нь `internal/datasources/repositories/interface/interface.go`
дахь `_interface` package дотор амьдардаг. `interface` нь Go-ийн нөөцлөгдсөн үг
тул package-ыг `_interface` гэж нэрлэсэн; урд талын доогуур зураас нь түүнийг
хүчинтэй identifier байлгана.

Usecase-ууд нь **зөвхөн** эдгээр interface-үүдээс хамаардаг — жишээ нь
`_interface.UserRepository` — хэзээ ч конкрет `postgres` adapter-ууддаа
хамаардаггүй. Конкрет adapter нь composition root дотор бүтээгдэж, inject хийгддэг:

```go
// server.go — repo returns the interface type, not the concrete struct
userRepo := userspostgres.NewUserRepository(pool) // → _interface.UserRepository
usersUC  := users.NewUsecase(userRepo, cache, users.Config{...})
```

Энэ нь бизнесийн кодод хүрэлгүйгээр storage engine-ыг солих боломжтой болгодог
зүйл юм; `interface.go` нь ирээдүйн боломжит `mongo/` adapter-ыг ах дүү хэрэгжүүлэлт
байдлаар нэрлэсэн ч байдаг.

!!! note "Interface-үүд файлууд хооронд хуваагдсан"
    Ихэнх gateway-ууд `interface.go` дотор байдаг ч `OrgStampRepository` болон
    `UserIntegrationsRepository` нь `interface_org_stamp.go` /
    `interface_user_integrations.go` дотор байрладаг. Зөвхөн `interface.go`-г
    grep хийвэл тэдгээрийг олохгүй.

## Алдаанууд цагиргуудыг typed утга байдлаар дамждаг

Usecase-ууд `apperror.*` утгуудыг буцаадаг (`apperror.NotFound`, `Unauthorized`,
`Forbidden`, `Conflict`, `BadRequest`, `Internal`). HTTP давхарга нь
`ErrorType`-ыг `handler_base_response.go` дотор статус код руу зураглана.
Library/internal алдаанууд `apperror.InternalCause(cause)`-ээр боож
өнхрүүлэгддэг — энэ нь client-т харагдах мессежийг `"internal server error"`
болгон тогтоож, бодит шалтгааныг лог хийхэд хадгалдаг — тул боож өнхрүүлсэн
`pgx`/`sql` алдаа хэзээ ч client-т хүрэхгүй.
[Backend Layers](backend-layers.md#error-handling)-ийг үз.
