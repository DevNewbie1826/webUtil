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
// ... (omitted for brevity)
```

### 2. Secure Cookie Management (`cookie`)

This middleware provides a `CookieManager` to handle signed cookies, preventing tampering.

**Example:**
```go
// ... (omitted for brevity)
```

### 3. Security Headers (`secure`)

These middlewares add various security headers to every response.

**Example:**
```go
// ... (omitted for brevity)
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
	"os"

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
// ... (omitted for brevity)
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
// ... (생략)
```

### 2. 안전한 쿠키 관리 (`cookie`)

이 미들웨어는 쿠키의 위변조를 방지하기 위해 서명된 쿠키를 다루는 `CookieManager`를 제공합니다.

**예제:**
```go
// ... (생략)
```

### 3. 보안 헤더 (`secure`)

이 미들웨어들은 모든 응답에 다양한 보안 헤더를 추가합니다.

**예제:**
```go
// ... (생략)
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
	"os"

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
// ... (생략)
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 라이선스

이 프로젝트는 MIT 라이선스에 따라 라이선스가 부여됩니다. 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하십시오.