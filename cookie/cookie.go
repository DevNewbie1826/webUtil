package cookie

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
	"time"
)

// cookieContextKey is an unexported type used as a key for context values.
// cookieContextKey는 컨텍스트 값의 키로 사용되는 비공개 타입입니다.
type cookieContextKey struct{}

// Middleware creates a CookieManager with a secret key and returns a middleware function
// that stores the manager in the request context.
// Middleware는 비밀키(secret)를 이용하여 CookieManager를 생성하고,
// 이를 요청 컨텍스트에 저장하는 미들웨어 함수를 반환합니다.
func Middleware(secret []byte) func(http.Handler) http.Handler {
	cm := &CookieManager{secret}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), cookieContextKey{}, cm))
			next.ServeHTTP(w, r)
		})
	}
}

// GetCookieManager retrieves the CookieManager instance from the current request's context.
// It returns the manager and a boolean indicating if the type assertion was successful.
// GetCookieManager 함수는 현재 요청의 컨텍스트에서 CookieManager 인스턴스를 추출하여 반환합니다.
// 인스턴스가 존재하고 타입이 올바르면 (*CookieManager, true)를, 그렇지 않으면 (nil, false)를 반환합니다.
func GetCookieManager(ctx context.Context) (*CookieManager, bool) {
	cm, ok := ctx.Value(cookieContextKey{}).(*CookieManager)
	return cm, ok
}

// CookieManager holds the secret key for signing and provides cookie manipulation functions.
// CookieManager는 쿠키 조작 기능과 보안 서명을 위한 비밀키를 보유합니다.
type CookieManager struct {
	// SecretKey is used for HMAC signing.
	// SecretKey는 HMAC 서명에 사용할 비밀키입니다.
	SecretKey []byte
}

// sign generates an HMAC-SHA256 signature for the given value and returns it as a base64 URL-encoded string.
// sign은 주어진 값(value)을 HMAC-SHA256으로 서명한 후, base64 URL 인코딩된 문자열로 반환합니다.
func (cm *CookieManager) sign(value string) string {
	h := hmac.New(sha256.New, cm.SecretKey)
	h.Write([]byte(value))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// SetCookie creates a signed cookie with the specified name, value, and maxAge, and sets it in the HTTP response.
// The cookie value is stored in the format "base64-encoded-value|signature".
// SetCookie는 지정한 이름(name), 값(value), 유효기간(maxAge)을 갖는 서명된 쿠키를 생성하여 HTTP 응답(response)에 설정합니다.
// 쿠키 값은 "base64로 인코딩된 값|서명" 형식으로 저장됩니다.
func (cm *CookieManager) SetCookie(w http.ResponseWriter, name, value string, maxAge int) {
	signature := cm.sign(value)
	encodedValue := base64.URLEncoding.EncodeToString([]byte(value))
	finalValue := encodedValue + "|" + signature

	cookie := &http.Cookie{
		Name:     name,
		Value:    finalValue,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
		MaxAge:   maxAge,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

// ReadCookie reads a cookie by name from the request, verifies its signature, and returns the original untampered value.
// It expects the cookie value to be in the "encodedValue|signature" format.
// ReadCookie는 요청(request)으로부터 지정한 이름(name)의 쿠키를 읽어, 서명을 확인하고 위변조되지 않은 원본 값을 반환합니다.
// 쿠키 값은 "encodedValue|signature" 형식이어야 합니다.
func (cm *CookieManager) ReadCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}

	// Use strings.Cut for safe splitting, even if the pipe character is in the value.
	// strings.Cut을 사용하여 파이프(|) 문자가 값에 포함되어도 안전하게 분리합니다.
	encodedValue, signature, valid := strings.Cut(cookie.Value, "|")
	if !valid {
		return ""
	}

	data, err := base64.URLEncoding.DecodeString(encodedValue)
	if err != nil {
		return ""
	}
	value := string(data)

	// Verify the signature.
	// 서명을 검증합니다.
	expectedSignature := cm.sign(value)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return ""
	}

	return value
}

// DelCookie deletes a cookie by name by setting its expiration time to the past.
// DelCookie는 지정한 이름(name)의 쿠키를 삭제하기 위해, 만료 시간을 과거로 설정하여 응답에 등록합니다.
func (cm *CookieManager) DelCookie(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	}
	http.SetCookie(w, cookie)
}

// ReadFlash implements flash cookie functionality. A flash cookie is read once and then automatically deleted.
// It's useful for one-time messages like notifications.
// ReadFlash는 플래시 쿠키 기능을 구현합니다. 플래시 쿠키는 한 번 읽은 후 자동으로 삭제되어 일회성 메시지(알림 등) 처리에 유용합니다.
func (cm *CookieManager) ReadFlash(w http.ResponseWriter, r *http.Request, name string) string {
	msg := cm.ReadCookie(r, name)
	cm.DelCookie(w, name)
	return msg
}