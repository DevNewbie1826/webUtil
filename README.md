# webUtil (English)

A collection of useful and secure middleware for Go (Golang) web applications.

## Features

- **Gzip Compression**: Middleware to compress HTTP responses with gzip.
- **Secure Cookie Management**: Middleware for creating and reading signed (HMAC-SHA256) and secure cookies.
- **Security Headers**: Middlewares to add important security headers like Content-Security-Policy (with nonce), HSTS, and CORS.
- **Secure File Server**: A secure and configurable handler for serving static files.

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
	"strings"

	"github.com/DevNewbie1826/webUtil/compress"
)

func main() {
	mux := http.NewServeMux()

	// This handler will be compressed.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		// Write a long string to ensure it's larger than the minSize for compression.
		fmt.Fprint(w, strings.Repeat("This will be compressed! ", 100))
	})

	// Default options: minSize=1024, level=DefaultCompression, default content types.
	// To customize, provide your own values, e.g., compress.GzipMiddleware(512, gzip.BestSpeed, []string{"text/plain", "application/json"})
	gzipMiddleware := compress.GzipMiddleware(1024, -1, nil)
	mux.Handle("/", gzipMiddleware(handler))

	fmt.Println("Server starting at :8080")
	fmt.Println("Test with: curl -H \"Accept-Encoding: gzip\" -v http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
```

### 2. Secure Cookie Management (`cookie`)

This middleware provides a `CookieManager` to handle signed cookies, preventing tampering.

**Example:**
```go
package main

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/DevNewbie1826/webUtil/cookie"
	"github.com/go-chi/chi/v5"
)

