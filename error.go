package http

import "fmt"

type HttpError struct {
	Status      int    `json:"-"`
	Type        string `json:"error_type"`
	Code        int    `json:"error_code"`
	Description string `json:"error_description"`
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("%d: [%d] %s", e.Status, e.Code, e.Description)
}
