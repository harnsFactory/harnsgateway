package response

type ErrCode int

const (
	_                          ErrCode = 10000 + iota
	ErrCodeMalformedJSON               // 10001
	ErrCodeRequestBody                 // 10002
	ErrCodeResourceExists              // 10003
	ErrCodeResourceNotFound            // 10004
	ErrCodeLegalActionNotFound         // 10005
)

// !!! IMPORTANT PLEASE READ FIRST !!!
// You SHOULD add new code at the end, and append comment of number
// Meanwhile, the corresponding error message SHOULD be appended in response.errors
// The order MUST be consistent between them
