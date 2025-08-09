# webUtil

A collection of useful and secure middleware for Go (Golang) web applications.

Go(Golang) 웹 애플리케이션을 위한 유용하고 안전한 미들웨어 모음입니다.

## Features

- **Gzip Compression**: Middleware to compress HTTP responses with gzip.
- **Secure Cookie Management**: Middleware for creating and reading signed (HMAC-SHA256) and secure cookies.
- **Security Headers**: Middlewares to add important security headers like Content-Security-Policy (with nonce), HSTS, and CORS.

- **Gzip 압축**: HTTP 응답을 gzip으로 압축하는 미들웨어입니다.
- **안전한 쿠키 관리**: 서명되고(HMAC-SHA256) 보안 설정이 적용된 쿠키를 생성하고 읽는 미들웨어입니다.
- **보안 헤더**: Content-Security-Policy(nonce 포함), HSTS, CORS 등 중요한 보안 관련 헤더를 추가하는 미들웨어입니다.

## Installation

```bash
go get github.com/your-username/webUtil
```
*(Please replace `your-username` with the actual path)*
*(실제 경로에 맞게 `your-username`을 수정해주세요)*

---

## Usage

### 1. Gzip Compression (`compress`)

This middleware compresses the response body for specified content types. If no types are provided, a default list of compressible types is used.

이 미들웨어는 지정된 콘텐츠 타입에 대해 응답 본문을 압축합니다. 만약 타입을 지정하지 않으면, 기본적으로 압축 가능한 타입 목록을 사용합니다.

**Example:**
```go
package main

import (
	"fmt"
	"net/http"
	"webUtil/compress"
)

func main() {
	mux := http.NewServeMux()

	// A handler that writes a long string
	// 긴 문자열을 응답으로 보내는 핸들러
	longTextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "This is a very long text that should be compressed. " +
			"This is a very long text that should be compressed. " +
			"This is a very long text that should be compressed.")
	})

	// Use the Gzip middleware with default settings
	// 기본 설정을 사용하여 Gzip 미들웨어 적용
	mux.Handle("/", compress.GzipMiddleware(10, 5, nil)(longTextHandler))

	fmt.Println("Server starting at :8080")
	http.ListenAndServe(":8080", mux)
}
```

### 2. Secure Cookie Management (`cookie`)

This middleware provides a `CookieManager` to handle signed cookies, preventing tampering.

이 미들웨어는 쿠키의 위변조를 방지하기 위해 서명된 쿠키를 다루는 `CookieManager`를 제공합니다.

**Example:**
```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"webUtil/cookie"
)

func main() {
	// A strong, secret key for signing cookies. Should be kept secret.
	// 쿠키 서명을 위한 강력한 비밀 키. 외부에 노출되면 안 됩니다.
	secretKey := []byte("a-very-strong-and-secret-key")

	// Create the cookie middleware
	// 쿠키 미들웨어 생성
	cookieMiddleware := cookie.Middleware(secretKey)

	mux := http.NewServeMux()

	// Handler to set a cookie
	// 쿠키를 설정하는 핸들러
	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		cm, ok := cookie.GetCookieManager(r.Context())
		if !ok {
			http.Error(w, "Could not get CookieManager", http.StatusInternalServerError)
			return
		}
		cm.SetCookie(w, "username", "john.doe", 3600) // name, value, maxAge (seconds)
		fmt.Fprintln(w, "Cookie 'username' has been set.")
	})

	// Handler to read a cookie
	// 쿠키를 읽는 핸들러
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		cm, ok := cookie.GetCookieManager(r.Context())
		if !ok {
			http.Error(w, "Could not get CookieManager", http.StatusInternalServerError)
			return
		}
		username := cm.ReadCookie(r, "username")
		if username == "" {
			fmt.Fprintln(w, "Cookie 'username' not found or invalid.")
			return
		}
		fmt.Fprintf(w, "Hello, %s!", username)
	})

	fmt.Println("Server starting at :8080")
	// Apply the middleware to all handlers
	// 모든 핸들러에 미들웨어 적용
	http.ListenAndServe(":8080", cookieMiddleware(mux))
}
```

### 3. Security Headers (`secure`)

These middlewares add various security headers to every response.

이 미들웨어들은 모든 응답에 다양한 보안 헤더를 추가합니다.

**Example:**
```go
package main

import (
	"fmt"
	"net/http"
	"webUtil/secure"
)

func main() {
	mux := http.NewServeMux()
	
	// Simple handler
	// 간단한 핸들러
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the nonce and use it in your HTML template
		// nonce 값을 가져와서 HTML 템플릿에 사용
		nonce := secure.GetNonce(r.Context())
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Secure App</title>
</head>
<body>
    <h1>Hello, Secure World!</h1>
    <script nonce="%s">
        // Your inline script here
        console.log('This script is protected by a nonce.');
    </script>
</body>
</html>`, nonce)
	})

	// Configure the Content-Security-Policy
	// Content-Security-Policy 설정
	cspConfig := secure.CSPConfig{
		DefaultSrc: []string{"'self'"},
		ScriptSrc:  []string{"'self'", "https://trusted.cdn.com"},
		StyleSrc:   []string{"'self'", "'unsafe-inline'"},
	}

	// Chain the security middlewares
	// 보안 미들웨어들을 순서대로 연결
	finalHandler := secure.CORSMiddleware(
		secure.SecurityHeaders(
			secure.NonceHeaders(cspConfig)(handler),
		),
	)

	mux.Handle("/", finalHandler)

	fmt.Println("Server starting at :8080")
	http.ListenAndServe(":8080", mux)
}
```
