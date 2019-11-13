package stl

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
	}
}
