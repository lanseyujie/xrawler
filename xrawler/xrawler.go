package xrawler

import (
    "crypto/tls"
    "errors"
    "log"
    "net/http"
    "net/http/cookiejar"
    "net/http/httputil"
    "net/url"
    "time"
)

type Proxy = func(*http.Request) (*url.URL, error)

type Xrawler struct {
    DebugRequest      bool
    DebugResponse     bool
    Attempts          int
    UserAgent         string
    ConnectTimeout    time.Duration
    ReadWriteTimeout  time.Duration
    DisableKeepAlives bool
    Proxy             Proxy
    TLSClientConfig   *tls.Config
}

var (
    defaultXrawler   *Xrawler
    defaultCookieJar http.CookieJar
)

func init() {
    defaultXrawler = NewXrawler()
    defaultCookieJar, _ = cookiejar.New(nil)
}

func NewXrawler() *Xrawler {
    return &Xrawler{
        UserAgent:        `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36`,
        ConnectTimeout:   20,
        ReadWriteTimeout: 20,
    }
}

func SetDebug(requestFlag, responseFlag bool) {
    defaultXrawler.DebugRequest = requestFlag
    defaultXrawler.DebugResponse = responseFlag
}

func SetUserAgent(ua string) {
    defaultXrawler.UserAgent = ua
}

func SetConnectTimeout(t time.Duration) {
    defaultXrawler.ConnectTimeout = t
}

func SetReadWriteTimeout(t time.Duration) {
    defaultXrawler.ReadWriteTimeout = t
}

func SetProxy(proxy Proxy) {
    defaultXrawler.Proxy = proxy
}

func SetTLSClientConfig(config *tls.Config) {
    defaultXrawler.TLSClientConfig = config
}

func (c *Xrawler) Client(request *Request) (resp *http.Response, err error) {
    client := &http.Client{
        Transport: &http.Transport{
            DisableKeepAlives: c.DisableKeepAlives,
            TLSClientConfig:   c.TLSClientConfig,
            Proxy:             c.Proxy,
            // DialContext:     func(ctx context.Context, network, addr string) (net.Conn, error),
            MaxIdleConnsPerHost: 100,
        },
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            if !request.redirect {
                return http.ErrUseLastResponse
            } else if len(via) >= 10 {
                return errors.New("stopped after 10 redirects")
            }

            if c.DebugRequest {
                // 输出 http 请求报文
                if dump, e := httputil.DumpRequest(req, true); e != nil {
                    log.Println("REDIRECT REQUEST\n" + e.Error())
                } else {
                    log.Println("REDIRECT REQUEST\n" + string(dump))
                }
            }

            return nil
        },
        Jar: defaultCookieJar,
    }

    if c.UserAgent != "" && request.Input.Header.Get("User-Agent") == "" {
        request.Input.Header.Set("User-Agent", c.UserAgent)
    }

    // 0 默认无论成功与否只请求 1 次； -1 不断重试请求直至成功
    for i := 0; c.Attempts == -1 || i <= c.Attempts; i++ {
        if c.DebugRequest {
            // 输出 http 请求报文
            if dump, e := httputil.DumpRequest(request.Input, true); e != nil {
                log.Println("REQUEST\n" + e.Error())
            } else {
                log.Println("REQUEST\n" + string(dump))
            }
        }

        resp, err = client.Do(request.Input)
        if err != nil {
            log.Println(err)
        } else {
            if c.DebugResponse {
                // 输出 http 响应报文
                if dump, e := httputil.DumpResponse(resp, true); e != nil {
                    log.Println("RESPONSE\n" + e.Error())
                } else {
                    log.Println("RESPONSE\n" + string(dump))
                }
            }

            break
        }
    }

    return
}
