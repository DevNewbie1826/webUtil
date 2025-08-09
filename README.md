# webUtil (English)

A collection of useful and secure middleware for Go (Golang) web applications.

## Features

- **Gzip Compression**: Middleware to compress HTTP responses with gzip.
- **Secure Cookie Management**: Middleware for creating and reading signed (HMAC-SHA256) and secure cookies.
- **Security Headers**: Middlewares to add important security headers like Content-Security-Policy (with nonce), HSTS, and CORS.

## Installation

```bash
go get github.com/DevNewbie1826/webUtil
```

---

## Usage

### 1. Gzip Compression (`compress`)

This middleware compresses the response body for specified content types. If no types are provided, a default list of compressible types is used.

**Example:**
```go
package main

import (
	"fmt"
	"net/http"
	"github.com/DevNewbie1826/webUtil/compress"
)

func main() {
	mux := http.NewServeMux()

	// A handler that writes a long string
	longTextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "This is a very long text that should be compressed. " +
			"This is a very long text that should be compressed. " +
			"This is a very long text that should be compressed.")
	})

	// Use the Gzip middleware with default settings
	mux.Handle("/", compress.GzipMiddleware(10, 5, nil)(longTextHandler))

	fmt.Println("Server starting at :8080")
	http.ListenAndServe(":8080", mux)
}
```

### 2. Secure Cookie Management (`cookie`)

This middleware provides a `CookieManager` to handle signed cookies, preventing tampering.

**Example:**
```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"github.com/DevNewbie1826/webUtil/cookie"
)

func main() {
	// A strong, secret key for signing cookies. Should be kept secret.
	secretKey := []byte("a-very-strong-and-secret-key")

	// Create the cookie middleware
	cookieMiddleware := cookie.Middleware(secretKey)

	mux := http.NewServeMux()

	// Handler to set a cookie
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
	http.ListenAndServe(":8080", cookieMiddleware(mux))
}
```

### 3. Security Headers (`secure`)

These middlewares add various security headers to every response.

**Example:**
```go
package main

import (
	"fmt"
	"net/http"
	"github.com/DevNewbie1826/webUtil/secure"
)

func main() {
	mux := http.NewServeMux()

	// Simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the nonce and use it in your HTML template
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
	cspConfig := secure.CSPConfig{
		DefaultSrc: []string{"'self'"},
		ScriptSrc:  []string{"'self'", "https://trusted.cdn.com"},
		StyleSrc:   []string{"'self'", "'unsafe-inline'"},
	}

	// Chain the security middlewares
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
	"github.com/DevNewbie1826/webUtil/compress"
	"github.com/DevNewbie1826/webUtil/cookie"
	"github.com/DevNewbie1826/webUtil/secure"
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
go get github.com/DevNewbie1826/webUtil
```

---

## 사용법

### 1. Gzip 압축 (`compress`)

이 미들웨어는 지정된 콘텐츠 타입에 대해 응답 본문을 압축합니다. 만약 타입을 지정하지 않으면, 기본적으로 압축 가능한 타입 목록을 사용합니다.

**예제:**
```go
package main

import (
	"fmt"
	"net/http"
	"github.com/DevNewbie1826/webUtil/compress"
)

func main() {
	mux := http.NewServeMux()

	// 긴 문자열을 응답으로 보내는 핸들러
	longTextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "이것은 압축되어야 하는 매우 긴 텍스트입니다. " +
			"이것은 압축되어야 하는 매우 긴 텍스트입니다. " +
			"이것은 압축되어야 하는 매우 긴 텍스트입니다.")
	})

	// 기본 설정을 사용하여 Gzip 미들웨어 적용
	mux.Handle("/", compress.GzipMiddleware(10, 5, nil)(longTextHandler))

	fmt.Println("서버가 :8080 포트에서 시작됩니다.")
	http.ListenAndServe(":8080", mux)
}
```

### 2. 안전한 쿠키 관리 (`cookie`)

이 미들웨어는 쿠키의 위변조를 방지하기 위해 서명된 쿠키를 다루는 `CookieManager`를 제공합니다.

**예제:**
```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"github.com/DevNewbie1826/webUtil/cookie"
)

func main() {
	// 쿠키 서명을 위한 강력한 비밀 키. 외부에 노출되면 안 됩니다.
	secretKey := []byte("a-very-strong-and-secret-key")

	// 쿠키 미들웨어 생성
	cookieMiddleware := cookie.Middleware(secretKey)

	mux := http.NewServeMux()

	// 쿠키를 설정하는 핸들러
	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		cm, ok := cookie.GetCookieManager(r.Context())
		if !ok {
			http.Error(w, "CookieManager를 가져올 수 없습니다.", http.StatusInternalServerError)
			return
		}
		cm.SetCookie(w, "username", "john.doe", 3600) // 이름, 값, 유효시간(초)
		fmt.Fprintln(w, "'username' 쿠키가 설정되었습니다.")
	})

	// 쿠키를 읽는 핸들러
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		cm, ok := cookie.GetCookieManager(r.Context())
		if !ok {
			http.Error(w, "CookieManager를 가져올 수 없습니다.", http.StatusInternalServerError)
			return
		}
		username := cm.ReadCookie(r, "username")
		if username == "" {
			fmt.Fprintln(w, "'username' 쿠키를 찾을 수 없거나 유효하지 않습니다.")
			return
		}
		fmt.Fprintf(w, "안녕하세요, %s님!", username)
	})

	fmt.Println("서버가 :8080 포트에서 시작됩니다.")
	// 모든 핸들러에 미들웨어 적용
	http.ListenAndServe(":8080", cookieMiddleware(mux))
}
```

### 3. 보안 헤더 (`secure`)

이 미들웨어들은 모든 응답에 다양한 보안 헤더를 추가합니다.

**예제:**
```go
package main

import (
	"fmt"
	"net/http"
	"github.com/DevNewbie1826/webUtil/secure"
)

func main() {
	mux := http.NewServeMux()

	// 간단한 핸들러
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// nonce 값을 가져와서 HTML 템플릿에 사용
		nonce := secure.GetNonce(r.Context())
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>보안 앱</title>
</head>
<body>
    <h1>안녕하세요, 안전한 세상!</h1>
    <script nonce="%s">
        // 여기에 인라인 스크립트를 작성하세요
        console.log('이 스크립트는 nonce로 보호됩니다.');
    </script>
</body>
</html>`, nonce)
	})

	// Content-Security-Policy 설정
	cspConfig := secure.CSPConfig{
		DefaultSrc: []string{"'self'"},
		ScriptSrc:  []string{"'self'", "https://trusted.cdn.com"},
		StyleSrc:   []string{"'self'", "'unsafe-inline'"},
	}

	// 보안 미들웨어들을 순서대로 연결
	finalHandler := secure.CORSMiddleware(
		secure.SecurityHeaders(
			secure.NonceHeaders(cspConfig)(handler),
		),
	)

	mux.Handle("/", finalHandler)

	fmt.Println("서버가 :8080 포트에서 시작됩니다.")
	http.ListenAndServe(":8080", mux)
}
```

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
	"github.com/DevNewbie1826/webUtil/compress"
	"github.com/DevNewbie1826/webUtil/cookie"
	"github.com/DevNewbie1826/webUtil/secure"
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
