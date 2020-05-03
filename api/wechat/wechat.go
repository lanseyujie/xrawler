package wechat

import (
    "errors"
    "math"
    "strconv"
    "strings"
    "xrawler/xrawler"
)

// API 20200501

var (
    EditPageUrl = `https://mp.weixin.qq.com/cgi-bin/appmsg?t=media/appmsg_edit_v2&action=edit&isNew=1&type=10&token=1334668015&lang=zh_CN`
    SearchApi   = `https://mp.weixin.qq.com/cgi-bin/searchbiz?action=search_biz&begin=0&count=5&query=golang%E6%9D%A5%E5%95%A6&token=1334668015&lang=zh_CN&f=json&ajax=1`
    ArticleApi  = `https://mp.weixin.qq.com/cgi-bin/appmsg?action=list_ex&begin=0&count=5&fakeid=MzI2MDA1MTcxMg==&type=9&query=&token=1334668015&lang=zh_CN&f=json&ajax=1`
)

// 更新并缓存 cookie
// token 和 cookie 需登录微信公众平台后获取
func UpdateCookie(token, cookies string) error {
    EditPageUrl = strings.Replace(EditPageUrl, "token=1334668015", "token="+token, 1)
    SearchApi = strings.Replace(SearchApi, "token=1334668015", "token="+token, 1)
    ArticleApi = strings.Replace(ArticleApi, "token=1334668015", "token="+token, 1)

    // 请求页面用于 cookiejar 缓存 cookie
    resp, err := xrawler.Get(EditPageUrl).SetCookie(cookies).Do()
    if err != nil {
        return err
    }
    if resp.StatusCode != 200 {
        return errors.New("api.wechat: status code is not 200 (code:" + strconv.Itoa(resp.StatusCode) + ")")
    }

    return nil
}

// 搜索结果
type SearchResult struct {
    BaseResp struct {
        Ret    int    `json:"ret"`
        ErrMsg string `json:"err_msg"`
    } `json:"base_resp"`
    List []struct {
        FakeId       string `json:"fakeid"`
        Nickname     string `json:"nickname"`
        Alias        string `json:"alias"`
        RoundHeadImg string `json:"round_head_img"`
        ServiceType  int    `json:"service_type"`
    } `json:"list"`
    Total int `json:"total"`
}

// 每页的文章列表
type AppMsgList []struct {
    AId               string `json:"aid"`
    AlbumId           string `json:"album_id"`
    AppMsgId          int64  `json:"appmsgid"`
    Checking          int    `json:"checking"`
    Cover             string `json:"cover"`
    CreateTime        int64  `json:"create_time"`
    Digest            string `json:"digest"`
    HasRedPacketCover int    `json:"has_red_packet_cover"`
    IsOriginal        int    `json:"is_original"`
    IsPaySubscribe    int    `json:"is_pay_subscribe"`
    ItemShowType      int    `json:"item_show_type"`
    ItemIdx           int    `json:"itemidx"`
    Link              string `json:"link"`
    TagId             []int  `json:"tagid"`
    Title             string `json:"title"`
    UpdateTime        int64  `json:"update_time"`
}

// 文章查询结果
type ArticleResult struct {
    AppMsgCnt  int `json:"app_msg_cnt"`
    AppMsgList `json:"app_msg_list"`
    BaseResp   struct {
        Ret    int    `json:"ret"`
        ErrMsg string `json:"err_msg"`
    } `json:"base_resp"`
}

// 查找公众号的 FakeId
func GetFakeId(name string) (fakeId string, err error) {
    var s SearchResult

    err = xrawler.
        Get(SearchApi).
        SetReferer(EditPageUrl).
        SetParam("query", name).
        Json(&s)

    if err != nil {
        return "", err
    }

    if s.BaseResp.Ret == 0 {
        if len(s.List) > 0 {
            for _, val := range s.List {
                if name == val.Alias {
                    fakeId = val.FakeId
                    break
                }
            }
        }

        if fakeId == "" {
            err = errors.New("api.wechat: can not find the fake id")
        }
    } else {
        err = errors.New(s.BaseResp.ErrMsg)
    }

    return
}

// 获取文章列表
// 参数 ageLimit 为限制页数，按新旧顺序抓取，每页默认 5 条不可修改
func GetArticleList(fakeId string, pageLimit int) (list AppMsgList, err error) {
    currentPage := 1
    begin := 0

    var a ArticleResult
    err = xrawler.
        Get(ArticleApi).
        SetReferer(EditPageUrl).
        SetParam("begin", strconv.Itoa(begin)).
        SetParam("fakeid", fakeId).
        Json(&a)

    if err != nil {
        return nil, err
    }

    if a.BaseResp.Ret != 0 {
        return nil, errors.New(a.BaseResp.ErrMsg)
    }

    // 先查第一页以获取文章总数
    list = append(list, a.AppMsgList...)
    pageCnt := int(math.Ceil(float64(a.AppMsgCnt / 5.0)))

    // 不宜使用多线程，容易被 freq control
    for currentPage < pageCnt {
        currentPage++
        if pageLimit > 0 && currentPage > pageLimit {
            break
        }

        begin = 5 * (currentPage - 1)

        err = xrawler.
            Get(ArticleApi).
            SetReferer(EditPageUrl).
            SetParam("begin", strconv.Itoa(begin)).
            SetParam("fakeid", fakeId).
            Json(&a)

        if err != nil {
            break
        }
        list = append(list, a.AppMsgList...)
    }

    return
}
