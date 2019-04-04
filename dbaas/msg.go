package dbaas

type OutuputMsg interface {
	String() string
}

type OutuputMsgDebug string

func (e OutuputMsgDebug) String() string {
	return string(e)
}

type OutuputMsgError string

func (e OutuputMsgError) String() string {
	return string(e)
}
