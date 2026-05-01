package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	. "turkey-clock/server/ntputils"
	"net/url"
	"github.com/beevik/ntp"
)

type TimeResponse struct {
	CurrentTime interface{} `json:"current_time"`
	TimeZone    string      `json:"time_zone"`
	Timestamp   int64       `json:"timestamp"`
	DateTime    string      `json:"datetime"`
	NTPResponse interface{} `json:"ntp_response"`
}

type TimeData struct {
	Time    time.Time
	NTPData *ntp.Response
	server  string
}

type PageData struct {
	NTPData             interface{}
	NTPJSON             template.JS
	Script              template.JS
	Style               template.CSS
	Host                string
	BaseURL             string
	GOOGLE_ANALYTICS_ID string
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

var config Config
var logger *slog.Logger

func main() {
	config = LoadConfig(flag.CommandLine)
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: config.LogLevel}))
	http.HandleFunc("/time", getTimeHandler)
	http.HandleFunc("/current_time", getTimeHandler)
	http.HandleFunc("/get_current_time", getTimeHandler)
	http.HandleFunc("/", homeHandler)
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	logger.Info("server starting", "addr", addr, "ntp_server", config.NtpHost)
	logger.Error("Server stopped", "error", http.ListenAndServe(addr, logRequests(http.DefaultServeMux)))
}

func isSane(resp time.Time) bool {
	now := time.Now().UTC()
	diff := resp.Sub(now)
	return diff <= 24*time.Hour && diff >= -24*time.Hour
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		writer := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(writer, r)
		logger.Info(
			"Handled request",
			"ip", getClientIP(r),
			"request", "method", r.Method,
			"path", r.URL.RequestURI(),
			"status", writer.statusCode,
			"duration", time.Since(startedAt).Round(time.Millisecond),
		)
	})
}

func getNTPResponse(host string) (*ntp.Response, error) {
	logger.Debug("Querying NTP server", "host", host)
	response, err := ntp.Query(host)
	if err != nil {
		logger.Warn("NTP query error", "error", err, "server", host)
		return response, err
	}
	if !isSane(response.Time) {
		logger.Warn("Received an insane time from NTP server", "time", response.Time, "server", host)
		return response, fmt.Errorf("Received an insane time from NTP server: %s", response.Time)
	}
	return response, nil
}

func getNTPData() TimeData {
	response, err := getNTPResponse(config.NtpHost)
	if err != nil {
		logger.Info("Using Fallback NTP server", "server", config.NtpFallback)
		response, err = getNTPResponse(config.NtpFallback)
		if err != nil {
			logger.Warn("Fallback NTP query error", "error", err)
			return TimeData{time.Now().UTC(), nil, "local_time"}
		}
		logger.Debug("NTP response", "server", config.NtpFallback, "time", response.Time.String())
		return TimeData{response.Time.UTC(), response, config.NtpFallback}
	}
	logger.Debug("NTP response", "server", config.Ntpdomain, "time", response.Time.String())
	return TimeData{response.Time.UTC(), response, config.Ntpdomain}
}

func getTime(r *http.Request, q url.Values, precision_unit string) TimeResponse {
	data := getNTPData()
	displayTime := data.Time

	tzParam := q.Get("timezone")
	loc := time.UTC
	if tzParam != "" {
		if l, err := time.LoadLocation(tzParam); err == nil {
			loc = l
			displayTime = displayTime.In(loc)
		}
	}

	currentTimeVar := FormatTime(displayTime, q.Get("timestamp"), q.Get("format"))

	if data.NTPData == nil {
		return TimeResponse{
			CurrentTime: currentTimeVar,
			TimeZone:    loc.String(),
			Timestamp:   data.Time.Unix(),
			DateTime:    data.Time.Format(time.RFC3339),
			NTPResponse: interface{}(nil),
		}
	}

	return TimeResponse{
		CurrentTime: currentTimeVar,
		TimeZone:    loc.String(),
		Timestamp:   data.Time.Unix(),
		DateTime:    data.Time.Format(time.RFC3339),
		NTPResponse: map[string]interface{}{
			"time":            FormatTime(data.NTPData.Time, q.Get("timestamp"), q.Get("format")),
			"server":          data.server,
			"unit_time":       precision_unit,
			"offset":          FormatDuration(data.NTPData.ClockOffset, precision_unit),
			"precision":       FormatDuration(data.NTPData.Precision, "ns"),
			"root_dispersion": FormatDuration(data.NTPData.RootDispersion, precision_unit),
			"root_distance":   FormatDuration(data.NTPData.RootDistance, precision_unit),
			"rtt":             FormatDuration(data.NTPData.RTT, precision_unit),
			"stratum":         data.NTPData.Stratum,
		},
	}
}

func getTimeHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	precision_unit := q.Get("precision_unit")
	if precision_unit == "" {
		precision_unit = "ms"
	}
	data := getTime(r, q, precision_unit)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	baseURL := scheme + "://" + r.Host
	data := getTime(r, url.Values{}, "auto")

	script, err := os.ReadFile("assets/script.js")
	if err != nil {
		panic(fmt.Sprintf("Error in read file script.js: %v", err))
	}

	style, err := os.ReadFile("assets/style.css")
	if err != nil {
		panic(fmt.Sprintf("Error in read file style.css: %v", err))
	}

	ntpJSONBytes, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshaling initial NTP data:", err)
		ntpJSONBytes = []byte("{}")
	}

	page := PageData{
		NTPData:             data,
		NTPJSON:             template.JS(ntpJSONBytes),
		Script:              template.JS(script),
		Style:               template.CSS(style),
		Host:                r.Host,
		BaseURL:             baseURL,
		GOOGLE_ANALYTICS_ID: config.GoogleAnalytics,
	}

	tmpl, _ := template.ParseFiles("assets/index.html")
	tmpl.Execute(w, page)
}
