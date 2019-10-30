package structs

type CaseData struct {
	Endpoint   string
	ReqType    string
	ReqData    []byte
	RespStatus int
	RespData   []byte
}
