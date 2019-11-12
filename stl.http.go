package stl

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type HTTPRequest  *http.Request
type HTTPResponse *http.Response

const (
	HTTP_POST    = "POST"
	HTTP_GET     = "GET"
	HTTP_HEAD    = "HEAD"
	HTTP_PUT     = "PUT"
	HTTP_DELETE  = "DELETE"
	HTTP_PATCH   = "PATCH"
	HTTP_OPTIONS = "OPTIONS"
)

type XPHttpImpl struct {
	Url               string
	Method            string
	Headers           map[string]string
	TargetType        string
	ForceType         string
	Data              map[string]interface{}
	SliceData         []interface{}
	FormData          url.Values
	QueryData         url.Values
	FileData          []HTTP_File
	BounceToRawString bool
	RawString         string
	Client            *http.Client
	Transport         *http.Transport
	ReqCookies        []*http.Cookie
	CookieJar         *cookiejar.Jar
	Errors            []error
	BasicAuth         struct{ Username, Password string }
	Debug             bool
	CurlCommand       bool
	logger            *log.Logger
	Retryable         struct {
		RetryableStatus []int
		RetryTime       time.Duration
		RetryCount      int
		Attempt         int
		Enable          bool
	}
}

var HttpDisableTransportSwap = false

func NewHttp() *XPHttpImpl {
	cookieOptions := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, _ := cookiejar.New(&cookieOptions)

	debug := os.Getenv("XPSUPERKIT_DEBUG") == "1"

	h := &XPHttpImpl{
		TargetType:        "json",
		Data:              make(map[string]interface{}),
		Headers:           make(map[string]string),
		RawString:         "",
		SliceData:         []interface{}{},
		FormData:          url.Values{},
		QueryData:         url.Values{},
		FileData:          make([]HTTP_File, 0),
		BounceToRawString: false,
		Client:            &http.Client{Jar: jar},
		Transport:         &http.Transport{},
		CookieJar:         jar,
		ReqCookies:        make([]*http.Cookie, 0),
		Errors:            nil,
		BasicAuth:         struct{ Username, Password string }{},
		Debug:             debug,
		CurlCommand:       false,
		logger:            log.New(os.Stderr, "[gorequest]", log.LstdFlags),
	}

	h.Transport.DisableKeepAlives = true
	return h
}

func (h *XPHttpImpl) SetDebug(enable bool) *XPHttpImpl {
	h.Debug = enable
	return h
}

func (h *XPHttpImpl) SetCurlCommand(enable bool) *XPHttpImpl {
	h.CurlCommand = enable
	return h
}

func (h *XPHttpImpl) SetLogger(logger *log.Logger) *XPHttpImpl {
	h.logger = logger
	return h
}

func (h *XPHttpImpl) GetCookie(u *url.URL, key string) *http.Cookie {
	cookies := h.CookieJar.Cookies(u)
	var result *http.Cookie
	for _, cookie := range cookies {
		if strings.EqualFold(key, cookie.Name) {
			result = cookie
			break
		}
	}
	return result
}

func (h *XPHttpImpl) Reset() {
	h.Url = ""
	h.Method = ""
	h.Headers = make(map[string]string)
	h.Data = make(map[string]interface{})
	h.SliceData = []interface{}{}
	h.FormData = url.Values{}
	h.QueryData = url.Values{}
	h.FileData = make([]HTTP_File, 0)
	h.BounceToRawString = false
	h.RawString = ""
	h.ForceType = ""
	h.TargetType = "json"
	h.ReqCookies = make([]*http.Cookie, 0)
	h.Errors = nil
}

func (h *XPHttpImpl) CustomMethod(method, targetUrl string) *XPHttpImpl {
	switch method {
	case HTTP_POST:
		return h.Post(targetUrl)
	case HTTP_GET:
		return h.Get(targetUrl)
	case HTTP_HEAD:
		return h.Head(targetUrl)
	case HTTP_PUT:
		return h.Put(targetUrl)
	case HTTP_DELETE:
		return h.Delete(targetUrl)
	case HTTP_PATCH:
		return h.Patch(targetUrl)
	case HTTP_OPTIONS:
		return h.Options(targetUrl)
	default:
		h.Reset()
		h.Method = method
		h.Url = targetUrl
		h.Errors = nil
		return h
	}
}

func (h *XPHttpImpl) Get(targetUrl string) *XPHttpImpl {
	h.Reset()
	h.Method = HTTP_GET
	h.Url = targetUrl
	h.Errors = nil
	return h
}

func (h *XPHttpImpl) Post(targetUrl string) *XPHttpImpl {
	h.Reset()
	h.Method = HTTP_POST
	h.Url = targetUrl
	h.Errors = nil
	return h
}

