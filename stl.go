package stl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/xpsuper/stl/adapter"
	"github.com/xpsuper/stl/canvas"
	"github.com/xpsuper/stl/dispatcher"
	"github.com/xpsuper/stl/eval"
	"github.com/xpsuper/stl/excel"
	"github.com/xpsuper/stl/helper"
	"github.com/xpsuper/stl/htmlparser"
	"github.com/xpsuper/stl/jwt"
	"github.com/xpsuper/stl/linq"
	"github.com/xpsuper/stl/memorycache"
	"github.com/xpsuper/stl/objassigner"
	"github.com/xpsuper/stl/srvmanager"
	"github.com/xpsuper/stl/taskbus"
	"image"
	"io"
	"reflect"
	"unsafe"
)

var (
	S         *XPStringImpl
	N         *XPNumberImpl
	Random    *XPRandomImpl
	DateTime  *XPDateTimeImpl
	Array     *XPArrayImpl
	Async     *XPAsyncImpl
	Config    *XPConfigImpl
	Encrypt   *XPEncryptImpl
	IPAddress *XPIPImpl
	Queue     *XPQueueImpl
	Regexp    *XPRegexpImpl
	Scheduler *XPSchedulerImpl
	Zip       *XPZipImpl
	Jwt       *jwt.XPJwtImpl
	IdCard    *XPIdCardImpl
)

func init() {
	S = &XPStringImpl{}
	N = &XPNumberImpl{}
	Random = &XPRandomImpl{}
	Array = &XPArrayImpl{}
	DateTime = &XPDateTimeImpl{}
	Async = NewAsync()
	Config = NewXPConfig(nil)
	Encrypt = &XPEncryptImpl{}
	IPAddress = NewIPAddress()
	Queue = NewXPQueue(500)
	Regexp = &XPRegexpImpl{}
	Scheduler = NewScheduler()
	Zip = &XPZipImpl{}
	Jwt = &jwt.XPJwtImpl{}
	IdCard = &XPIdCardImpl{}
}

// IsEmpty 判断是否为空
func IsEmpty(data interface{}) bool {
	if data == nil {
		return true
	}
	dataRef := reflect.ValueOf(data)
	for dataRef.Kind() == reflect.Ptr {
		dataRef = dataRef.Elem()
	}
	switch dataRef.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice:
		return dataRef.Len() == 0
	case reflect.String:
		return S.IsEmpty(data.(string))
	}
	//switch data.(type) {
	//case string:
	//	return S.IsEmpty(data.(string))
	//case []interface{}:
	//	return len(data.([]interface{})) == 0
	//}
	return false
}

// ToJson 转为 Json 字符串
func ToJson(data interface{}) string {
	jsonByte, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Marshal with error: %+v\n", err)
		return "{}"
	}
	return string(jsonByte)
}

func FromJson(jsonStr string, data interface{}) error {
	return json.Unmarshal([]byte(jsonStr), data)
}

// ToJsonIndent 转为 Json 格式化字符串
func ToJsonIndent(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Marshal with error: %+v\n", err)
		return "{}"
	}
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "    ")
	return out.String()
}

// ToMap 转为 map
func ToMap(data interface{}) map[string]interface{} {
	dataRef := reflect.ValueOf(data)
	for dataRef.Kind() == reflect.Ptr {
		dataRef = dataRef.Elem()
	}
	var jsonByte []byte
	var err error
	if dataRef.Kind() == reflect.String {
		jsonByte = []byte(dataRef.String())
	} else {
		jsonByte, err = json.Marshal(data)
	}
	if err != nil {
		fmt.Printf("Marshal with error: %+v\n", err)
		return nil
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(jsonByte, &m)
	if err != nil {
		fmt.Printf("Unmarshal with error: %+v\n", err)
		return nil
	}
	return m
}

// ToArray 转为 slice
func ToArray(data interface{}) []interface{} {
	dataRef := reflect.ValueOf(data)
	for dataRef.Kind() == reflect.Ptr {
		dataRef = dataRef.Elem()
	}
	var jsonByte []byte
	var err error
	if dataRef.Kind() == reflect.String {
		jsonByte = []byte(dataRef.String())
	} else {
		jsonByte, err = json.Marshal(data)
	}
	if err != nil {
		fmt.Printf("Marshal with error: %+v\n", err)
		return nil
	}
	m := make([]interface{}, 0)
	err = json.Unmarshal(jsonByte, &m)
	if err != nil {
		fmt.Printf("Unmarshal with error: %+v\n", err)
		return nil
	}
	return m
}

// StringToBytes string 转为 []byte
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// BytesToString []byte 转为 string
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// MapValue 获取 map 的值
func MapValue[K comparable, V any](m map[K]interface{}, key K, def V) V {
	if m == nil {
		return def
	}
	if v, has := m[key]; has {
		if value, ok := v.(V); ok {
			return value
		}
	}
	return def
}

// MapKeys 获取 map 的所有 key
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))

	var i int
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

// MapValues 获取 map 的所有 value
func MapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, len(m))

	var i int
	for _, v := range m {
		values[i] = v
		i++
	}

	return values
}

// AdapterDecode 对象转换适配器
func AdapterDecode(input, output interface{}) error {
	return adapter.WeakDecode(input, output)
}

// AdapterDecodeByTag 对象转换-自定义标签适配器
func AdapterDecodeByTag(input, output interface{}, tag string) error {
	return adapter.WeakDecodeByTag(input, output, tag)
}

