package db

type MsgBackend interface {
	CreateMessage(meg *Message) error
	QueryMsg(id int) ([]Message, error)
}

var MsgDB MsgBackendImpl

func InitMsgBackend(mb MsgBackend) {
	if mb == nil {
		panic("不允许空实现")
	}
	MsgDB = MsgBackendImpl{proxy: mb}
}

type MsgBackendImpl struct {
	proxy MsgBackend
}

func (m MsgBackendImpl) CreateMessage(msg *Message) error {
	return m.proxy.CreateMessage(msg)
}

func (m MsgBackendImpl) QueryMsg(minId int) ([]Message, error) {
	return m.proxy.QueryMsg(minId)
}
