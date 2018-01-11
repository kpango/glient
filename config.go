package glient

import (
	"crypto/tls"
	"time"
)

type Config struct {
	DNSResolveCacheFlg    bool
	DNSCacheTimeout       time.Duration
	DialerTimeout         time.Duration
	DialerKeepAlive       time.Duration
	DialerDualStack       bool
	DisableKeepAlives     bool
	ExpectContinueTimeout time.Duration
	IdleConnTimeout       time.Duration
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	ResponseHeaderTimeout time.Duration
	TLSHandshakeTimeout   time.Duration
	TLSConfig             *tls.Config
}

var (
	DefaultConfig = &Config{
		DNSResolveCacheFlg:    true,
		DNSCacheTimeout:       30 * time.Second,
		DialerTimeout:         30 * time.Second,
		DialerKeepAlive:       30 * time.Second,
		DialerDualStack:       true,
		DisableKeepAlives:     false,
		ExpectContinueTimeout: 0 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          810,
		MaxIdleConnsPerHost:   18,
		ResponseHeaderTimeout: 10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		TLSConfig:             &tls.Config{InsecureSkipVerify: true},
	}
)

func NewConfig() *Config {
	return DefaultConfig
}