func (h *XPHttpImpl) Head(targetUrl string) *XPHttpImpl {
	h.Reset()
	h.Method = HTTP_HEAD
	h.Url = targetUrl
	h.Errors = nil
	return h
}

func (h *XPHttpImpl) Put(targetUrl string) *XPHttpImpl {
	h.Reset()
	h.Method = HTTP_PUT
	h.Url = targetUrl
	h.Errors = nil
	return h
}

func (h *XPHttpImpl) Delete(targetUrl string) *XPHttpImpl {
	h.Reset()
	h.Method = HTTP_DELETE
	h.Url = targetUrl
	h.Errors = nil
	return h
}

func (h *XPHttpImpl) Patch(targetUrl string) *XPHttpImpl {
	h.Reset()
	h.Method = HTTP_PATCH
	h.Url = targetUrl
	h.Errors = nil
	return h
}

func (h *XPHttpImpl) Options(targetUrl string) *XPHttpImpl {
	h.Reset()
	h.Method = HTTP_OPTIONS
	h.Url = targetUrl
	h.Errors = nil
	return h
}

// 自定义请求头
//
// 例如
//    XPSuperKit.NewHttp().
//      Header("Accept", "application/json").
//      Post("/gamelist").
//      End()
func (h *XPHttpImpl) Header(key string, value string) *XPHttpImpl {
	h.Headers[key] = value
	return h
}

// 用于设置一个重试机制
//
// 例如 每隔5秒重试一次，最多重试3次，当状态为 StatusInternalServerError 或 StatusInternalServerError 时重试
//    XPSuperKit.NewHttp().
//      Post("/gamelist").
//      Retry(3, 5 * time.seconds, http.StatusBadRequest, http.StatusInternalServerError).
//      End()
func (h *XPHttpImpl) Retry(retryCount int, retryTime time.Duration, statusCode ...int) *XPHttpImpl {
	for _, code := range statusCode {
		statusText := http.StatusText(code)
		if len(statusText) == 0 {
			h.Errors = append(h.Errors, NewErrors("StatusCode '"+strconv.Itoa(code)+"' doesn't exist in http package"))
		}
	}

	h.Retryable = struct {
		RetryableStatus []int
		RetryTime     time.Duration
		RetryCount    int
		Attempt       int
		Enable        bool
	}{
		statusCode,
		retryTime,
		retryCount,
		0,
		true,
	}
	return h
}

// 用于设置基本认证
//
// 例如 设置认证用户名 authUser 和 认证密码 authPassword
//    XPSuperKit.NewHttp().
//      Post("/gamelist").
//      Auth("authUser", "authPassword").
//      End()
func (h *XPHttpImpl) Auth(username string, password string) *XPHttpImpl {
	h.BasicAuth = struct{ Username, Password string }{username, password}
	return h
}

// 添加 Cookie
func (h *XPHttpImpl) Cookie(c *http.Cookie) *XPHttpImpl {
	h.ReqCookies = append(h.ReqCookies, c)
	return h
}

// 添加 Cookies
func (h *XPHttpImpl) Cookies(cookies []*http.Cookie) *XPHttpImpl {
	h.ReqCookies = append(h.ReqCookies, cookies...)
	return h
}

var HTTP_ContentTypes = map[string]string{
	"html":       "text/html",
	"json":       "application/json",
	"xml":        "application/xml",
	"text":       "text/plain",
	"urlencoded": "application/x-www-form-urlencoded",
	"form":       "application/x-www-form-urlencoded",
	"form-data":  "application/x-www-form-urlencoded",
	"multipart":  "multipart/form-data",
}

// 用于设置请求数据类型
//
// 例如 请求将发送 `application/x-www-form-urlencoded` 格式的数据
//    XPSuperKit.NewHttp().
//      Post("/recipe").
//      Type("form").
//      Send(`{ "name": "egg benedict", "category": "brunch" }`).
//      End()
func (h *XPHttpImpl) ContentType(typeStr string) *XPHttpImpl {
	if _, ok := HTTP_ContentTypes[typeStr]; ok {
		h.ForceType = typeStr
	} else {
		h.Errors = append(h.Errors, NewErrors("Type func: incorrect type \""+typeStr+"\""))
	}
	return h
}