func main() {
	// Generate a secure, random secret key.
	// In a real application, this should be loaded from a secure configuration.
	secretKey := make([]byte, 32)
	if _, err := rand.Read(secretKey); err != nil {
		panic("Failed to generate secret key")
	}

	r := chi.NewRouter()

	// Apply the cookie middleware.
	r.Use(cookie.Middleware(secretKey))

	// Route to set a cookie.
	r.Get("/set", func(w http.ResponseWriter, r *http.Request) {
		cm := cookie.GetCookieManager(r.Context())
		cm.SetCookie(w, "myCookie", "hello world", 3600) // name, value, maxAge (seconds)
		fmt.Fprint(w, "Cookie 'myCookie' has been set!")
	})

	// Route to read a cookie.
	r.Get("/read", func(w http.ResponseWriter, r *http.Request) {
		cm := cookie.GetCookieManager(r.Context())
		value := cm.ReadCookie(r, "myCookie")
		if value == "" {
			fmt.Fprint(w, "Cookie 'myCookie' not found or is invalid.")
			return
		}
		fmt.Fprintf(w, "The value of 'myCookie' is: %s", value)
	})

	// Route to delete a cookie.
	r.Get("/delete", func(w http.ResponseWriter, r *http.Request) {
		cm := cookie.GetCookieManager(r.Context())
		cm.DelCookie(w, "myCookie")
		fmt.Fprint(w, "Cookie 'myCookie' has been deleted.")
	})

	fmt.Println("Server starting at :8080")
	fmt.Println("Visit http://localhost:8080/set, then /read, then /delete")
	http.ListenAndServe(":8080", r)
}
```

### 3. Security Headers (`secure`)

These middlewares add various security headers to every response. `NonceHeaders` is particularly useful for a strict Content-Security-Policy (CSP).

**Example:**
```go
package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/DevNewbie1826/webUtil/secure"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	// --- Middleware Setup ---
	// 1. CSP with Nonce
	cspConfig := secure.CSPConfig{
		DefaultSrc: []string{"'self'"},
		// Add other directives as needed. The nonce will be added automatically
		// to default-src, script-src, and style-src.
	}
	r.Use(secure.NonceHeaders(cspConfig))

	// 2. Other standard security headers
	r.Use(secure.SecurityHeaders)

	// --- Route ---
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// The nonce is automatically added to the response header.
		// We can retrieve it to use in our HTML template.
		nonce := secure.GetNonce(r.Context())

		// Use the nonce in an inline script tag.
		// This script will be allowed by the CSP, while others will be blocked.
		html := `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Secure App</title>
			</head>
			<body>
				<h1>Hello with Nonce!</h1>
				<p>Check your browser's developer tools to see the Content-Security-Policy header.</p>
				<script nonce="{{.Nonce}}">
					console.log("This inline script is allowed because of the nonce!");
				</script>
			</body>
			</html>
		`
		tmpl, _ := template.New("index").Parse(html)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, map[string]string{"Nonce": nonce})
	})

	fmt.Println("Server starting at :8080")
	fmt.Println("Visit http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
```

### 4. Secure File Server (`fileserver`)

Provides a secure and efficient handler for serving static files from a filesystem. It prevents directory listing, serves `index.html` for directory requests, and allows for configurable browser caching.

**Example:**
This example serves files from a `./public` directory at the `/static` URL path.

First, create some files to serve:
```bash
mkdir -p public
echo '<h1>Hello from index.html!</h1>' > public/index.html
echo 'console.log("Hello from app.js!");' > public/app.js
```

Then, use the `fileserver.Run` function with the `chi` router:
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/DevNewbie1826/webUtil/fileserver"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	// --- File Server Setup ---
	// The path on your URL where files will be served (e.g., /static/app.js)
	urlPath := "/static"

	// The directory on your filesystem to serve files from.
	fileSystemRoot := http.Dir("./public")

	// An optional prefix to add to the path inside the filesystem.
	// For this example, we leave it empty.
	stripPrefix := ""

	// Cache duration in seconds.
	// > 0: Enables caching. (e.g., 3600 for 1 hour)
	// == 0: Disables caching.
	// < 0: Sets Cache-Control: no-store.
	cacheDuration := 3600

	// Register the file server.
	fileserver.Run(r, urlPath, fileSystemRoot, stripPrefix, cacheDuration)

	// --- Example Route to link to the static files ---
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `
			<h1>File Server Example</h1>
			<p>Check out the static files:</p>
			<ul>
				<li><a href="/static/">/static/ (serves index.html)</a></li>
				<li><a href="/static/app.js">/static/app.js</a></li>
			</ul>
			<p>Open your browser's developer tools to inspect the 'Cache-Control' header on the .js file.</p>
		`)
	})

	fmt.Println("Chi server starting at :8080")
	fmt.Println("Serving files from './public' at 'http://localhost:8080/static'")
	http.ListenAndServe(":8080", r)
}
```

### 5. Full Example with Chi Router

Here is an example of how to use all middlewares together with the popular `chi` router.

**Installation for Chi:**
```bash
go get github.com/go-chi/chi/v5
```

**Example:**
```go
package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"net/http"

	"github.com/DevNewbie1826/webUtil/compress"
	"github.com/DevNewbie1826/webUtil/cookie"
	"github.com/DevNewbie1826/webUtil/fileserver"
	"github.com/DevNewbie1826/webUtil/secure"
	"github.com/go-chi/chi/v5"
)

func main() {
	// --- Key Generation ---
	secretKey := make([]byte, 32)
	if _, err := rand.Read(secretKey); err != nil {
		panic("Failed to generate secret key")
	}

	r := chi.NewRouter()

	// --- Middleware Setup (Order Matters) ---
	// 1. Gzip compression for responses
	r.Use(compress.GzipMiddleware(1024, -1, nil))

	// 2. CSP with Nonce
	cspConfig := secure.CSPConfig{
		DefaultSrc: []string{"'self'"},
		StyleSrc:   []string{"'self'", "https://cdn.jsdelivr.net"}, // Allow external styles
	}
	r.Use(secure.NonceHeaders(cspConfig))

	// 3. Other standard security headers
	r.Use(secure.SecurityHeaders)

	// 4. Secure cookie management
	r.Use(cookie.Middleware(secretKey))

	// --- Static File Server ---
	// Create dummy files first
	_ = http.Dir("./public").MkdirAll("css", 0755)
	_ = http.Dir("./public").WriteFile("index.html", []byte("<h1>Static Index</h1>"), 0644)
	_ = http.Dir("./public").WriteFile("css/style.css", []byte("body { font-family: sans-serif; }"), 0644)

	fileserver.Run(r, "/static", http.Dir("./public"), "", 3600)

	// --- Routes ---
	r.Get("/", rootHandler)
	r.Get("/set-cookie", setCookieHandler)
	r.Get("/read-cookie", readCookieHandler)

	fmt.Println("Full server starting at :8080")
	fmt.Println("Access at http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	nonce := secure.GetNonce(r.Context())
	cm := cookie.GetCookieManager(r.Context())
	flashMessage := cm.ReadFlash(w, r, "flash") // Read and delete flash message

	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Full Web App</title>
			<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/water.css@2/out/water.css">
			<link rel="stylesheet" href="/static/css/style.css">
		</head>
		<body>
			<h1>Welcome!</h1>
			{{if .FlashMessage}}
			<p style="color: green;"><strong>Message:</strong> {{.FlashMessage}}</p>
			{{end}}
			<p>This is a full example using all middlewares.</p>
			<nav>
				<a href="/set-cookie">Set a flash cookie</a> | 
				<a href="/read-cookie">Read the cookie (it will be gone)</a> |
				<a href="/static/">Visit static files</a>
			</nav>
			<script nonce="{{.Nonce}}">
				console.log("This script runs thanks to the CSP nonce!");
			</script>
		</body>
		</html>
	`
	tmpl, _ := template.New("root").Parse(html)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, map[string]string{
		"Nonce":        nonce,
		"FlashMessage": flashMessage,
	})
}

