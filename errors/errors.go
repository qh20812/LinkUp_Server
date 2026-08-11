package errors

// AppError is a structured error with code, message, and optional params.
type AppError struct {
	Code    string         `json:"error"`
	Message string         `json:"message"`
	Params  map[string]any `json:"params,omitempty"`
}

func (e *AppError) Error() string     { return e.Message }
func (e *AppError) HasParams() bool   { return len(e.Params) > 0 }
func (e *AppError) GetCode() string   { return e.Code }
func (e *AppError) GetParams() map[string]any { return e.Params }

// New creates a static AppError (no params).
func New(code string) *AppError {
	return &AppError{Code: code, Message: Messages[code]}
}

// Newf creates a dynamic AppError with params.
func Newf(code string, params map[string]any) *AppError {
	msg := RenderTemplate(Messages[code], params)
	return &AppError{Code: code, Message: msg, Params: params}
}
