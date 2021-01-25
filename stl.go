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
	"github.com/xpsuper/stl/srvmanager"
	"github.com/xpsuper/stl/taskbus"
	"image"
	"io"
)

type IStl struct {
	Array      *XPArrayImpl
	Async      *XPAsyncImpl
	Config     *XPConfigImpl
	DateTime   *XPDateTimeImpl
	Encrypt    *XPEncryptImpl
	IPAddress  *XPIPImpl
	Number     *XPNumberImpl
	Queue      *XPQueueImpl
	Regexp     *XPRegexpImpl
	String     *XPStringImpl
	Scheduler  *XPSchedulerImpl
	Zip        *XPZipImpl
	Jwt        *jwt.XPJwtImpl
}

//Adapter
func (instance *IStl) AdapterDecode(input, output interface{}) error {
	return adapter.WeakDecode(input, output)
}

func (instance *IStl) AdapterDecodeByTag(input, output interface{}, tag string) error {
	return adapter.WeakDecodeByTag(input, output, tag)
}

//IpAddress
func (instance *IStl) ConfigIpAddress(keyFileUrl, dataFileUrl string) {
	KeyFileUrl  = keyFileUrl
	DataFileUrl = dataFileUrl
}

//Interface Deep Copy
func (instance *IStl) DeepCopy(src interface{}) *XPDeepCPImpl {
	return DeepCopy(src)
}

//Dispatcher
func (instance *IStl) Dispatcher(cnt int) (*dispatcher.Dispatcher, error)  {
	return dispatcher.NewDispatcher(cnt)
}

//FilePath
func (instance *IStl) FilePath(path string) (filepath *XPFilePathImpl, err error) {
	return NewFilePath(path)
}

func (instance *IStl) FilePathCurrent() (filepath *XPFilePathImpl, err error) {
	return NewFilePathFromCurrentPath()
}

//IdGenerator
func (instance *IStl) IdGenerator(workerId int64) (idGenerator *XPIdGeneratorImpl, err error) {
	return NewIdGenerator(workerId)
}

//Helper
func (instance *IStl) Helper(v interface{}) helper.Helper {
	return helper.Chain(v)
}

func (instance *IStl) HelperLazy(v interface{}) helper.Helper {
	return helper.LazyChain(v)
}

//Http
func (instance *IStl) Http() *XPHttpImpl {
	return NewHttp()
}

//Json
func (instance *IStl) JsonValid(json string) bool {
	return Valid(json)
}

func (instance *IStl) JsonDataValid(json []byte) bool {
	return ValidBytes(json)
}

func (instance *IStl) JsonParse(json string) JsonItem {
	return Parse(json)
}

func (instance *IStl) JsonDataParse(json []byte) JsonItem {
	return ParseBytes(json)
}

func (instance *IStl) JsonGet(json, path string) JsonItem {
	return Get(json, path)
}

//MemoryCache
func (instance *IStl) Cache(config *memorycache.Configuration) *memorycache.Cache {
	return memorycache.NewCache(config)
}

//OrderMap
func (instance *IStl) OrderMap(less OrderMapKeyLess) *OrderMap {
	return NewOrderMap(less)
}

//Promise
func (instance *IStl) Promise(executor func(resolve func(interface{}), reject func(error))) *XPPromiseImpl {
	return NewPromise(executor)
}

func (instance *IStl) Resolve(resolution interface{}) *XPPromiseImpl {
	return ResolvePromise(resolution)
}

func (instance *IStl) Reject(err error) *XPPromiseImpl {
	return RejectPromise(err)
}

func (instance *IStl) All(promises ...*XPPromiseImpl) *XPPromiseImpl {
	return All(promises ...)
}

func (instance *IStl) Race(promises ...*XPPromiseImpl) *XPPromiseImpl {
	return Race(promises ...)
}

//SpinLocker
func (instance *IStl) SpinLocker() *SpinLock {
	return &SpinLock{}
}

//Canvas
func (instance *IStl) Canvas(width, height int) *canvas.Context {
	return canvas.NewContext(width, height)
}

func (instance *IStl) CanvasForImage(img image.Image) *canvas.Context {
	return canvas.NewContextForImage(img)
}

func (instance *IStl) CanvasForRGBA(rgba *image.RGBA) *canvas.Context {
	return canvas.NewContextForRGBA(rgba)
}

//ServiceManager
func (instance *IStl) ServiceBind(fn func()) {
	srvmanager.InitManage()
	srvmanager.Bind(fn)
}

func (instance *IStl) ServiceWait() {
	srvmanager.Wait()
}

//TaskBus
func (instance *IStl) TaskBusChain(stack taskbus.Tasks, firstArgs ...interface{}) ([]interface{}, error) {
	return taskbus.Chain(stack, firstArgs ...)
}

func (instance *IStl) TaskBusMax(stack taskbus.Taskier) (taskbus.Results, error) {
	return taskbus.Max(stack)
}

func (instance *IStl) TaskBusAll(stack taskbus.Taskier) (taskbus.Results, error) {
	return taskbus.All(stack)
}

//TaskTicker
func (instance *IStl) TaskTicker(scanInterval int, execOnStart bool) *TickerTasks {
	return NewTicker(scanInterval, execOnStart)
}

//Evaluate
func (instance *IStl) Eval(expression string, parameter interface{}, opts ...eval.Language) (interface{}, error) {
	return eval.Evaluate(expression, parameter, opts ...)
}

//HtmlParser
func (instance *IStl) HtmlParser(r io.Reader) (*htmlparser.Node, error) {
	return htmlparser.Parse(r)
}

//ExcelParser
func (instance *IStl) ExcelParser(filePath string, container interface{}) error {
	return excel.UnmarshalXLSX(filePath, container)
}

var IMP IStl

func init()  {
	IMP = IStl{
		Array:  &XPArrayImpl{},
		Async:  NewAsync(),
		Config: NewXPConfig(nil),
		DateTime: &XPDateTimeImpl{},
		Encrypt: &XPEncryptImpl{},
		IPAddress: NewIPAddress(),
		Number:  &XPNumberImpl{},
		Queue: NewXPQueue(500),
		Regexp: &XPRegexpImpl{},
		String: &XPStringImpl{},
		Scheduler: NewScheduler(),
		Zip: &XPZipImpl{},
		Jwt: &jwt.XPJwtImpl{},
	}
}
