package compress

import (
	"compress/gzip"
	"log"
	"net/http"

	"github.com/klauspost/compress/gzhttp"
)

// GzipMiddleware creates a gzip compression middleware with the given options (minSize, level, types).
// GzipMiddleware는 주어진 옵션(minSize, level, types)을 사용하여 gzip 압축 미들웨어를 생성합니다.
func GzipMiddleware(minSize int, level int, types []string) func(http.Handler) http.Handler {
	// Validate the compression level, and if it's out of range, set it to the default.
	// 압축 레벨이 유효한 범위 내에 있는지 확인하고, 그렇지 않으면 기본값으로 설정합니다.
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	// Create a defensive copy to prevent modification of the slice from outside.
	// 외부에서 슬라이스가 수정되는 것을 방지하기 위해 방어적 복사본을 생성합니다.
	var compressibleTypes []string
	if len(types) > 0 {
		compressibleTypes = make([]string, len(types))
		copy(compressibleTypes, types)
	} else {
		compressibleTypes = make([]string, len(DefaultCompressibleContentTypes))
		copy(compressibleTypes, DefaultCompressibleContentTypes)
	}

	// Create a new wrapper function: func(http.Handler) http.HandlerFunc.
	// gzhttp.NewWrapper: func(http.Handler) http.HandlerFunc 생성
	wrapperFunc, err := gzhttp.NewWrapper(
		gzhttp.MinSize(minSize),
		gzhttp.CompressionLevel(level),
		gzhttp.ContentTypes(compressibleTypes),
	)
	if err != nil {
		// If initialization fails, log the error instead of panicking and return a middleware that does not perform compression.
		// 초기화 실패 시 패닉 대신 오류를 로깅하고, 압축을 수행하지 않는 미들웨어를 반환합니다.
		log.Printf("GzipMiddleware initialization failed: %v", err)
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Adapter to match the http.Handler interface.
	// http.Handler 인터페이스에 맞추기 위한 어댑터
	return func(next http.Handler) http.Handler {
		return wrapperFunc(next)
	}
}