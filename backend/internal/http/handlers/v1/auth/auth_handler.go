// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package auth нь /auth/* HTTP endpoint-уудыг үйлчилдэг — register,
// login, OTP, refresh, logout. Хэрэглэгчийн профайлын endpoint-ууд нь
// ах дүү package болох internal/http/handlers/v1/users-д байрладаг.
package auth

import (
	"context"

	"template/internal/business/usecases/audit"
	"template/internal/business/usecases/auth"
	"template/pkg/eid"
)

// LoginAppResolver нь OIDC login_challenge-аас нэвтэрч буй бүртгэгдсэн RP апп-ийн
// "дэд систем" контекстийг (subsystem нэр, sub_url домэйн) тодорхойлно. Base SSO,
// first-party client, эсвэл хоосон/буруу challenge үед хоосон утга буцаана
// (провайдер usecase хэрэгжүүлнэ; Hydra тохируулаагүй бол nil — hooson).
type LoginAppResolver interface {
	LoginAppContext(ctx context.Context, challenge string) (subsystem, subURL string)
}

// Handler нь auth-handler-ийн нэгтгэл; endpoint бүрийн method-ууд
// өөрсдийн файлд (auth.register.go, auth.login.go, г.м.) тодорхойлогддог
// тул нэг endpoint-д хүрэх PR diff-үүд бусад руу нэвчдэггүй.
//
// auditUC нь persisted hash-chained audit log-д бичих use case (eID нэвтрэлт
// амжилттай болоход best-effort бичлэг хийнэ). nil байж болно — тэр үед audit
// бичлэг алгасагдана (тестүүдэд эсвэл audit идэвхгүй орчинд).
//
// loginApp нь eID push-д дамжуулах RP апп контекстийг login_challenge-аас
// resolve хийнэ. nil байж болно (Hydra тохируулаагүй / тест) — тэр үед subsystem
// хоосон (base SSO гэж үзнэ).
type Handler struct {
	usecase  auth.Usecase
	auditUC  audit.Usecase
	loginApp LoginAppResolver
}

func NewHandler(usecase auth.Usecase) Handler {
	return Handler{usecase: usecase}
}

// NewHandlerWithAudit нь audit use case + login-app resolver-ийг тарьж handler
// үүсгэнэ. loginApp nil байж болно (subsystem хоосон болно).
func NewHandlerWithAudit(usecase auth.Usecase, auditUC audit.Usecase, loginApp LoginAppResolver) Handler {
	return Handler{usecase: usecase, auditUC: auditUC, loginApp: loginApp}
}

// resolveApp нь login_challenge-аас eID.AppContext-ийг угсарна. loginApp nil бол
// (эсвэл base/first-party) хоосон контекст — eID тал "SSO өөрөө" гэж үзнэ.
func (h Handler) resolveApp(ctx context.Context, challenge string) eid.AppContext {
	if h.loginApp == nil || challenge == "" {
		return eid.AppContext{}
	}
	sub, url := h.loginApp.LoginAppContext(ctx, challenge)
	return eid.AppContext{Subsystem: sub, SubURL: url}
}
