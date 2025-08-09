package compress

// DefaultCompressibleContentTypes is a list of Content-Type values that are typically compressible.
// DefaultCompressibleContentTypes는 일반적으로 압축 가능한 Content-Type 값의 목록입니다.
var DefaultCompressibleContentTypes = []string{
	// Text-related types
	// 텍스트 관련 타입
	"text/html",
	"text/richtext",
	"text/plain",
	"text/css",
	"text/x-script",
	"text/x-component",
	"text/x-java-source",
	"text/x-markdown",
	"text/event-stream",

	// JavaScript-related types
	// 자바스크립트 관련 타입
	"application/javascript",
	"application/x-javascript",
	"text/javascript",
	"text/js",

	// Icon-related types
	// 아이콘 관련 타입
	"image/x-icon",
	"image/vnd.microsoft.icon",

	// CGI and Perl script types
	// CGI 및 Perl 스크립트 타입
	"application/x-perl",
	"application/x-httpd-cgi",

	// XML-related types
	// XML 관련 타입
	"text/xml",
	"application/xml",
	"application/rss+xml",

	// JSON and API-related types
	// JSON 및 API 관련 타입
	"application/vnd.api+json",
	"application/json",
	"application/manifest+json",
	"application/ld+json",
	"application/graphql+json",
	"application/geo+json",

	// Multipart-related types
	// Multipart 관련 타입
	"multipart/bag",
	"multipart/mixed",

	// XHTML-related types
	// XHTML 관련 타입
	"application/xhtml+xml",

	// Font-related types
	// 폰트 관련 타입
	"font/ttf",
	"font/otf",
	"font/x-woff",
	"image/svg+xml",
	"application/vnd.ms-fontobject",
	"application/ttf",
	"application/x-ttf",
	"application/otf",
	"application/x-otf",
	"application/truetype",
	"application/opentype",
	"application/x-opentype",
	"application/font-woff",
	"application/eot",
	"application/font",
	"application/font-sfnt",

	// Other related types
	// 기타 관련 타입
	"application/wasm",
	"application/javascript-binast",
}