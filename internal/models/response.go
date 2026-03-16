package models

// APIResponse is the standard API response structure
type APIResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Details string `json:"details"`
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
