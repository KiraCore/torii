package handlers

type ServiceErr struct {
	Code    string `json:"code" example:"ERROR_CODE"`
	Message string `json:"message" example:"error description"`
}

var (
	errInternalServer = &ServiceErr{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "internal server error",
	}

	errBadRequest = &ServiceErr{
		Code:    "BAD_REQUEST",
		Message: "bad request",
	}
)
