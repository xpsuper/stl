package stl

type IStl struct {
	Async   *XPAsyncImpl
	Config  *XPConfigImpl
	Encrypt *XPEncryptImpl
	IPAddress *XPIPImpl
	String  *XPStringImpl
}

var IMP IStl

func init()  {
	IMP = IStl{
		Async:  NewAsync(),
		Config: NewXPConfig(nil),
		Encrypt: &XPEncryptImpl{},
		IPAddress: NewIPAddress(),
		String: &XPStringImpl{},
	}
}
