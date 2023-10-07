package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	appVersion string
	version    = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "version",
		Help: "Version information about this binary",
		ConstLabels: map[string]string{
			"version": appVersion,
		},
	})

	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of all HTTP requests",
	}, []string{"code", "method"})

	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of all HTTP requests",
	}, []string{"code", "handler", "method"})
)

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	version.Set(1)
	port := os.Getenv("PORT")
	bind := ":3000"
	if port != "" {
		bind = ":" + port
	}
	enableH2c := false
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagset.StringVar(&bind, "bind", bind, "The socket to bind to.")
	flagset.BoolVar(&enableH2c, "h2c", false, "Enable h2c (http/2 over tcp) protocol.")
	flagset.Parse(os.Args[1:])

	r := prometheus.NewRegistry()
	r.MustRegister(httpRequestsTotal)
	r.MustRegister(httpRequestDuration)
	r.MustRegister(version)

	foundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(hostname))
	})
	foundHandlerProcess := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		sum := 0
		for i := 0; i < 100000000; i++ {
			sum += 1
		}
		strSum := strconv.Itoa(sum)
		w.Write([]byte(strSum))
	})
	notfoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	internalErrorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	foundChain := promhttp.InstrumentHandlerDuration(
		httpRequestDuration.MustCurryWith(prometheus.Labels{"handler": "found"}),
		promhttp.InstrumentHandlerCounter(httpRequestsTotal, foundHandler),
	)
	foundChainProcess := promhttp.InstrumentHandlerDuration(
		httpRequestDuration.MustCurryWith(prometheus.Labels{"handler": "found"}),
		promhttp.InstrumentHandlerCounter(httpRequestsTotal, foundHandlerProcess),
	)

	mux := http.NewServeMux()
	mux.Handle("/", foundChain)
	mux.Handle("/health_check", foundChain)
	mux.Handle("/err", promhttp.InstrumentHandlerCounter(httpRequestsTotal, notfoundHandler))
	mux.Handle("/internal-err", promhttp.InstrumentHandlerCounter(httpRequestsTotal, internalErrorHandler))
	mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	mux.Handle("/process", foundChainProcess)

	var srv *http.Server
	if enableH2c {
		srv = &http.Server{Addr: bind, Handler: h2c.NewHandler(mux, &http2.Server{})}
	} else {
		srv = &http.Server{Addr: bind, Handler: mux}
	}

	log.Fatal(srv.ListenAndServe())
}