// 用于设置 Query 参数
// For example, making "/search?query=bicycle&size=50x50&weight=20kg" using GET method:
// 例如 要构建 "/search?query=bicycle&size=50x50&weight=20kg" 形式的 Query 参数
//      XPSuperKit.NewHttp().
//        Get("/search").
//        Query(`{ query: 'bicycle' }`).
//        Query(`{ size: '50x50' }`).
//        Query(`{ weight: '20kg' }`).
//        End()
//
// 也可以直接接收 Json 形式的字符串
//
//      XPSuperKit.NewHttp().
//        Get("/search").
//        Query(`{ query: 'bicycle', size: '50x50', weight: '20kg' }`).
//        End()
//
// 或者直接是 Query 参数形式（key=value&key=value）字符串
//
//      XPSuperKit.NewHttp().
//        Get("/search").
//        Query("query=bicycle&size=50x50").
//        Query("weight=20kg").
//        End()
//
// 或者两者混合
//
//      XPSuperKit.NewHttp().
//        Get("/search").
//        Query("query=bicycle").
//        Query(`{ size: '50x50', weight:'20kg' }`).
//        End()
//
func (h *XPHttpImpl) Query(content interface{}) *XPHttpImpl {
	switch v := reflect.ValueOf(content); v.Kind() {
	case reflect.String:
		h.queryString(v.String())
	case reflect.Struct:
		h.queryStruct(v.Interface())
	case reflect.Map:
		h.queryMap(v.Interface())
	default:
	}
	return h
}

func (h *XPHttpImpl) queryStruct(content interface{}) *XPHttpImpl {
	if marshalContent, err := json.Marshal(content); err != nil {
		h.Errors = append(h.Errors, err)
	} else {
		var val map[string]interface{}
		if err := json.Unmarshal(marshalContent, &val); err != nil {
			h.Errors = append(h.Errors, err)
		} else {
			for k, v := range val {
				k = strings.ToLower(k)
				var queryVal string
				switch t := v.(type) {
				case string:
					queryVal = t
				case float64:
					queryVal = strconv.FormatFloat(t, 'f', -1, 64)
				case time.Time:
					queryVal = t.Format(time.RFC3339)
				default:
					j, err := json.Marshal(v)
					if err != nil {
						continue
					}
					queryVal = string(j)
				}
				h.QueryData.Add(k, queryVal)
			}
		}
	}
	return h
}

func (h *XPHttpImpl) queryString(content string) *XPHttpImpl {
	var val map[string]string
	if err := json.Unmarshal([]byte(content), &val); err == nil {
		for k, v := range val {
			h.QueryData.Add(k, v)
		}
	} else {
		if queryData, err := url.ParseQuery(content); err == nil {
			for k, queryValues := range queryData {
				for _, queryValue := range queryValues {
					h.QueryData.Add(k, string(queryValue))
				}
			}
		} else {
			h.Errors = append(h.Errors, err)
		}
		// TODO: need to check correct format of 'field=val&field=val&...'
	}
	return h
}

func (h *XPHttpImpl) queryMap(content interface{}) *XPHttpImpl {
	return h.queryStruct(content)
}

// 用于接收特殊形式的参数，如 fields=f1;f2;f3
func (h *XPHttpImpl) Param(key string, value string) *XPHttpImpl {
	h.QueryData.Add(key, value)
	return h
}

// 用于设置超时
func (h *XPHttpImpl) Timeout(timeout time.Duration) *XPHttpImpl {
	h.Transport.Dial = func(network, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(network, addr, timeout)
		if err != nil {
			h.Errors = append(h.Errors, err)
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(timeout))
		return conn, nil
	}
	return h
}

// 用于设置 TLS
//
// 例如 可以用以下形式禁用 HTTPS 安全校验
//      XPSuperKit.NewHttp().
//        TLS(&tls.Config{ InsecureSkipVerify: true}).
//        Get("https://disable-security-check.com").
//        End()
func (h *XPHttpImpl) TLS(config *tls.Config) *XPHttpImpl {
	h.Transport.TLSClientConfig = config
	return h
}

// 用于设置请求代理
//
//      XPSuperKit.NewHttp().
//        Proxy("http://myproxy:9999").
//        Post("http://www.google.com").
//        End()
//
// 取消代理, 只需要传空字符串:
//
//      XPSuperKit.NewHttp().
//        Proxy("").
//        Post("http://www.google.com").
//        End()
//
func (h *XPHttpImpl) Proxy(proxyUrl string) *XPHttpImpl {
	parsedProxyUrl, err := url.Parse(proxyUrl)
	if err != nil {
		h.Errors = append(h.Errors, err)
	} else if proxyUrl == "" {
		h.Transport.Proxy = nil
	} else {
		h.Transport.Proxy = http.ProxyURL(parsedProxyUrl)
	}
	return h
}

// 用于接收一个函数来处理 Redirect，如果该函数返回 error, 则当跳转指令返回后不会创建下一个请求
// 该函数的入参是即将跳转的 Request 以及之前的 Request 序列
func (h *XPHttpImpl) Redirect(policy func(req HTTPRequest, via []HTTPRequest) error) *XPHttpImpl {
	h.Client.CheckRedirect = func(r *http.Request, v []*http.Request) error {
		vv := make([]HTTPRequest, len(v))
		for i, r := range v {
			vv[i] = HTTPRequest(r)
		}
		return policy(HTTPRequest(r), vv)
	}
	return h
}

