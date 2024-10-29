package lsego

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/helin0815/gowebmonitor/pkg/log"
	"github.com/helin0815/gowebmonitor/pkg/tools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag"
)

var internalPaths = []string{
	"/metrics",
	"/healthz",
	"/dayu/prometheus",
	"/dayu/kube/healthz/alive",
	"/dayu/kube/healthz/ready",
	"/dayu/graceful-shutdown",
}

func requestShouldSkip(r *http.Request) bool {
	for _, ignore := range internalPaths {
		if strings.HasPrefix(r.URL.Path, ignore) {
			return true
		}
	}
	return false
}

func (lg *Lsego) Handle(pattern string, handler http.Handler) {
	log.Printf("Any %-25s --> %s", pattern, tools.NameOfFunction(handler))
	http.Handle(pattern, handler)
}
func (lg *Lsego) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if handler == nil {
		panic("http: nil handler")
	}

	lg.Handle(pattern, http.HandlerFunc(handler))
}

// Expose the registered metrics via HTTP.
func (lg *Lsego) setupMetricsHandler() {
	lg.Use(metricMiddleware)
	handler := promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	)
	lg.Handle("/metrics", handler)
	lg.Handle("/dayu/prometheus", handler)
}

func (lg *Lsego) SetupSwagger(pattern string, root fs.FS) {
	lg.Handle(pattern, http.StripPrefix(pattern, http.FileServer(http.FS(root))))
	swaggerHandler := httpSwagger.Handler(httpSwagger.URL(defaultSwaggerURL(pattern, root)))
	if _, err := swag.ReadDoc(); err == nil {
		swaggerHandler = httpSwagger.Handler()
	}

	lg.Handle("/swagger/", swaggerHandler)
}

func (lg *Lsego) setupLifecycleHandler() {
	lg.HandleFunc("/healthz", healthz)
	lg.HandleFunc("/dayu/kube/healthz/alive", healthz)
	lg.HandleFunc("/dayu/kube/healthz/ready", healthz)
	lg.HandleFunc("/dayu/graceful-shutdown", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received graceful shutdown request: %s %s, Headers: %v", r.Method, r.RequestURI, r.Header)
		for _, f := range lg.onShutdown {
			log.Printf("Calling shutdown function: %s", tools.NameOfFunction(f))
			f()
		}
		log.Printf("Call Shutdown function completed")
	})
}

func healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Ok!")
}

func defaultSwaggerURL(urlPrefix string, root fs.FS) string {
	defaultSwaggerFilename := "index.json"
	fs.WalkDir(root, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".yaml") {
			defaultSwaggerFilename = path
		}
		return nil
	})

	return fmt.Sprintf("%s%s", urlPrefix, defaultSwaggerFilename)
}