func setCookieHandler(w http.ResponseWriter, r *http.Request) {
	cm := cookie.GetCookieManager(r.Context())
	cm.SetCookie(w, "flash", "Hello from the flash message!", 60)
	http.Redirect(w, r, "/", http.StatusFound)
}

func readCookieHandler(w http.ResponseWriter, r *http.Request) {
	// The flash message is read on the root page, so by the time you navigate
	// here, it will already be gone. This redirect demonstrates that.
	http.Redirect(w, r, "/", http.StatusFound)
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
- **안전한 파일 서버**: 정적 파일을 제공하기 위한 안전하고 설정 가능한 핸들러입니다.

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
	"strings"

	"github.com/DevNewbie1826/webUtil/compress"
)

func main() {
	mux := http.NewServeMux()

	// 이 핸들러는 압축됩니다.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		// 압축을 위한 최소 크기보다 큰 응답을 위해 긴 문자열을 작성합니다.
		fmt.Fprint(w, strings.Repeat("이 내용은 압축될 것입니다! ", 100))
	})

	// 기본 옵션: minSize=1024, level=DefaultCompression, 기본 콘텐츠 타입.
	// 사용자 정의하려면, compress.GzipMiddleware(512, gzip.BestSpeed, []string{"text/plain", "application/json"}) 와 같이 값을 제공하세요.
	gzipMiddleware := compress.GzipMiddleware(1024, -1, nil)
	mux.Handle("/", gzipMiddleware(handler))

	fmt.Println("서버가 :8080 포트에서 시작됩니다.")
	fmt.Println("테스트: curl -H \"Accept-Encoding: gzip\" -v http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
```

### 2. 안전한 쿠키 관리 (`cookie`)

이 미들웨어는 쿠키의 위변조를 방지하기 위해 서명된 쿠키를 다루는 `CookieManager`를 제공합니다.

**예제:**
```go
package main

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/DevNewbie1826/webUtil/cookie"
	"github.com/go-chi/chi/v5"
)