// 用于POST 和 PUT 方法发送数据，无特定参数形式，可以是以下各种类型的参数：
// 1、Json 字符串
//      XPSuperKit.NewHttp().
//        Post("/search").
//        Send(`{ query: 'sushi' }`).
//        End()
//
// 2、键值对字符串
//      XPSuperKit.NewHttp().
//        Post("/search").
//        Send("query=tonkatsu").
//        End()
//
// 3、Json 与 键值对混合
//      XPSuperKit.NewHttp().
//        Post("/search").
//        Send("query=bicycle&size=50x50").
//        Send(`{ wheel: '4'}`).
//        End()
//
// 4、Struct
//      type BrowserVersionSupport struct {
//        Chrome string
//        Firefox string
//      }
//      ver := BrowserVersionSupport{ Chrome: "37.0.2041.6", Firefox: "30.0" }
//      XPSuperKit.NewHttp().
//        Post("/update_version").
//        Send(ver).
//        Send(`{"Safari":"5.1.10"}`).
//        End()
//
// 5、普通字符串
//      XPSuperKit.NewHttp().
//        Post("/greet").
//        Type("text").
//        Send("hello world").
//        End()
func (h *XPHttpImpl) Send(content interface{}) *XPHttpImpl {
	// TODO: add normal text mode or other mode to Send func
	switch v := reflect.ValueOf(content); v.Kind() {
	case reflect.String:
		h.SendString(v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // includes rune
		h.SendString(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // includes byte
		h.SendString(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float64:
		h.SendString(strconv.FormatFloat(v.Float(), 'f', -1, 64))
	case reflect.Float32:
		h.SendString(strconv.FormatFloat(v.Float(), 'f', -1, 32))
	case reflect.Bool:
		h.SendString(strconv.FormatBool(v.Bool()))
	case reflect.Struct:
		h.SendStruct(v.Interface())
	case reflect.Slice:
		h.SendSlice(makeSliceOfReflectValue(v))
	case reflect.Array:
		h.SendSlice(makeSliceOfReflectValue(v))
	case reflect.Ptr:
		h.Send(v.Elem().Interface())
	case reflect.Map:
		h.SendMap(v.Interface())
	default:
		// TODO: leave default for handling other types in the future, such as complex numbers, (nested) maps, etc
		return h
	}
	return h
}

func makeSliceOfReflectValue(v reflect.Value) (slice []interface{}) {

	kind := v.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return slice
	}

	slice = make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		slice[i] = v.Index(i).Interface()
	}

	return slice
}

// 发送 Slice
func (h *XPHttpImpl) SendSlice(content []interface{}) *XPHttpImpl {
	h.SliceData = append(h.SliceData, content...)
	return h
}

// 发送 Map
func (h *XPHttpImpl) SendMap(content interface{}) *XPHttpImpl {
	return h.SendStruct(content)
}

// 发送 Struct
func (h *XPHttpImpl) SendStruct(content interface{}) *XPHttpImpl {
	if marshalContent, err := json.Marshal(content); err != nil {
		h.Errors = append(h.Errors, err)
	} else {
		var val map[string]interface{}
		d := json.NewDecoder(bytes.NewBuffer(marshalContent))
		d.UseNumber()
		if err := d.Decode(&val); err != nil {
			h.Errors = append(h.Errors, err)
		} else {
			for k, v := range val {
				h.Data[k] = v
			}
		}
	}
	return h
}

// 发送字符串
func (h *XPHttpImpl) SendString(content string) *XPHttpImpl {
	if !h.BounceToRawString {
		var val interface{}
		d := json.NewDecoder(strings.NewReader(content))
		d.UseNumber()
		if err := d.Decode(&val); err == nil {
			switch v := reflect.ValueOf(val); v.Kind() {
			case reflect.Map:
				for k, v := range val.(map[string]interface{}) {
					h.Data[k] = v
				}
			// add to SliceData
			case reflect.Slice:
				h.SendSlice(val.([]interface{}))
			// bounce to rawstring if it is arrayjson, or others
			default:
				h.BounceToRawString = true
			}
		} else if formData, err := url.ParseQuery(content); err == nil {
			for k, formValues := range formData {
				for _, formValue := range formValues {
					// make it array if already have key
					if val, ok := h.Data[k]; ok {
						var strArray []string
						strArray = append(strArray, string(formValue))
						// check if previous data is one string or array
						switch oldValue := val.(type) {
						case []string:
							strArray = append(strArray, oldValue...)
						case string:
							strArray = append(strArray, oldValue)
						}
						h.Data[k] = strArray
					} else {
						// make it just string if does not already have same key
						h.Data[k] = formValue
					}
				}
			}
			h.TargetType = "form"
		} else {
			h.BounceToRawString = true
		}
	}
	// Dump all contents to RawString in case in the end user doesn't want json or form.
	h.RawString += content
	return h
}

type HTTP_File struct {
	Filename  string
	Fieldname string
	Data      []byte
}

// 用于通过 "multipart" 形式发送文件
//
// 1、可以以文件路径字符串作为参数
//      XPSuperKit.NewHttp().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile("./example_file.ext").
//        End()
//
// 2、也可以以已被类似于 ioutil.ReadFile 读取的 []byte Slice 作为参数
//      b, _ := ioutil.ReadFile("./example_file.ext")
//      XPSuperKit.NewHttp().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile(b).
//        End()
//
// 3、也可以以 os.File 作为参数
//      f, _ := os.Open("./example_file.ext")
//      XPSuperKit.NewHttp().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile(f).
//        End()
//
// 4、当参数多于 1 时，第一个参数为文件，可以是路径字符串、已读的[]byte Slice、os.File，当为[]byte Slice时，（第二个参数）默认为"filename"，（第二个参数）均可
//    自定义
//      b, _ := ioutil.ReadFile("./example_file.ext")
//      XPSuperKit.NewHttp().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile(b, "my_custom_filename").
//        End()
//
// 5、当参数多于 2 时，第三个参数默认 为 file1, file2, ...(发送多个文件的情况)，也可以自定义
//      b, _ := ioutil.ReadFile("./example_file.ext")
//      XPSuperKit.NewHttp().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile(b, "", "my_custom_fieldname"). // filename left blank, will become "example_file.ext"
//        End()
//
func (h *XPHttpImpl) SendFile(file interface{}, args ...string) *XPHttpImpl {

	filename := ""
	fieldname := "file"

	if len(args) >= 1 && len(args[0]) > 0 {
		filename = strings.TrimSpace(args[0])
	}
	if len(args) >= 2 && len(args[1]) > 0 {
		fieldname = strings.TrimSpace(args[1])
	}
	if fieldname == "file" || fieldname == "" {
		fieldname = "file" + strconv.Itoa(len(h.FileData)+1)
	}

	switch v := reflect.ValueOf(file); v.Kind() {
	case reflect.String:
		pathToFile, err := filepath.Abs(v.String())
		if err != nil {
			h.Errors = append(h.Errors, err)
			return h
		}
		if filename == "" {
			filename = filepath.Base(pathToFile)
		}
		data, err := ioutil.ReadFile(v.String())
		if err != nil {
			h.Errors = append(h.Errors, err)
			return h
		}
		h.FileData = append(h.FileData, HTTP_File{
			Filename:  filename,
			Fieldname: fieldname,
			Data:      data,
		})
	case reflect.Slice:
		slice := makeSliceOfReflectValue(v)
		if filename == "" {
			filename = "filename"
		}
		f := HTTP_File{
			Filename:  filename,
			Fieldname: fieldname,
			Data:      make([]byte, len(slice)),
		}
		for i := range slice {
			f.Data[i] = slice[i].(byte)
		}
		h.FileData = append(h.FileData, f)
	case reflect.Ptr:
		if len(args) == 1 {
			return h.SendFile(v.Elem().Interface(), args[0])
		}
		if len(args) >= 2 {
			return h.SendFile(v.Elem().Interface(), args[0], args[1])
		}
		return h.SendFile(v.Elem().Interface())
	default:
		if v.Type() == reflect.TypeOf(os.File{}) {
			osfile := v.Interface().(os.File)
			if filename == "" {
				filename = filepath.Base(osfile.Name())
			}
			data, err := ioutil.ReadFile(osfile.Name())
			if err != nil {
				h.Errors = append(h.Errors, err)
				return h
			}
			h.FileData = append(h.FileData, HTTP_File{
				Filename:  filename,
				Fieldname: fieldname,
				Data:      data,
			})
			return h
		}

		h.Errors = append(h.Errors, NewErrors("SendFile currently only supports either a string (path/to/file), a slice of bytes (file content itself), or a os.File!"))
	}

	return h
}

func changeMapToURLValues(data map[string]interface{}) url.Values {
	var newUrlValues = url.Values{}
	for k, v := range data {
		switch val := v.(type) {
		case string:
			newUrlValues.Add(k, val)
		case bool:
			newUrlValues.Add(k, strconv.FormatBool(val))
		// if a number, change to string
		// json.Number used to protect against a wrong (for GoRequest) default conversion
		// which always converts number to float64.
		// This type is caused by using Decoder.UseNumber()
		case json.Number:
			newUrlValues.Add(k, string(val))
		case int:
			newUrlValues.Add(k, strconv.FormatInt(int64(val), 10))
		// TODO add all other int-Types (int8, int16, ...)
		case float64:
			newUrlValues.Add(k, strconv.FormatFloat(float64(val), 'f', -1, 64))
		case float32:
			newUrlValues.Add(k, strconv.FormatFloat(float64(val), 'f', -1, 64))
		// following slices are mostly needed for tests
		case []string:
			for _, element := range val {
				newUrlValues.Add(k, element)
			}
		case []int:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatInt(int64(element), 10))
			}
		case []bool:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatBool(element))
			}
		case []float64:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatFloat(float64(element), 'f', -1, 64))
			}
		case []float32:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatFloat(float64(element), 'f', -1, 64))
			}
		// these slices are used in practice like sending a struct
		case []interface{}:

			if len(val) <= 0 {
				continue
			}

			switch val[0].(type) {
			case string:
				for _, element := range val {
					newUrlValues.Add(k, element.(string))
				}
			case bool:
				for _, element := range val {
					newUrlValues.Add(k, strconv.FormatBool(element.(bool)))
				}
			case json.Number:
				for _, element := range val {
					newUrlValues.Add(k, string(element.(json.Number)))
				}
			}
		default:
			// TODO add ptr, arrays, ...
		}
	}
	return newUrlValues
}

