package response

var errors = map[ErrCode]string{
	ErrCodeMalformedJSON:       "The JSON you provided was not well-formed or did not validate against our published format.",
	ErrCodeRequestBody:         "Request body error",
	ErrCodeLegalActionNotFound: "Legal action not found.",
	ErrCodeDeviceNotFound:      "Device [%s] not found.",
	ErrCodeResourceExists:      "Variable [%s] already exists.",
	ErrCodeResourceNotFound:    "Variable [%s] not found.",
	ErrCodeVariableNotWritable: "Variable [%s] not writable.",
	ErrCodeDeviceNotConnect:    "Device [%s] not connect.",
}

// !!! IMPORTANT PLEASE READ FIRST !!!
// You SHOULD add new code at the end of enum firstly.

var ErrMalformedJSON = &responseError{
	Code:    ErrCodeMalformedJSON,
	Message: errors[ErrCodeMalformedJSON],
}

var ErrRequestBody = &responseError{
	Code:    ErrCodeRequestBody,
	Message: errors[ErrCodeRequestBody],
}

var ErrLegalActionNotFound = &responseError{
	Code:    ErrCodeLegalActionNotFound,
	Message: errors[ErrCodeLegalActionNotFound],
}
