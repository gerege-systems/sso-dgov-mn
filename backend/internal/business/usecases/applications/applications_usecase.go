// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package applications нь API Gateway consumer + SSO RP-ийг нэгтгэсэн
// "Applications" загварын business logic. Application бүр Hydra OAuth2 client-тэй
// тохирно (web/spa/native = authorization_code RP, m2m = client_credentials).
// Аппад зөвшөөрсөн gateway service-үүд нь client-ийн OAuth scope болно. Client
// secret-ыг Hydra эзэмшинэ; бид зөвхөн overlay (name/tags/type/enabled/services)-
// ыг PostgreSQL-д хадгална.
package applications

import (
	"context"

	"template/internal/business/domain"
)

type Usecase interface {
	// List нь бүх апп-ыг буцаана (secret-гүй).
	List(ctx context.Context) ([]domain.Application, error)
	// Get нь нэг апп-ыг id-гээр буцаана (secret-гүй).
	Get(ctx context.Context, id string) (domain.Application, error)
	// Create нь Hydra OAuth2 client + overlay мөр үүсгэнэ. Confidential (web/m2m)
	// апп-ын client_secret-ыг хариунд НЭГ удаа буцаана.
	Create(ctx context.Context, in Input) (domain.Application, error)
	// Update нь апп-ын overlay + Hydra client-ийн desired state-ыг шинэчилнэ
	// (secret хэвээр).
	Update(ctx context.Context, id string, in Input) (domain.Application, error)
	// Delete нь Hydra client болон overlay мөрийг устгана.
	Delete(ctx context.Context, id string) error
	// RotateSecret нь confidential апп-ын client_secret-ыг сольж НЭГ удаа буцаана.
	RotateSecret(ctx context.Context, id string) (domain.Application, error)
	// SetSecret нь confidential апп-д админаас өгсөн ТОДОРХОЙ client_secret-ыг
	// тавина (гадаад RP-ийн аль хэдийн тохируулсан secret-тэй тулгах хэрэгцээнд).
	SetSecret(ctx context.Context, id, secret string) (domain.Application, error)
	// SetServices нь апп-ын зөвшөөрсөн service-үүдийг (scope) сольж, Hydra-г шинэчилнэ.
	SetServices(ctx context.Context, id string, serviceIDs []string) (domain.Application, error)
	// ReconcileClients нь seed хийсэн RP overlay мөрүүд (created_by='seed-rp')-ийн
	// Hydra client дутуу байвал үүсгэнэ (startup bootstrap). Үүсгэсэн тоог буцаана.
	// Idempotent — байгаа client-ыг алгасна; устгасан RP-г дахин үүсгэхгүй (мөр
	// байхгүй тул). Түр зуур Hydra бэлэн бус бол алдаа буцаана (дуудагч warn-лоно).
	ReconcileClients(ctx context.Context) (int, error)
}

// Input нь апп үүсгэх/шинэчлэх талбарууд. AppType нь grant/auth-method-ыг тодорхойлно.
type Input struct {
	Name         string
	AppType      string
	RedirectURIs []string
	Tags         []string
	ServiceIDs   []string
	Enabled      bool
}
