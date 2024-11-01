package common

type ResponseResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewResponseResult(code int, message string, data interface{}) ResponseResult {
	return ResponseResult{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
