package stl

type IStl struct {
	Async      *XPAsyncImpl
	Config     *XPConfigImpl
	DateTime   *XPDateTimeImpl
	Encrypt    *XPEncryptImpl
	IPAddress  *XPIPImpl
	Number     *XPNumberImpl
	Queue      *XPQueueImpl
	String     *XPStringImpl
	Scheduler  *XPSchedulerImpl
}

func (instance *IStl) ConfigIpAddress(keyFileUrl, dataFileUrl string) {
	KeyFileUrl  = keyFileUrl
	DataFileUrl = dataFileUrl
}

func (instance *IStl) FilePath(path string) (filepath *XPFilePathImpl, err error) {
	return NewFilePath(path)
}

func (instance *IStl) FilePathCurrent() (filepath *XPFilePathImpl, err error) {
	return NewFilePathFromCurrentPath()
}

func (instance *IStl) IdGenerator(workerId int64) (idGenerator *XPIdGeneratorImpl, err error) {
	return NewIdGenerator(workerId)
}

func (instance *IStl) Http() *XPHttpImpl {
	return NewHttp()
}

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

func (instance *IStl) SpinLocker() *SpinLock {
	return &SpinLock{}
}

var IMP IStl

func init()  {
	IMP = IStl{
		Async:  NewAsync(),
		Config: NewXPConfig(nil),
		DateTime: &XPDateTimeImpl{},
		Encrypt: &XPEncryptImpl{},
		IPAddress: NewIPAddress(),
		Number:  &XPNumberImpl{},
		Queue: NewXPQueue(500),
		String: &XPStringImpl{},
		Scheduler: NewScheduler(),
	}
}