func main() {
	// 안전한 랜덤 비밀키를 생성합니다.
	// 실제 애플리케이션에서는 보안 설정 파일 등에서 불러와야 합니다.
	secretKey := make([]byte, 32)
	if _, err := rand.Read(secretKey); err != nil {
		panic("비밀키 생성 실패")
	}

	r := chi.NewRouter()

	// 쿠키 미들웨어를 적용합니다.
	r.Use(cookie.Middleware(secretKey))

	// 쿠키를 설정하는 라우트.
	r.Get("/set", func(w http.ResponseWriter, r *http.Request) {
		cm := cookie.GetCookieManager(r.Context())
		cm.SetCookie(w, "myCookie", "안녕하세요", 3600) // 이름, 값, 유효시간(초)
		fmt.Fprint(w, "'myCookie' 쿠키가 설정되었습니다!")
	})

	// 쿠키를 읽는 라우트.
	r.Get("/read", func(w http.ResponseWriter, r *http.Request) {
		cm := cookie.GetCookieManager(r.Context())
		value := cm.ReadCookie(r, "myCookie")
		if value == "" {
			fmt.Fprint(w, "'myCookie' 쿠키를 찾을 수 없거나 유효하지 않습니다.")
			return
		}
		fmt.Fprintf(w, "'myCookie'의 값은: %s", value)
	})

	// 쿠키를 삭제하는 라우트.
	r.Get("/delete", func(w http.ResponseWriter, r *http.Request) {
		cm := cookie.GetCookieManager(r.Context())
		cm.DelCookie(w, "myCookie")
		fmt.Fprint(w, "'myCookie' 쿠키가 삭제되었습니다.")
	})

	fmt.Println("서버가 :8080 포트에서 시작됩니다.")
	fmt.Println("http://localhost:8080/set, 그 다음 /read, /delete 순으로 방문해보세요.")
	http.ListenAndServe(":8080", r)
}
```

### 3. 보안 헤더 (`secure`)

이 미들웨어들은 모든 응답에 다양한 보안 헤더를 추가합니다. `NonceHeaders`는 특히 엄격한 CSP(Content-Security-Policy) 설정에 유용합니다.

**예제:**
```go
package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/DevNewbie1826/webUtil/secure"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	// --- 미들웨어 설정 ---
	// 1. Nonce를 사용한 CSP
	cspConfig := secure.CSPConfig{
		DefaultSrc: []string{"'self'"},
		// 필요에 따라 다른 지시문을 추가하세요. Nonce는 default-src, script-src,
		// style-src에 자동으로 추가됩니다.
	}
	r.Use(secure.NonceHeaders(cspConfig))

	// 2. 기타 표준 보안 헤더
	r.Use(secure.SecurityHeaders)

	// --- 라우트 ---
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Nonce는 응답 헤더에 자동으로 추가됩니다.
		// HTML 템플릿에서 사용하기 위해 값을 가져올 수 있습니다.
		nonce := secure.GetNonce(r.Context())

		// 인라인 스크립트 태그에서 nonce를 사용합니다.
		// 이 스크립트는 CSP에 의해 허용되지만, 다른 스크립트들은 차단됩니다.
		html := `
			<!DOCTYPE html>
			<html>
			<head>
				<title>보안 앱</title>
			</head>
			<body>
				<h1>Nonce와 함께!</h1>
				<p>브라우저 개발자 도구에서 Content-Security-Policy 헤더를 확인해보세요.</p>
				<script nonce="{{.Nonce}}">
					console.log("이 인라인 스크립트는 Nonce 덕분에 허용됩니다!");
				</script>
			</body>
			</html>
		`
		tmpl, _ := template.New("index").Parse(html)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, map[string]string{"Nonce": nonce})
	})

	fmt.Println("서버가 :8080 포트에서 시작됩니다.")
	fmt.Println("http://localhost:8080 에서 확인하세요.")
	http.ListenAndServe(":8080", r)
}
```

### 4. 안전한 파일 서버 (`fileserver`)

파일 시스템의 정적 파일을 안전하고 효율적으로 제공하는 핸들러입니다. 디렉토리 리스팅을 방지하고, 디렉토리 요청 시 `index.html`을 제공하며, 브라우저 캐싱을 설정할 수 있습니다.

**예제:**
이 예제는 `./public` 디렉토리의 파일들을 `/static` URL 경로로 제공합니다.

먼저, 제공할 파일들을 생성합니다:
```bash
mkdir -p public
echo '<h1>안녕하세요, index.html입니다!</h1>' > public/index.html
echo 'console.log("app.js에서 보냅니다!");' > public/app.js
```

그 다음, `chi` 라우터와 함께 `fileserver.Run` 함수를 사용합니다:
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/DevNewbie1826/webUtil/fileserver"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	// --- 파일 서버 설정 ---
	// 파일이 제공될 URL 경로 (예: /static/app.js)
	urlPath := "/static"

	// 파일을 제공할 파일 시스템 상의 디렉토리
	fileSystemRoot := http.Dir("./public")

	// 파일 시스템 내부 경로에 추가할 접두사.
	// 이 예제에서는 비워둡니다.
	stripPrefix := ""

	// 캐시 기간(초).
	// > 0: 캐싱 활성화 (예: 3600은 1시간)
	// == 0: 캐싱 비활성화
	// < 0: Cache-Control: no-store 설정
	cacheDuration := 3600

	// 파일 서버 등록
	fileserver.Run(r, urlPath, fileSystemRoot, stripPrefix, cacheDuration)

	// --- 정적 파일로 링크를 거는 예제 라우트 ---
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `
			<h1>파일 서버 예제</h1>
			<p>아래 정적 파일들을 확인해보세요:</p>
			<ul>
				<li><a href="/static/">/static/ (index.html을 제공)</a></li>
				<li><a href="/static/app.js">/static/app.js</a></li>
			</ul>
			<p>브라우저 개발자 도구를 열어 .js 파일의 'Cache-Control' 헤더를 확인해보세요.</p>
		`)
	})

	fmt.Println("Chi 서버가 :8080 포트에서 시작됩니다.")
	fmt.Println("'./public' 디렉토리의 파일들을 'http://localhost:8080/static' 에서 제공합니다.")
	http.ListenAndServe(":8080", r)
}
```

