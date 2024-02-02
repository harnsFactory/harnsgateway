package response

type ErrCode int

const (
	_                                ErrCode = 10000 + iota
	ErrCodeMalformedJSON                     // 10001
	ErrCodeRequestBody                       // 10002
	ErrCodeResourceExists                    // 10003
	ErrCodeResourceNotFound                  // 10004
	ErrCodeLegalActionNotFound               // 10005
	ErrCodeBooleanInvalid                    // 10006
	ErrCodeInteger16Invalid                  // 10007
	ErrCodeInteger32Invalid                  // 10008
	ErrCodeInteger64Invalid                  // 10009
	ErrCodeFloat32Invalid                    // 10008
	ErrCodeFloat64Invalid                    // 10009
	ErrCodeDeviceNotFound                    // 10010
	ErrCodeVariableNotWritable               // 10011
	ErrCodeDeviceNotConnect                  // 10012
	ErrCodeDeviceOperatorUnSupported         // 10013
)

// !!! IMPORTANT PLEASE READ FIRST !!!
// You SHOULD add new code at the end, and append comment of number
// Meanwhile, the corresponding error message SHOULD be appended in response.errors
// The order MUST be consistent between them
