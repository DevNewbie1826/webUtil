package fileserver

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/DevNewbie1826/httperror"
	"github.com/go-chi/chi/v5"
)

// --- Custom FileSystem wrappers ---

// noListFileSystem wraps an http.FileSystem to prevent directory listing.
// If a directory is requested, it attempts to serve 'index.html' from that directory.
// If 'index.html' is not found, it returns a 'not found' error.
//
// noListFileSystem은 디렉토리 리스팅을 방지하기 위해 http.FileSystem을 래핑합니다.
// 디렉토리가 요청되면 해당 디렉토리의 'index.html'을 서빙하려고 시도합니다.
// 'index.html'을 찾지 못하면 'not found' 오류를 반환합니다.
type noListFileSystem struct {
	// fs is the underlying file system.
	// fs는 내부 파일 시스템입니다.
	fs http.FileSystem
}

// Open opens the named file. It overrides the default behavior for directories.
//
// Open은 지정된 이름의 파일을 엽니다. 디렉토리에 대한 기본 동작을 재정의합니다.
func (nlfs noListFileSystem) Open(name string) (http.File, error) {
	f, err := nlfs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	s, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if s.IsDir() {
		f.Close()
		return nil, os.ErrPermission
	}
	return f, nil
}

// prefixAddingFileSystem wraps an http.FileSystem to prepend a path prefix to every request.
// This is used to maintain backward compatibility with the old 'stripPrefix' parameter,
// which behaved as an "add-prefix".
//
// prefixAddingFileSystem은 모든 요청에 경로 접두사를 추가하기 위해 http.FileSystem을 래핑합니다.
// 이것은 "add-prefix"처럼 동작했던 이전 'stripPrefix' 파라미터와의 하위 호환성을 유지하기 위해 사용됩니다.
type prefixAddingFileSystem struct {
	// prefix is the path prefix to add to each request.
	// prefix는 각 요청에 추가할 경로 접두사입니다.
	prefix string
	// fs is the underlying file system.
	// fs는 내부 파일 시스템입니다.
	fs http.FileSystem
}

// Open opens the named file after prepending the configured prefix.
//
// Open은 설정된 접두사를 앞에 붙인 후 지정된 이름의 파일을 엽니다.
func (pfs prefixAddingFileSystem) Open(name string) (http.File, error) {
	prefixedName := path.Join(pfs.prefix, name)
	return pfs.fs.Open(prefixedName)
}

// --- Main Function ---

// Run sets up a handler on the given Chi router to serve static files.
// It allows for configurable caching and uses a custom 404 error handler.
//
// Run은 정적 파일을 제공하기 위해 주어진 Chi 라우터에 핸들러를 설정합니다.
// 캐싱을 설정할 수 있으며 사용자 정의 404 오류 핸들러를 사용합니다.
//
// Parameters:
//   - r: The Chi router.
//   - r: Chi 라우터입니다.
//   - urlPath: The URL path to serve files from (e.g., "/static").
//   - urlPath: 파일을 제공할 URL 경로입니다 (예: "/static").
//   - fs: The http.FileSystem to serve files from.
//   - fs: 파일을 제공할 http.FileSystem입니다.
//   - stripPrefix: A prefix to be added to the file path inside the http.FileSystem.
//   - stripPrefix: http.FileSystem 내부의 파일 경로에 추가될 접두사입니다.
//   - cacheMaxAgeSeconds: Caching duration in seconds.
//   - > 0: Sets "Cache-Control: public, max-age=<value>".
//   - < 0: Sets "Cache-Control: no-store".
//   - == 0: Caching is disabled (no header is set).
//   - cacheMaxAgeSeconds: 캐시 기간(초)입니다.
//   - > 0: "Cache-Control: public, max-age=<값>"을 설정합니다.
//   - < 0: "Cache-Control: no-store"를 설정합니다.
//   - == 0: 캐싱이 비활성화됩니다 (헤더가 설정되지 않음).
func Run(r *chi.Mux, urlPath string, fs http.FileSystem, stripPrefix string, cacheMaxAgeSeconds int) {
	// --- Input Validation ---
	if strings.ContainsAny(urlPath, "{}*") {
		panic(fmt.Sprintf("FileServer does not permit URL parameters in urlPath: %s", urlPath))
	}

	// --- Filesystem Setup ---
	var effectiveFs http.FileSystem = fs
	if stripPrefix != "" {
		effectiveFs = prefixAddingFileSystem{prefix: stripPrefix, fs: fs}
	}
	finalFs := noListFileSystem{fs: effectiveFs}

	// --- Handler Setup ---
	if urlPath != "/" && urlPath[len(urlPath)-1] != '/' {
		r.Get(urlPath, http.RedirectHandler(urlPath+"/", http.StatusMovedPermanently).ServeHTTP)
		urlPath += "/"
	}

	fileServer := http.FileServer(finalFs)

	handlerWithCustom404 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The optimization path using http.ServeFile has been removed
		// as it can conflict with middleware like gzip handlers, causing infinite loading issues.
		// All requests are now handled through the standard http.FileServer
		// to ensure middleware compatibility.
		//
		// http.ServeFile을 사용하는 최적화 경로는 gzip 핸들러와 같은 미들웨어와 충돌하여
		// 무한 로딩 문제를 일으킬 수 있으므로 제거되었습니다.
		// 이제 모든 요청은 미들웨어 호환성을 보장하기 위해 표준 http.FileServer를 통해 처리됩니다.

		// Check if the file exists and handle errors
		// 파일이 존재하는지 확인하고 오류 처리
		f, err := finalFs.Open(r.URL.Path)
		if err != nil {
			if os.IsPermission(err) {
				httperror.ReportForbidden(r)
			} else if os.IsNotExist(err) {
				httperror.ReportNotFound(r)
			} else {
				httperror.ReportInternalServerError(r)
			}
			return
		}
		f.Close()

		// Apply caching policy
		// 캐시 정책 적용
		if cacheMaxAgeSeconds > 0 {
			w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheMaxAgeSeconds))
		} else if cacheMaxAgeSeconds < 0 {
			w.Header().Set("Cache-Control", "no-store")
		}

		fileServer.ServeHTTP(w, r)
	})

	fsHandler := http.StripPrefix(urlPath, handlerWithCustom404)
	r.Get(urlPath+"*", fsHandler.ServeHTTP)
}
