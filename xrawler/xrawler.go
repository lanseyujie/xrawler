package xrawler

import (
    "crypto/tls"
    "io"
    "io/ioutil"
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

func (c *Xrawler) debug(req *http.Request, resp *http.Response) {
    if c.DebugRequest {
        // 输出 http 请求报文
        if dump, err := httputil.DumpRequest(req, true); err != nil {
            log.Println(err.Error())
        } else {
            log.Println(string(dump))
        }
    }

    if c.DebugResponse {
        // 输出 http 响应报文
        if dump, err := httputil.DumpResponse(resp, true); err != nil {
            log.Println(err.Error())
        } else {
            log.Println(string(dump))
        }
    }
}

func (c *Xrawler) Client(req *http.Request) (resp *http.Response, err error) {
    client := &http.Client{
        Transport: &http.Transport{
            DisableKeepAlives: c.DisableKeepAlives,
            TLSClientConfig:   c.TLSClientConfig,
            Proxy:             c.Proxy,
            // DialContext:     func(ctx context.Context, network, addr string) (net.Conn, error),
            MaxIdleConnsPerHost: 100,
        },
        Jar: defaultCookieJar,
    }

    if c.UserAgent != "" && req.Header.Get("User-Agent") == "" {
        req.Header.Set("User-Agent", c.UserAgent)
    }

    // 0 默认无论成功与否只请求 1 次； -1 不断重试请求直至成功
    for i := 0; c.Attempts == -1 || i <= c.Attempts; i++ {
        resp, err = client.Do(req)
        c.debug(req, resp)

        if err == nil {
            break
        } else if resp != nil {
            // 得到一个重定向的错误时，两个变量都将是 non-nil
            // http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/#anameclose_http_resp_bodyaclosinghttpresponsebody
            _, _ = io.Copy(ioutil.Discard, resp.Body)
            _ = resp.Body.Close()
        }
    }

    return
}
