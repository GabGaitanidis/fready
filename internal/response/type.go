package response

type SuccessResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}