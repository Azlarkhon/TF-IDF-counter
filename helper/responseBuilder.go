package helper

type Response struct {
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
	IsSuccess bool   `json:"is_success"`
}

func NewErrorResponse(errMessage string) Response {
	return Response{
		Data:      nil,
		Error:     errMessage,
		IsSuccess: false,
	}
}

func NewSuccessResponse(data any) Response {
	return Response{
		Data:      data,
		Error:     "",
		IsSuccess: true,
	}
}
