package stl

import (
	"github.com/xpsuper/stl/adapter"
	"github.com/xpsuper/stl/canvas"
	"github.com/xpsuper/stl/dispatcher"
	"github.com/xpsuper/stl/eval"
	"github.com/xpsuper/stl/excel"
	"github.com/xpsuper/stl/helper"
	"github.com/xpsuper/stl/htmlparser"
	"github.com/xpsuper/stl/jwt"
	"github.com/xpsuper/stl/memorycache"
	"github.com/xpsuper/stl/objassigner"
	"github.com/xpsuper/stl/srvmanager"
	"github.com/xpsuper/stl/taskbus"
	"image"
	"io"
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
}

func IsEmpty(data interface{}) bool {
	if data == nil {
		return true
	}
	switch data.(type) {
	case string:
		return S.IsEmpty(data.(string))
	case []interface{}:
		return len(data.([]interface{})) == 0
	}
	return false
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
