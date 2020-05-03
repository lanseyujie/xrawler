package xrawler

import (
    "bytes"
    "encoding/json"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
)

type Request struct {
    Input  *http.Request
    Output *http.Response
    params url.Values
    err    error
}

func NewRequest(rawUrl, method string) *Request {
    req, err := http.NewRequest(method, rawUrl, nil)
    if err != nil {
        return &Request{
            err: err,
        }
    }

    // 缓存请求参数
    params := url.Values{}
    if method == http.MethodHead || method == http.MethodGet {
        params = req.URL.Query()
    }

    return &Request{
        Input:  req,
        Output: &http.Response{},
        params: params,
    }
}

func (req *Request) SetHost(host string) *Request {
    req.Input.Host = host

    return req
}

func (req *Request) SetReferer(url string) *Request {
    req.Input.Header.Set("Referer", url)

    return req
}

func (req *Request) SetCookie(cookie interface{}) *Request {
    var ck string
    switch c := cookie.(type) {
    case *http.Cookie:
        ck = c.String()
    case []byte:
        ck = string(c)
    case string:
        ck = c
    }

    req.Input.Header.Add("Cookie", ck)

    return req
}

func (req *Request) SetHeader(key, value string) *Request {
    if req.err != nil {
        return req
    }

    req.Input.Header.Set(key, value)

    return req
}

// 增加 get 或 post 参数，不会覆盖之前的参数
// 如：id=1&id=2
func (req *Request) AddParam(key string, value string) *Request {
    if req.err != nil {
        return req
    }

    req.params.Add(key, value)

    return req
}

// 增加 get 或 post 参数，会覆盖之前的同名参数
func (req *Request) SetParam(key string, value string) *Request {
    if req.err != nil {
        return req
    }

    req.params.Set(key, value)

    return req
}

func (req *Request) SetBody(body interface{}) *Request {
    if req.err != nil {
        return req
    }

    switch b := body.(type) {
    case []byte:
        buf := bytes.NewBuffer(b)
        req.Input.Body = ioutil.NopCloser(buf)
        req.Input.ContentLength = int64(len(b))

    case string:
        buf := bytes.NewBufferString(b)
        req.Input.Body = ioutil.NopCloser(buf)
        req.Input.ContentLength = int64(len(b))
    }

    return req
}

func (req *Request) SetJsonBody(obj interface{}) *Request {
    if req.err != nil {
        return req
    }

    if req.Input.Body == nil && obj != nil {
        bs, err := json.Marshal(obj)
        if err != nil {
            req.err = err

            return req
        }

        req.Input.Body = ioutil.NopCloser(bytes.NewReader(bs))
        req.Input.ContentLength = int64(len(bs))
        req.Input.Header.Set("Content-Type", "application/json")
    }

    return req
}

func (req *Request) Error() error {
    return req.err
}

func (req *Request) Do(c ...*Xrawler) (*http.Response, error) {
    if req.err != nil {
        return nil, req.err
    }

    if req.Input.Method == http.MethodHead || req.Input.Method == http.MethodGet {
        req.Input.URL.RawQuery = req.params.Encode()
    } else if req.Input.Method == http.MethodPost {
        req.Input.Header.Set("Content-Type", "application/x-www-form-urlencoded")
        req.SetBody(req.params.Encode())
    }

    var err error
    if len(c) > 0 && c[0] != nil {
        req.Output, err = c[0].Client(req.Input)
    } else {
        req.Output, err = defaultXrawler.Client(req.Input)
    }

    return req.Output, err
}

// 输出响应 body
func (req *Request) Bytes(c ...*Xrawler) ([]byte, error) {
    if req.err != nil {
        return nil, req.err
    }

    resp, err := req.Do(c...)
    if err != nil {
        return nil, err
    }

    if resp.Body == nil {
        return nil, nil
    }
    defer resp.Body.Close()

    buffer := bytes.NewBuffer(make([]byte, 0, 4096))
    _, err = io.Copy(buffer, resp.Body)
    if err != nil {
        return nil, err
    }

    return buffer.Bytes(), nil
}

// 输出响应 body
func (req *Request) String(c ...*Xrawler) (string, error) {
    if req.err != nil {
        return "", req.err
    }

    data, err := req.Bytes(c...)
    if err != nil {
        return "", err
    }

    return string(data), nil
}

// 解析响应 body 中的 json
func (req *Request) Json(m interface{}, c ...*Xrawler) error {
    if req.err != nil {
        return req.err
    }

    data, err := req.Bytes(c...)
    if err != nil {
        return err
    }

    return json.Unmarshal(data, m)
}

func Get(url string) *Request {
    return NewRequest(url, http.MethodGet)
}

// func Post(url string) *Request {
//     return NewRequest(url, http.MethodPost)
// }

// func Head(url string) *Request {
//     return NewRequest(url, http.MethodHead)
// }
