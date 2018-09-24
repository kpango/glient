package glient

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"

	"github.com/kpango/gache"
	"golang.org/x/sync/singleflight"
)

type Glient struct {
	storage     gache.Gache
	ipMap       gache.Gache
	cacheFlg    bool
	errors      []error
	maxRedirect int
	redirectFlg bool
	userAgent   string
	client      *http.Client
	tran        *http.Transport
	jar         *cookiejar.Jar
}

const (
	defaultUA = "golang kpango Glient"
)

var (
	glient *Glient
	once   sync.Once
	mu     sync.Mutex
)

func newTransport(g gache.Gache, conf *Config) *http.Transport {
	var group singleflight.Group

	var dial func(ctx context.Context, network, host string) (net.Conn, error)

	if conf.DNSResolveCacheFlg {
		dial = func(ctx context.Context, network, host string) (conn net.Conn, err error) {
			sep := strings.LastIndex(host, ":")
			ip, ok := g.Get(host[:sep])
			if ok {
				conn, err = net.Dial("tcp", ip.(string)+host[sep:])
				if err == nil {
					return
				}
				if conn != nil {
					conn.Close()
				}
			}

			ip, err, _ = group.Do(host[:sep], func() (interface{}, error) {
				r, err := net.DefaultResolver.LookupIPAddr(context.Background(), host[:sep])
				if err != nil {
					return nil, err
				}
				url := r[0].String()
				g.SetWithExpire(host[:sep], url, conf.DNSCacheTimeout)
				return url, err
			})

			if err == nil {
				conn, err = net.Dial("tcp", ip.(string)+host[sep:])
				if err == nil {
					return
				}
				if conn != nil {
					conn.Close()
				}
				g.Delete(host[:sep])
			}

			return (&net.Dialer{
				Timeout:   conf.DialerTimeout,
				KeepAlive: conf.DialerKeepAlive,
				DualStack: conf.DialerDualStack,
			}).DialContext(ctx, network, host)
		}
	} else {
		dial = func(ctx context.Context, network, host string) (net.Conn, error) {
			return (&net.Dialer{
				Timeout:   conf.DialerTimeout,
				KeepAlive: conf.DialerKeepAlive,
				DualStack: conf.DialerDualStack,
			}).DialContext(ctx, network, host)
		}
	}

	return &http.Transport{
		DialContext:           dial,
		DisableKeepAlives:     conf.DisableKeepAlives,
		ExpectContinueTimeout: conf.ExpectContinueTimeout,
		IdleConnTimeout:       conf.IdleConnTimeout,
		MaxIdleConns:          conf.MaxIdleConns,
		MaxIdleConnsPerHost:   conf.MaxIdleConnsPerHost,
		Proxy:                 http.ProxyFromEnvironment,
		ResponseHeaderTimeout: conf.ResponseHeaderTimeout,
		TLSHandshakeTimeout:   conf.TLSHandshakeTimeout,
		TLSClientConfig:       conf.TLSConfig,
	}
}

// New Returns new Glient instance
func New(conf *Config) *Glient {
	if conf == nil {
		conf = DefaultConfig
	}
	g := gache.New()
	tran := newTransport(g, conf)

	http.DefaultTransport = tran

	jar, err := cookiejar.New(&cookiejar.Options{})

	if err != nil {
		return &Glient{
			client:      http.DefaultClient,
			tran:        tran,
			maxRedirect: 0,
			cacheFlg:    false,
			userAgent:   defaultUA,
			ipMap:       g,
			storage:     gache.New(),
		}
	}

	return &Glient{
		client: &http.Client{
			Jar:       jar,
			Transport: tran,
		},
		tran:        tran,
		jar:         jar,
		maxRedirect: 0,
		cacheFlg:    false,
		userAgent:   defaultUA,
		ipMap:       g,
		storage:     gache.New(),
	}
}

// GetGlient Returns singleton Glient instance
func Init(conf *Config) *Glient {
	// instantiate once
	once.Do(func() {
		glient = New(conf)
	})
	return glient
}

func (g *Glient) SetUserAgent(userAgent string) *Glient {
	g.userAgent = userAgent
	return g
}

func LoadCachedIPs() (m map[string]string) {
	return glient.LoadCachedIPs()
}

func (g *Glient) LoadCachedIPs() (m map[string]string) {
	for k, v := range g.ipMap.ToMap() {
		m[k.(string)] = v.(string)
	}
	return
}

func Head(url string, body io.Reader) (*http.Response, error) {
	return glient.Head(url, body)
}

func (g *Glient) Head(url string, body io.Reader) (*http.Response, error) {
	return g.Request(http.MethodHead, url, nil, body)
}

func Get(url string) (*http.Response, error) {
	return glient.Get(url)
}

func (g *Glient) Get(url string) (*http.Response, error) {
	return g.Request(http.MethodGet, url, nil, nil)
}

func Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return glient.Post(url, contentType, body)
}

func (g *Glient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return g.client.Post(url, contentType, body)
}

func Put(url string, body io.Reader) (*http.Response, error) {
	return Put(url, body)
}

func (g *Glient) Put(url string, body io.Reader) (*http.Response, error) {
	return g.Request(http.MethodPut, url, nil, body)
}

func Delete(url string, body io.Reader) (*http.Response, error) {
	return glient.Delete(url, body)
}

func (g *Glient) Delete(url string, body io.Reader) (*http.Response, error) {
	return g.Request(http.MethodDelete, url, nil, body)
}

func Request(method, url string, header map[string][]string, body io.Reader) (*http.Response, error) {
	return glient.Request(method, url, header, body)
}

func (g *Glient) Request(method, url string, header map[string][]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		for _, val := range v {
			req.Header.Add(k, val)
		}
	}
	return g.client.Do(req)
}

func Do(req *http.Request) (*http.Response, error) {
	return glient.Do(req)
}

func (g *Glient) Do(req *http.Request) (*http.Response, error) {
	return g.client.Do(req)
}

func (g *Glient) GetResolvedIP(host string) (string, bool) {
	sep := strings.LastIndex(host, ":")
	val, ok := g.ipMap.Get(host[:sep])

	if ok {
		return val.(string), ok
	}
	return "Not Found", ok
}

func (g *Glient) DeleteResolvedIP(host string) {
	sep := strings.LastIndex(host, ":")
	g.ipMap.Delete(host[:sep])
}
