package ntputils

import (
	"flag"
	"os"
	"strconv"
	"log/slog"
	"strings"
)

type Config struct {
	NtpHost         string
	Ntpdomain       string
	NtpFallback     string
	Host            string
	Port            int
	GoogleAnalytics string
	LogLevel        slog.Level
}

func getLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func loadDefaultConfig() Config {
	return Config{
		NtpHost:         "",
		Ntpdomain:       "turkey-clock.aecrypto.io",
		NtpFallback:     "pool.ntp.org",
		Host:            "0.0.0.0",
		Port:            8080,
		GoogleAnalytics: "",
		LogLevel:        slog.LevelInfo,
	}
}

func (c *Config) loadFromEnv() {
	if v := os.Getenv("NTP_HOST"); v != "" {
		c.NtpHost = v
	}
	if v := os.Getenv("NTP_DOMAIN"); v != "" {
		c.Ntpdomain = v
	}
	if v := os.Getenv("NTP_FALLBACK"); v != "" {
		c.NtpFallback = v
	}
	if v := os.Getenv("HOST"); v != "" {
		c.Host = v
	}
	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Port = n
		}
	}
	if v := os.Getenv("GA"); v != "" {
		c.GoogleAnalytics = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.LogLevel = getLogLevel(v)
	}
}

func loadFromFlags(c *Config, fs *flag.FlagSet) {
	fs.StringVar(&c.NtpHost, "ntp-host", c.NtpHost, "NTP SERVER HOST ip/domain:port, in case you have an NTP server on the host machine")
	fs.StringVar(&c.Ntpdomain, "ntp-domain", c.Ntpdomain, "NTP SERVER DOMAIN default turkey-clock.aecrypto.io")
	fs.StringVar(&c.NtpFallback, "ntp-fallback", c.NtpFallback, "NTP FALLBACK SERVER ADDRESS default pool.ntp.org")
	fs.StringVar(&c.Host, "host", c.Host, "Server host address default localhost")
	fs.IntVar(&c.Port, "port", c.Port, "Server port default 8080")
	fs.StringVar(&c.GoogleAnalytics, "ga", c.GoogleAnalytics, "Google Analytics ID")
	log_level := fs.String("log-level", c.LogLevel.String(), "Log level (debug, info, warn, error) default info")
	_ = fs.Parse(os.Args[1:])
	if log_level != nil {
		c.LogLevel = getLogLevel(*log_level)
	}
}

func LoadConfig(fs *flag.FlagSet) Config {
	config := loadDefaultConfig()
	config.loadFromEnv()
	loadFromFlags(&config, fs)
	if config.NtpHost == "" {
		config.NtpHost = config.Ntpdomain
	}
	return config
}
