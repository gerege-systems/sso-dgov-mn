// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Core клиентийн гэрээг тогтоосон тестүүд. Жинхэнэ Core рүү хандахгүй —
// httptest сервер нь Core-ийн БОДИТ зан үйлийг дуурайна (бодит API-аас
// ажиглаж тэмдэглэсэн: user/find нь country_code шаарддаг, "олдсонгүй"-г
// 500 статустай хэрнээ мессежтэй буцаадаг).
package core_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	coreuc "template/internal/business/usecases/core"
)

func TestFindUsersSendsCountryCode(t *testing.T) {
	var gotPath, gotMethod, gotAuth string
	var body map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod, gotAuth = r.URL.Path, r.Method, r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"civil_id":123}`))
	}))
	defer srv.Close()

	uc := coreuc.NewUsecase(srv.URL, "test-token")
	out, err := uc.FindUsers(context.Background(), "111949212017")
	if err != nil {
		t.Fatalf("FindUsers: %v", err)
	}

	if gotMethod != http.MethodPost || gotPath != "/api/user/find" {
		t.Fatalf("called %s %s, want POST /api/user/find", gotMethod, gotPath)
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if body["search_text"] != "111949212017" {
		t.Fatalf("search_text = %q", body["search_text"])
	}
	// Үүнгүйгээр Core "country_code is required" гэж татгалздаг — энэ талбар
	// дутуу байсан нь хайлт ажиллахгүй байсан жинхэнэ шалтгаануудын нэг.
	if body["country_code"] != "MN" {
		t.Fatalf("country_code = %q; Core rejects the request without it", body["country_code"])
	}
	if !strings.Contains(string(out), `"civil_id"`) {
		t.Fatalf("response not passed through: %s", out)
	}
}

func TestFindOrganizationsUsesQueryParam(t *testing.T) {
	var gotMethod, gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.Query().Get("search_text")
		_, _ = w.Write([]byte(`{"reg_no":"6160617"}`))
	}))
	defer srv.Close()

	uc := coreuc.NewUsecase(srv.URL, "test-token")
	if _, err := uc.FindOrganizations(context.Background(), "6160617"); err != nil {
		t.Fatalf("FindOrganizations: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/organization/find" {
		t.Fatalf("called %s %s, want GET /api/organization/find", gotMethod, gotPath)
	}
	if gotQuery != "6160617" {
		t.Fatalf("search_text query = %q", gotQuery)
	}
}

// Core нь "олдсонгүй"-г 500 статустай хэрнээ уншигдах мессежтэй буцаадаг.
// Түүнийг дотоод алдаа болгон залгивал оператор зөвхөн "internal server error"
// хараад, юу болсныг мэдэхгүй байсан.
func TestUpstreamMessageIsPassedThroughOnErrorStatus(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"not found", http.StatusInternalServerError, `{"message":"мэдээлэл олдсонгүй"}`},
		{"validation", http.StatusBadRequest, `{"message":"Байгууллагын регистр оруулна уу"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			out, err := coreuc.NewUsecase(srv.URL, "t").FindOrganizations(context.Background(), "x")
			if err != nil {
				t.Fatalf("an upstream message must reach the caller, not become an error: %v", err)
			}
			if !strings.Contains(string(out), "олдсонгүй") && !strings.Contains(string(out), "регистр") {
				t.Fatalf("upstream message lost: %s", out)
			}
		})
	}
}

// JSON биш хог биетийг дамжуулахгүй — тэр нь дотоод алдаа.
func TestNonJSONErrorBodyBecomesAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html>gateway error</html>"))
	}))
	defer srv.Close()

	if _, err := coreuc.NewUsecase(srv.URL, "t").FindUsers(context.Background(), "x"); err == nil {
		t.Fatal("a non-JSON upstream failure should surface as an error")
	}
}

// Token тохируулаагүй бол Core инерт — 500 биш, тайлбартай мессеж.
func TestWithoutTokenTheServiceIsInert(t *testing.T) {
	out, err := coreuc.NewUsecase("https://core.gerege.mn", "").FindUsers(context.Background(), "x")
	if err != nil {
		t.Fatalf("an unconfigured Core must not error: %v", err)
	}
	if !strings.Contains(string(out), "CORE_API_TOKEN") {
		t.Fatalf("the message should tell the operator what to configure: %s", out)
	}
}
