package models

// APIResponse is the standard API response structure
type APIResponse struct {
	Status  bool        `json:"status" example:"true"`
	Message string      `json:"message" example:"Success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code    string `json:"code" example:"ERROR_CODE"`
	Details string `json:"details" example:"Detailed error message"`
}

// SuccessResponse creates a success response
func SuccessResponse(message string, data interface{}) *APIResponse {
	return &APIResponse{
		Status:  true,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse creates an error response
func ErrorResponse(message string, code string, details string) *APIResponse {
	return &APIResponse{
		Status:  false,
		Message: message,
		Error: &ErrorInfo{
			Code:    code,
			Details: details,
		},
	}
}