func (h *XPHttpImpl) BuildUrlValues(data map[string]interface{}) url.Values {
	return changeMapToURLValues(data)
}

func (h *XPHttpImpl) BuildUrlQuery(data map[string]interface{}) string {
	v := changeMapToURLValues(data)
	return v.Encode()
}

// End is the most important function that you need to call when ending the chain. The request won't proceed without calling it.
// End function returns Response which matchs the structure of Response type in Golang's http package (but without Body data). The body data itself returns as a string in a 2nd return value.
// Lastly but worth noticing, error array (NOTE: not just single error value) is returned as a 3rd value and nil otherwise.
//
// For example:
//
//    resp, body, errs := gorequest.New().Get("http://www.google.com").End()
//    if (errs != nil) {
//      fmt.Println(errs)
//    }
//    fmt.Println(resp, body)
//
// Moreover, End function also supports callback which you can put as a parameter.
// This extends the flexibility and makes GoRequest fun and clean! You can use GoRequest in whatever style you love!
//
// For example:
//
//    func printBody(resp gorequest.Response, body string, errs []error){
//      fmt.Println(resp.Status)
//    }
//    gorequest.New().Get("http://www..google.com").End(printBody)
//
func (h *XPHttpImpl) End(callback ...func(response HTTPResponse, body string, errs []error)) (HTTPResponse, []*http.Cookie, string, []error) {
	var bytesCallback []func(response HTTPResponse, body []byte, errs []error)
	if len(callback) > 0 {
		bytesCallback = []func(response HTTPResponse, body []byte, errs []error){
			func(response HTTPResponse, body []byte, errs []error) {
				callback[0](response, string(body), errs)
			},
		}
	}

	resp, cookies, body, errs := h.EndBytes(bytesCallback...)
	bodyString := string(body)

	return resp, cookies, bodyString, errs
}

