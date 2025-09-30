package secure

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/valyala/bytebufferpool"
)

const NonceSize = 16

// nonceContextKey is an unexported type used as a key for context values.
// nonceContextKey는 컨텍스트 값의 키로 사용되는 비공개 타입입니다.
type nonceContextKey struct{}

// cryptoRandNonce generates NonceSize bytes of random data using crypto/rand and writes it to the given io.Writer after base64 encoding.
// It panics if random data generation fails.
// cryptoRandNonce는 crypto/rand를 이용해 NonceSize 바이트의 랜덤 데이터를 생성하고, base64 인코딩 후 주어진 io.Writer에 씁니다.
// 랜덤 데이터 생성 실패 시 panic을 발생시킵니다.
func cryptoRandNonce(w io.Writer) {
	var buf [NonceSize]byte
	if _, err := io.ReadFull(rand.Reader, buf[:]); err != nil {
		panic("secure: " + err.Error())
	}
	var encoded [24]byte
	base64.RawStdEncoding.Encode(encoded[:], buf[:])
	w.Write(encoded[:base64.RawStdEncoding.EncodedLen(NonceSize)])
}

// CSPConfig is a configuration struct for dynamically generating the Content-Security-Policy header.
// Each field corresponds to a CSP directive's source list.
// CSPConfig는 Content-Security-Policy 헤더를 동적으로 생성하기 위한 설정 구조체입니다.
// 각 필드는 CSP 지시문에 해당하는 소스 목록을 담습니다.
type CSPConfig struct {
	DefaultSrc  []string
	StyleSrc    []string
	ScriptSrc   []string
	ImgSrc      []string
	FontSrc     []string
	ConnectSrc  []string
	FrameSrc    []string
	MediaSrc    []string
	ObjectSrc   []string
	ManifestSrc []string
	FormAction  []string
}

// NonceHeaders is a middleware factory that takes a CSPConfig and returns a middleware function.
// This middleware generates a random nonce, stores it in the request context, and applies a dynamic Content-Security-Policy header to the response.
// NonceHeaders는 CSPConfig를 받아 미들웨어 함수를 반환하는 미들웨어 팩토리입니다.
// 이 미들웨어는 랜덤 nonce 값을 생성하여 요청 컨텍스트에 저장하고, 설정을 기반으로 동적 Content-Security-Policy 헤더를 응답에 적용합니다.
func NonceHeaders(config CSPConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buff := bytebufferpool.Get()
			defer bytebufferpool.Put(buff) // Ensure the buffer is returned even if a panic occurs.

			cryptoRandNonce(buff)
			nonce := buff.String()
			r = r.WithContext(context.WithValue(r.Context(), nonceContextKey{}, nonce))

			cspValue := buildCSP(config, nonce)
			if cspValue != "" {
				w.Header().Set("Content-Security-Policy", cspValue)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// buildCSP constructs the final CSP header string based on the CSPConfig and a nonce value.
// buildCSP는 CSPConfig와 nonce 값을 기반으로 최종 CSP 헤더 문자열을 생성합니다.
func buildCSP(config CSPConfig, nonce string) string {
	var directives []string
	nonceStr := "'nonce-" + nonce + "'"

	addDirective := func(name string, values []string, addNonce bool) {
		if values == nil {
			return // Skip if the directive is not set.
		}

		allValues := make([]string, len(values))
		copy(allValues, values)

		if addNonce {
			allValues = append(allValues, nonceStr)
		}

		directives = append(directives, name+" "+strings.Join(allValues, " "))
	}

	addDirective("default-src", config.DefaultSrc, true)
	addDirective("style-src", config.StyleSrc, true)
	addDirective("script-src", config.ScriptSrc, true)
	addDirective("img-src", config.ImgSrc, false)
	addDirective("font-src", config.FontSrc, false)
	addDirective("connect-src", config.ConnectSrc, false)
	addDirective("frame-src", config.FrameSrc, false)
	addDirective("media-src", config.MediaSrc, false)
	addDirective("object-src", config.ObjectSrc, false)
	addDirective("manifest-src", config.ManifestSrc, false)
	addDirective("form-action", config.FormAction, false)

	return strings.Join(directives, "; ")
}

// GetNonce retrieves the nonce value from the context. It panics if the nonce is not found.
// GetNonce는 컨텍스트에서 nonce 값을 가져옵니다. Nonce를 찾지 못하면 panic을 발생시킵니다.
func GetNonce(ctx context.Context) string {
	val := ctx.Value(nonceContextKey{})
	if val == nil {
		panic("nonce empty")
	}
	return val.(string)
}

// SecurityHeaders is a middleware that sets several security-related HTTP headers to the response.
// SecurityHeaders 미들웨어는 여러 보안 관련 HTTP 헤더들을 응답에 설정합니다.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware sets Cross-Origin Resource Sharing (CORS) headers and handles preflight requests.
// CORSMiddleware 미들웨어는 CORS 관련 헤더를 설정하며, preflight 요청을 처리합니다.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
