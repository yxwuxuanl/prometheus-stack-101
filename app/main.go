package main

import (
	"cmp"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const MetricNamespace = "app"

var (
	requestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: MetricNamespace, // 指标名前缀
		Name:      "requests",      // 指标名
	}, []string{"status", "method"})
	inflightRequests = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: MetricNamespace,
		Name:      "inflight_requests",
	}, []string{"method"})
	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: MetricNamespace,
		Name:      "request_duration_seconds",
		Buckets:   []float64{.01, .025, .05, .1, .15, .2},
	}, []string{"method"})
)

var ran = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	http.HandleFunc("/{$}", func(rw http.ResponseWriter, r *http.Request) {
		// 正在处理请求数 + 1
		inflightRequests.With(prometheus.Labels{"method": r.Method}).Inc()

		// 记录请求开始时间
		start := time.Now()

		// 处理请求 & 返回响应
		status := handleRequest(r)
		rw.WriteHeader(status)

		// 计算 & 记录请求耗时
		dur := time.Since(start).Seconds()
		requestDuration.With(prometheus.Labels{"method": r.Method}).Observe(dur)

		// 完成请求数 + 1
		requestCounter.With(prometheus.Labels{
			"status": strconv.Itoa(status),
			"method": r.Method,
		}).Inc()

		// 正在处理请求数 - 1
		inflightRequests.With(prometheus.Labels{"method": r.Method}).Desc()
	})

	registry := prometheus.NewRegistry()

	// 注册指标
	registry.MustRegister(requestCounter, inflightRequests, requestDuration)

	// 暴露指标
	http.Handle(
		"GET /metrics",
		middleware.Logger(
			promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
		),
	)

	addr := cmp.Or(os.Getenv("SERVER_ADDR"), ":8000")
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleRequest(r *http.Request) int {
	delay := time.Duration(ran.Intn(100-10+1)+10) * time.Millisecond
	time.Sleep(delay)

	if n, _ := strconv.ParseFloat(r.URL.Query().Get("error_rate"), 64); n > 0 {
		if n > ran.Float64() {
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}