// EndBytes should be used when you want the body as bytes. The callbacks work the same way as with `End`, except that a byte array is used instead of a string.
func (h *XPHttpImpl) EndBytes(callback ...func(response HTTPResponse, body []byte, errs []error)) (HTTPResponse, []*http.Cookie, []byte, []error) {
	var (
		errs    []error
		resp    HTTPResponse
		body    []byte
		cookies []*http.Cookie
	)

	for {
		resp, cookies, body, errs = h.getResponseBytes()
		if errs != nil {
			return nil, nil, nil, errs
		}
		if h.isRetryableRequest(resp) {
			resp.Header.Set("Retry-Count", strconv.Itoa(h.Retryable.Attempt))
			break
		}
	}

	respCallback := *resp
	if len(callback) != 0 {
		callback[0](&respCallback, body, h.Errors)
	}
	return resp, cookies, body, nil
}

func (h *XPHttpImpl) isRetryableRequest(resp HTTPResponse) bool {
	if h.Retryable.Enable && h.Retryable.Attempt < h.Retryable.RetryCount && contains(resp.StatusCode, h.Retryable.RetryableStatus) {
		time.Sleep(h.Retryable.RetryTime)
		h.Retryable.Attempt++
		return false
	}
	return true
}

func contains(respStatus int, statuses []int) bool {
	for _, status := range statuses {
		if status == respStatus {
			return true
		}
	}
	return false
}

