# webUtil (English)

A collection of useful and secure middleware for Go (Golang) web applications.

## Features

- **Gzip Compression**: Middleware to compress HTTP responses with gzip.
- **Secure Cookie Management**: Middleware for creating and reading signed (HMAC-SHA256) and secure cookies.
- **Security Headers**: Middlewares to add important security headers like Content-Security-Policy (with nonce), HSTS, and CORS.

## Installation

```bash
go get github.com/your-username/webUtil
```
*(Please replace `your-username` with the actual path)*

---

## Usage

### 1. Gzip Compression (`compress`)
*(...omitted for brevity...)*

### 2. Secure Cookie Management (`cookie`)
*(...omitted for brevity...)*

### 3. Security Headers (`secure`)
*(...omitted for brevity...)*

### 4. Full Example with Chi Router

Here is an example of how to use all middlewares together with the popular `chi` router.

**Installation for Chi:**
```bash
go get github.com/go-chi/chi/v5
```

**Example:**
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"webUtil/compress"
	"webUtil/cookie"
	"webUtil/secure"
)

func main() {
	r := chi.NewRouter()

	// --- Middleware Setup ---
	secretKey := []byte("a-very-strong-and-secret-key-for-chi")
	cspConfig := secure.CSPConfig{
		DefaultSrc: []string{"'self'"},
		ScriptSrc:  []string{"'self'"},
		StyleSrc:   []string{"'self'"},
	}

	// Apply middlewares to the entire router
	r.Use(cookie.Middleware(secretKey))
	r.Use(secure.CORSMiddleware)
	r.Use(secure.SecurityHeaders)
	r.Use(secure.NonceHeaders(cspConfig))
	r.Use(compress.GzipMiddleware(20, 5, nil)) // minSize, level, types

	// --- Routes ---
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		nonce := secure.GetNonce(r.Context())
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
			<h1>Welcome!</h1>
			<p>This page is secure and compressed.</p>
			<p><a href="/set-cookie">Set a test cookie</a></p>
			<p><a href="/read-cookie">Read the cookie</a></p>
			<script nonce="%s">console.log("Hello from secure inline script!");</script>
		`, nonce)
	})

	r.Get("/set-cookie", func(w http.ResponseWriter, r *http.Request) {
		cm, ok := cookie.GetCookieManager(r.Context())
		if !ok {
			http.Error(w, "CookieManager not available", 500)
			return
		}
		cm.SetCookie(w, "my-test-cookie", "hello-chi", 600)
		fmt.Fprintln(w, "Cookie has been set!")
	})

	r.Get("/read-cookie", func(w http.ResponseWriter, r *http.Request) {
		cm, ok := cookie.GetCookieManager(r.Context())
		if !ok {
			http.Error(w, "CookieManager not available", 500)
			return
		}
		val := cm.ReadCookie(r, "my-test-cookie")
		if val == "" {
			fmt.Fprintln(w, "Cookie not found or is invalid.")
			return
		}
		fmt.Fprintf(w, "Cookie value is: %s", val)
	})

	fmt.Println("Chi server starting at :8080")
	http.ListenAndServe(":8080", r)
}
```

---
<br>

# webUtil (한국어)

Go(Golang) 웹 애플리케이션을 위한 유용하고 안전한 미들웨어 모음입니다.

## 기능

- **Gzip 압축**: HTTP 응답을 gzip으로 압축하는 미들웨어입니다.
- **안전한 쿠키 관리**: 서명되고(HMAC-SHA256) 보안 설정이 적용된 쿠키를 생성하고 읽는 미들웨어입니다.
- **보안 헤더**: Content-Security-Policy(nonce 포함), HSTS, CORS 등 중요한 보안 관련 헤더를 추가하는 미들웨어입니다.

## 설치

```bash
go get github.com/your-username/webUtil
```
*(실제 경로에 맞게 `your-username`을 수정해주세요)*

---

## 사용법

### 1. Gzip 압축 (`compress`)
*(...생략...)*

### 2. 안전한 쿠키 관리 (`cookie`)
*(...생략...)*

### 3. 보안 헤더 (`secure`)
*(...생략...)*

### 4. Chi 라우터 전체 예제

인기 있는 `chi` 라우터와 모든 미들웨어를 함께 사용하는 예제입니다.

**Chi 설치:**
```bash
go get github.com/go-chi/chi/v5
```

**예제:**
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"webUtil/compress"
	"webUtil/cookie"
	"webUtil/secure"
)

func main() {
	r := chi.NewRouter()

	// --- 미들웨어 설정 ---
	secretKey := []byte("a-very-strong-and-secret-key-for-chi")
	cspConfig := secure.CSPConfig{
		DefaultSrc: []string{"'self'"},
		ScriptSrc:  []string{"'self'"},
		StyleSrc:   []string{"'self'"},
	}

	// 전체 라우터에 미들웨어 적용
	r.Use(cookie.Middleware(secretKey))
	r.Use(secure.CORSMiddleware)
	r.Use(secure.SecurityHeaders)
	r.Use(secure.NonceHeaders(cspConfig))
	r.Use(compress.GzipMiddleware(20, 5, nil)) // 최소크기, 압축레벨, 타입

	// --- 라우트 ---
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		nonce := secure.GetNonce(r.Context())
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
			<h1>환영합니다!</h1>
			<p>이 페이지는 안전하고 압축되어 있습니다.</p>
			<p><a href="/set-cookie">테스트 쿠키 설정하기</a></p>
			<p><a href="/read-cookie">쿠키 읽기</a></p>
			<script nonce="%s">console.log("안전한 인라인 스크립트에서 보냅니다!");</script>
		`, nonce)
	})

	r.Get("/set-cookie", func(w http.ResponseWriter, r *http.Request) {
		cm, ok := cookie.GetCookieManager(r.Context())
		if !ok {
			http.Error(w, "CookieManager를 사용할 수 없습니다.", 500)
			return
		}
		cm.SetCookie(w, "my-test-cookie", "hello-chi", 600)
		fmt.Fprintln(w, "쿠키가 설정되었습니다!")
	})

	r.Get("/read-cookie", func(w http.ResponseWriter, r *http.Request) {
		cm, ok := cookie.GetCookieManager(r.Context())
		if !ok {
			http.Error(w, "CookieManager를 사용할 수 없습니다.", 500)
			return
		}
		val := cm.ReadCookie(r, "my-test-cookie")
		if val == "" {
			fmt.Fprintln(w, "쿠키를 찾을 수 없거나 유효하지 않습니다.")
			return
		}
		fmt.Fprintf(w, "쿠키 값: %s", val)
	})

	fmt.Println("Chi 서버가 :8080 포트에서 시작됩니다.")
	http.ListenAndServe(":8080", r)
}
```
*참고: 간결함을 위해 이전 예제들은 생략했습니다.*
