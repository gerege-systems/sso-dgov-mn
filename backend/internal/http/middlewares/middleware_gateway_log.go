// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// GatewayRequestRecorder нь нэг бодит /api хүсэлтийн телеметрийг хүлээн авна.
// Хэрэгжүүлэлт (server.go) үүнийг detached context дээр async бичдэг тул
// хариуны хоцролт нэмэгдэхгүй.
type GatewayRequestRecorder func(method, path, clientIP string, status, latencyMS int)

// GatewayRequestLogMiddleware нь DAN backend руу ирсэн бодит /api хүсэлт бүрийг
// (method/path/status/latency/ip) API Gateway-ийн хүсэлтийн лог руу бичнэ.
// Зөвхөн "/api/" замуудыг лог-лоно (health/metrics/swagger/static-ыг алгасна).
func GatewayRequestLogMiddleware(record GatewayRequestRecorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Зөвхөн /api хүсэлтийг лог-лоно. Gateway-ийн ӨӨРИЙН телеметр
			// endpoint-ууд (лог/тойм) нь admin UI-аас байнга poll хийгддэг тул
			// лог-оо өөрийн polling-оор дүүргэхээс сэргийлж алгасна.
			p := r.URL.Path
			if !strings.HasPrefix(p, "/api/") ||
				strings.HasPrefix(p, "/api/v1/gateway/logs") ||
				strings.HasPrefix(p, "/api/v1/gateway/overview") {
				next.ServeHTTP(w, r)
				return
			}
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}
			record(r.Method, r.URL.Path, clientIP(r), status, int(time.Since(start).Milliseconds()))
		})
	}
}
