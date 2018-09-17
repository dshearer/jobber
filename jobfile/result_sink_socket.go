package jobfile

const _SOCKET_RESULT_SINK_NAME = "socket"

type SocketResultSink struct {
	Proto   string              `yaml:"proto"`
	Address string              `yaml:"address"`
	Data    ResultSinkDataParam `yaml:"data"`
}

func (self SocketResultSink) CheckParams() error {
	return nil
}

func (self SocketResultSink) String() string {
	return _SOCKET_RESULT_SINK_NAME
}

func (self SocketResultSink) Equals(other ResultSink) bool {
	otherSocket, ok := other.(*SocketResultSink)
	if !ok {
		return false
	}
	if otherSocket.Proto != self.Proto {
		return false
	}
	if otherSocket.Address != self.Address {
		return false
	}
	if otherSocket.Data != self.Data {
		return false
	}
	return true
}

func (self SocketResultSink) Handle(runRec RunRec) {
	runRecStr := SerializeRunRec(runRec, self.Data)
	GlobalRunRecServerRegistry.Push(self.Proto, self.Address, runRecStr)
}