// EndStruct should be used when you want the body as a struct. The callbacks work the same way as with `End`, except that a struct is used instead of a string.
func (h *XPHttpImpl) EndStruct(v interface{}, callback ...func(response HTTPResponse, v interface{}, body []byte, errs []error)) (HTTPResponse, []*http.Cookie, []byte, []error) {
	resp, cookies, body, errs := h.EndBytes()
	if errs != nil {
		return nil, nil, body, errs
	}
	err := json.Unmarshal(body, &v)
	if err != nil {
		h.Errors = append(h.Errors, err)
		return resp, nil, body, h.Errors
	}
	respCallback := *resp
	if len(callback) != 0 {
		callback[0](&respCallback, v, body, h.Errors)
	}
	return resp, cookies, body, nil
}

func (h *XPHttpImpl) getResponseBytes() (HTTPResponse, []*http.Cookie, []byte, []error) {
	var (
		req     *http.Request
		oresp   *http.Response
		resp    HTTPResponse
		cookies []*http.Cookie
		err     error
	)
	// check whether there is an error. if yes, return all errors
	if len(h.Errors) != 0 {
		return nil, nil, nil, h.Errors
	}
	// check if there is forced type
	switch h.ForceType {
	case "json", "form", "xml", "text", "multipart":
		h.TargetType = h.ForceType
	// If forcetype is not set, check whether user set Content-Type header.
	// If yes, also bounce to the correct supported TargetType automatically.
	default:
		for k, v := range HTTP_ContentTypes {
			if h.Headers["Content-Type"] == v {
				h.TargetType = k
			}
		}
	}

	// if slice and map get mixed, let's bounce to rawstring
	if len(h.Data) != 0 && len(h.SliceData) != 0 {
		h.BounceToRawString = true
	}

	// Make Request
	req, err = h.MakeRequest()
	if err != nil {
		h.Errors = append(h.Errors, err)
		return nil, nil, nil, h.Errors
	}

	// Set Transport
	if !HttpDisableTransportSwap {
		h.Client.Transport = h.Transport
	}

	// Log details of this request
	if h.Debug {
		dump, err := httputil.DumpRequest(req, true)
		h.logger.SetPrefix("[http] ")
		if err != nil {
			h.logger.Println("Error:", err)
		} else {
			h.logger.Printf("HTTP Request: %s", string(dump))
		}
	}

	// Display CURL command line
	if h.CurlCommand {
		curl, err := getCurlCommand(req)
		h.logger.SetPrefix("[curl] ")
		if err != nil {
			h.logger.Println("Error:", err)
		} else {
			h.logger.Printf("CURL command line: %s", curl)
		}
	}

	// Send request
	resp, err = h.Client.Do(req)
	if err != nil {
		h.Errors = append(h.Errors, err)
		return nil, nil, nil, h.Errors
	}
	oresp   = resp
	cookies = oresp.Cookies()
	defer resp.Body.Close()

	// Log details of this response
	if h.Debug {
		dump, err := httputil.DumpResponse(resp, true)
		if nil != err {
			h.logger.Println("Error:", err)
		} else {
			h.logger.Printf("HTTP Response: %s", string(dump))
		}
	}

	body, _ := ioutil.ReadAll(resp.Body)
	// Reset resp.Body so it can be use again
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return resp, cookies, body, nil
}