// ConfigIpAddress 纯真IP库配置
func ConfigIpAddress(keyFileUrl, dataFileUrl string) {
	KeyFileUrl = keyFileUrl
	DataFileUrl = dataFileUrl
}

// Dispatcher 任务分发器
func Dispatcher(cnt int) (*dispatcher.Dispatcher, error) {
	return dispatcher.NewDispatcher(cnt)
}

// FilePath 获取文件路径对象
func FilePath(path string) (filepath *XPFilePathImpl, err error) {
	return NewFilePath(path)
}

func FilePathCurrent() (filepath *XPFilePathImpl, err error) {
	return NewFilePathFromCurrentPath()
}

// IdGenerator 唯一ID生成器
func IdGenerator(workerId int64) (idGenerator *XPIdGeneratorImpl, err error) {
	return NewIdGenerator(workerId)
}

// LINQ 集成查询
func LINQ(source interface{}) linq.LinQuery {
	return linq.From(source)
}

// Helper 帮助类
func Helper(v interface{}) helper.Helper {
	return helper.Chain(v)
}

func HelperLazy(v interface{}) helper.Helper {
	return helper.LazyChain(v)
}

// Http HttpClient
func Http() *XPHttpImpl {
	return NewHttp()
}

// JsonValid Json对象验证器
func JsonValid(json string) bool {
	return Valid(json)
}

func JsonDataValid(json []byte) bool {
	return ValidBytes(json)
}

func JsonParse(json string) JsonItem {
	return Parse(json)
}

func JsonDataParse(json []byte) JsonItem {
	return ParseBytes(json)
}

func JsonGet(json, path string) JsonItem {
	return Get(json, path)
}

// Cache MemoryCache
func Cache(config *memorycache.Configuration) *memorycache.Cache {
	return memorycache.NewCache(config)
}

// MapWithOrder 排序的map
func MapWithOrder(less OrderMapKeyLess) *OrderMap {
	return NewOrderMap(less)
}

// Promise 异步执行
func Promise(executor func(resolve func(interface{}), reject func(error))) *XPPromiseImpl {
	return NewPromise(executor)
}

func Resolve(resolution interface{}) *XPPromiseImpl {
	return ResolvePromise(resolution)
}

func Reject(err error) *XPPromiseImpl {
	return RejectPromise(err)
}

func All(promises ...*XPPromiseImpl) *XPPromiseImpl {
	return PromiseAll(promises...)
}

func Race(promises ...*XPPromiseImpl) *XPPromiseImpl {
	return PromiseRace(promises...)
}

// SpinLocker 自旋锁
func SpinLocker() *SpinLock {
	return &SpinLock{}
}

// Canvas 画布
func Canvas(width, height int) *canvas.Context {
	return canvas.NewContext(width, height)
}

func CanvasForImage(img image.Image) *canvas.Context {
	return canvas.NewContextForImage(img)
}

func CanvasForRGBA(rgba *image.RGBA) *canvas.Context {
	return canvas.NewContextForRGBA(rgba)
}

// ServiceBind ServiceManager 服务管理器
func ServiceBind(fn func()) {
	srvmanager.InitManage()
	srvmanager.Bind(fn)
}

// ServiceWait ServiceManager 服务管理器
func ServiceWait() {
	srvmanager.Wait()
}

// TaskBusChain TaskBus 异步任务总线
func TaskBusChain(stack taskbus.Tasks, firstArgs ...interface{}) ([]interface{}, error) {
	return taskbus.Chain(stack, firstArgs...)
}

// TaskBusMax TaskBus 异步任务总线
func TaskBusMax(stack taskbus.Taskier) (taskbus.Results, error) {
	return taskbus.Max(stack)
}

// TaskBusAll TaskBus 异步任务总线
func TaskBusAll(stack taskbus.Taskier) (taskbus.Results, error) {
	return taskbus.All(stack)
}

// TaskTicker 定时任务
func TaskTicker(scanInterval int, execOnStart bool) *TickerTasks {
	return NewTicker(scanInterval, execOnStart)
}

// Eval	执行表达式
func Eval(expression string, parameter interface{}, opts ...eval.Language) (interface{}, error) {
	return eval.Evaluate(expression, parameter, opts...)
}

func EvalFull(extensions ...eval.Language) eval.Language {
	return eval.Full(extensions...)
}

func EvalConstant(name string, value interface{}) eval.Language {
	return eval.Constant(name, value)
}

func EvalFunction(name string, value interface{}) eval.Language {
	return eval.Function(name, value)
}

// HtmlParser 解析html
func HtmlParser(r io.Reader) (*htmlparser.Node, error) {
	return htmlparser.Parse(r)
}

// ExcelParser 解析Excel
func ExcelParser(filePath string, container interface{}) error {
	return excel.UnmarshalXLSX(filePath, container)
}

// ObjDeepCopy Interface Deep Copy
func ObjDeepCopy(src interface{}) *XPDeepCPImpl {
	return DeepCopy(src)
}

// ObjAssign Interface Assign
func ObjAssign(target, source interface{}) error {
	return objassigner.Assign(target, source)
}

// ObjAssignWithOption Interface Assign With Option
func ObjAssignWithOption(target, source interface{}, opt objassigner.Option) error {
	return objassigner.AssignWithOption(target, source, opt)
}