### 5. Chi 라우터 전체 예제

인기 있는 `chi` 라우터와 모든 미들웨어를 함께 사용하는 예제입니다.

**Chi 설치:**
```bash
go get github.com/go-chi/chi/v5
```

**예제:**
```go
package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"net/http"

	"github.com/DevNewbie1826/webUtil/compress"
	"github.com/DevNewbie1826/webUtil/cookie"
	"github.com/DevNewbie1826/webUtil/fileserver"
	"github.com/DevNewbie1826/webUtil/secure"
	"github.com/go-chi/chi/v5"
)

func main() {
	// --- 키 생성 ---
	secretKey := make([]byte, 32)
	if _, err := rand.Read(secretKey); err != nil {
		panic("비밀키 생성 실패")
	}

	r := chi.NewRouter()

	// --- 미들웨어 설정 (순서가 중요합니다) ---
	// 1. 응답 Gzip 압축
	r.Use(compress.GzipMiddleware(1024, -1, nil))

	// 2. Nonce를 사용한 CSP
	cspConfig := secure.CSPConfig{
		DefaultSrc: []string{"'self'"},
		StyleSrc:   []string{"'self'", "https://cdn.jsdelivr.net"}, // 외부 스타일시트 허용
	}
	r.Use(secure.NonceHeaders(cspConfig))

	// 3. 기타 표준 보안 헤더
	r.Use(secure.SecurityHeaders)

	// 4. 안전한 쿠키 관리
	r.Use(cookie.Middleware(secretKey))

	// --- 정적 파일 서버 ---
	// 먼저 더미 파일들을 생성합니다
	_ = http.Dir("./public").MkdirAll("css", 0755)
	_ = http.Dir("./public").WriteFile("index.html", []byte("<h1>정적 인덱스</h1>"), 0644)
	_ = http.Dir("./public").WriteFile("css/style.css", []byte("body { font-family: sans-serif; }"), 0644)

	fileserver.Run(r, "/static", http.Dir("./public"), "", 3600)

	// --- 라우트 ---
	r.Get("/", rootHandler)
	r.Get("/set-cookie", setCookieHandler)
	r.Get("/read-cookie", readCookieHandler)

	fmt.Println("전체 기능 서버가 :8080 포트에서 시작됩니다.")
	fmt.Println("http://localhost:8080 에서 접속하세요.")
	http.ListenAndServe(":8080", r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	nonce := secure.GetNonce(r.Context())
	cm := cookie.GetCookieManager(r.Context())
	flashMessage := cm.ReadFlash(w, r, "flash") // 플래시 메시지를 읽고 삭제합니다

	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>통합 웹 앱</title>
			<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/water.css@2/out/water.css">
			<link rel="stylesheet" href="/static/css/style.css">
		</head>
		<body>
			<h1>환영합니다!</h1>
			{{if .FlashMessage}}
			<p style="color: green;"><strong>메시지:</strong> {{.FlashMessage}}</p>
			{{end}}
			<p>이것은 모든 미들웨어를 사용한 전체 예제입니다.</p>
			<nav>
				<a href="/set-cookie">플래시 쿠키 설정하기</a> | 
				<a href="/read-cookie">쿠키 읽기 (사라집니다)</a> |
				<a href="/static/">정적 파일 방문하기</a>
			</nav>
			<script nonce="{{.Nonce}}">
				console.log("이 스크립트는 CSP Nonce 덕분에 실행됩니다!");
			</script>
		</body>
		</html>
	`
	tmpl, _ := template.New("root").Parse(html)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, map[string]string{
		"Nonce":        nonce,
		"FlashMessage": flashMessage,
	})
}

func setCookieHandler(w http.ResponseWriter, r *http.Request) {
	cm := cookie.GetCookieManager(r.Context())
	cm.SetCookie(w, "flash", "플래시 메시지가 도착했습니다!", 60)
	http.Redirect(w, r, "/", http.StatusFound)
}

func readCookieHandler(w http.ResponseWriter, r *http.Request) {
	// 플래시 메시지는 루트 페이지에서 읽히므로, 이 페이지로 이동했을 때는
	// 이미 사라진 상태입니다. 이 리다이렉트는 그것을 보여줍니다.
	http.Redirect(w, r, "/", http.StatusFound)
}
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 라이선스

이 프로젝트는 MIT 라이선스에 따라 라이선스가 부여됩니다. 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하십시오.