func (h *XPHttpImpl) MakeRequest() (*http.Request, error) {
	var (
		req *http.Request
		err error
	)

	switch h.Method {
	case HTTP_POST, HTTP_PUT, HTTP_PATCH:
		if h.TargetType == "json" {
			// If-case to give support to json array. we check if
			// 1) Map only: send it as json map from s.Data
			// 2) Array or Mix of map & array or others: send it as rawstring from s.RawString
			var contentJson []byte
			if h.BounceToRawString {
				contentJson = []byte(h.RawString)
			} else if len(h.Data) != 0 {
				contentJson, _ = json.Marshal(h.Data)
			} else if len(h.SliceData) != 0 {
				contentJson, _ = json.Marshal(h.SliceData)
			}
			contentReader := bytes.NewReader(contentJson)
			req, err = http.NewRequest(h.Method, h.Url, contentReader)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "application/json")
		} else if h.TargetType == "form" || h.TargetType == "form-data" || h.TargetType == "urlencoded" {
			var contentForm []byte
			if h.BounceToRawString || len(h.SliceData) != 0 {
				contentForm = []byte(h.RawString)
			} else {
				formData := changeMapToURLValues(h.Data)
				contentForm = []byte(formData.Encode())
			}
			contentReader := bytes.NewReader(contentForm)
			req, err = http.NewRequest(h.Method, h.Url, contentReader)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		} else if h.TargetType == "text" {
			req, err = http.NewRequest(h.Method, h.Url, strings.NewReader(h.RawString))
			req.Header.Set("Content-Type", "text/plain")
		} else if h.TargetType == "xml" {
			req, err = http.NewRequest(h.Method, h.Url, strings.NewReader(h.RawString))
			req.Header.Set("Content-Type", "application/xml")
		} else if h.TargetType == "multipart" {

			var buf bytes.Buffer
			mw := multipart.NewWriter(&buf)

			if h.BounceToRawString {
				fieldName, ok := h.Headers["data_fieldname"]
				if !ok {
					fieldName = "data"
				}
				fw, _ := mw.CreateFormField(fieldName)
				fw.Write([]byte(h.RawString))
			}

			if len(h.Data) != 0 {
				formData := changeMapToURLValues(h.Data)
				for key, values := range formData {
					for _, value := range values {
						fw, _ := mw.CreateFormField(key)
						fw.Write([]byte(value))
					}
				}
			}

			if len(h.SliceData) != 0 {
				fieldName, ok := h.Headers["json_fieldname"]
				if !ok {
					fieldName = "data"
				}
				// copied from CreateFormField() in mime/multipart/writer.go
				head := make(textproto.MIMEHeader)
				fieldName = strings.Replace(strings.Replace(fieldName, "\\", "\\\\", -1), `"`, "\\\"", -1)
				head.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, fieldName))
				head.Set("Content-Type", "application/json")
				fw, _ := mw.CreatePart(head)
				contentJson, err := json.Marshal(h.SliceData)
				if err != nil {
					return nil, err
				}
				fw.Write(contentJson)
			}

			// add the files
			if len(h.FileData) != 0 {
				for _, file := range h.FileData {
					fw, _ := mw.CreateFormFile(file.Fieldname, file.Filename)
					fw.Write(file.Data)
				}
			}

			// close before call to FormDataContentType ! otherwise its not valid multipart
			mw.Close()

			req, err = http.NewRequest(h.Method, h.Url, &buf)
			req.Header.Set("Content-Type", mw.FormDataContentType())
		} else {
			// let's return an error instead of an nil pointer exception here
			return nil, NewErrors("TargetType '" + h.TargetType + "' could not be determined")
		}
	case "":
		return nil, NewErrors("No method specified")
	default:
		req, err = http.NewRequest(h.Method, h.Url, nil)
		if err != nil {
			return nil, err
		}
	}

	for k, v := range h.Headers {
		req.Header.Set(k, v)
		// Setting the host header is a special case, see this issue: https://github.com/golang/go/issues/7682
		if strings.EqualFold(k, "host") {
			req.Host = v
		}
	}
	// Add all querystring from Query func
	q := req.URL.Query()
	for k, v := range h.QueryData {
		for _, vv := range v {
			q.Add(k, vv)
		}
	}
	req.URL.RawQuery = q.Encode()

	// Add basic auth
	if h.BasicAuth != struct{ Username, Password string }{} {
		req.SetBasicAuth(h.BasicAuth.Username, h.BasicAuth.Password)
	}

	// Add cookies
	for _, cookie := range h.ReqCookies {
		if cookie != nil {
			req.AddCookie(cookie)
		}
	}

	return req, nil
}

// AsCurlCommand returns a string representing the runnable `curl' command
// version of the request.
func (h *XPHttpImpl) AsCurlCommand() (string, error) {
	req, err := h.MakeRequest()
	if err != nil {
		return "", err
	}
	cmd, err := getCurlCommand(req)
	if err != nil {
		return "", err
	}
	return cmd.String(), nil
}

type curlCommand struct {
	slice []string
}

// append appends a string to the CurlCommand
func (c *curlCommand) append(newSlice ...string) {
	c.slice = append(c.slice, newSlice...)
}

// String returns a ready to copy/paste command
func (c *curlCommand) String() string {
	return strings.Join(c.slice, " ")
}

// nopCloser is used to create a new io.ReadCloser for req.Body
type nopCloser struct {
	io.Reader
}

func bashEscape(str string) string {
	return `'` + strings.Replace(str, `'`, `'\''`, -1) + `'`
}

func (nopCloser) Close() error { return nil }

// GetCurlCommand returns a CurlCommand corresponding to an http.Request
func getCurlCommand(req *http.Request) (*curlCommand, error) {
	command := curlCommand{}

	command.append("curl")

	command.append("-X", bashEscape(req.Method))

	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = nopCloser{bytes.NewBuffer(body)}
		bodyEscaped := bashEscape(string(body))
		command.append("-d", bodyEscaped)
	}

	var keys []string

	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		command.append("-H", bashEscape(fmt.Sprintf("%s: %s", k, strings.Join(req.Header[k], " "))))
	}

	command.append(bashEscape(req.URL.String()))

	return &command, nil
}