// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package core нь Gerege Core (core.dgov.mn)-ийн USER FIND / ORG FIND
// үйлчилгээг wrap хийнэ. core.dgov.mn-ий хариуг (json) дамжуулна.
package core

import (
	"context"
	"encoding/json"
)

type Usecase interface {
	// FindUsers нь core.dgov.mn /api/user/find руу search_text-ээр хайна
	// (core_id эсвэл регистрийн дугаар). Хариуг raw JSON-оор буцаана.
	FindUsers(ctx context.Context, searchText string) (json.RawMessage, error)
	// FindOrganizations нь core.dgov.mn /api/organization/find руу хайна.
	FindOrganizations(ctx context.Context, searchText string) (json.RawMessage, error)
}
