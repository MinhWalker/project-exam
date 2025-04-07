package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/project-exam/pkg/domain/entity"
)

// Response is the standard response format for the API
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// AddressInfoResponse is the response format for address information
type AddressInfoResponse struct {
	Address      string           `json:"address"`
	GasPrice     GasPriceResponse `json:"gasPrice"`
	CurrentBlock uint64           `json:"currentBlock"`
	Balance      BalanceResponse  `json:"balance"`
	Timestamp    string           `json:"timestamp"`
}

// GasPriceResponse is the response format for gas price
type GasPriceResponse struct {
	Wei  string  `json:"wei"`
	Gwei float64 `json:"gwei"`
}

// BalanceResponse is the response format for balance
type BalanceResponse struct {
	Wei   string  `json:"wei"`
	Ether float64 `json:"ether"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	c.JSON(statusCode, Response{
		Status:  "error",
		Message: message,
		Error:   errMsg,
	})
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response{
		Status: "success",
		Data:   data,
	})
}

// FormatAddressInfo formats an AddressInfo entity into an API response
func FormatAddressInfo(info *entity.AddressInfo) AddressInfoResponse {
	return AddressInfoResponse{
		Address: info.Address,
		GasPrice: GasPriceResponse{
			Wei:  info.GasPrice.Wei.String(),
			Gwei: info.GasPrice.Gwei,
		},
		CurrentBlock: info.CurrentBlock,
		Balance: BalanceResponse{
			Wei:   info.Balance.Wei.String(),
			Ether: info.Balance.Ether,
		},
		Timestamp: info.Timestamp.Format(time.RFC3339),
	}
}

// Success400 sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string, err error) {
	NewErrorResponse(c, http.StatusBadRequest, message, err)
}

// Success200 sends a 200 OK response
func Success(c *gin.Context, data interface{}) {
	NewSuccessResponse(c, http.StatusOK, data)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, err error) {
	NewErrorResponse(c, http.StatusInternalServerError, "Internal server error", err)
}

// TooManyRequests sends a 429 Too Many Requests response
func TooManyRequests(c *gin.Context) {
	NewErrorResponse(c, http.StatusTooManyRequests, "Rate limit exceeded. Try again later.", nil)
}
