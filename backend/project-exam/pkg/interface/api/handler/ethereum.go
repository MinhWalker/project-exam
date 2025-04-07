package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/project-exam/pkg/interface/api/response"
	"github.com/project-exam/pkg/interface/validator"
	"github.com/project-exam/pkg/usecase"
)

// EthereumHandler handles Ethereum-related HTTP requests
type EthereumHandler struct {
	useCase   usecase.EthereumUseCase
	validator *validator.EthereumValidator
}

// NewEthereumHandler creates a new EthereumHandler
func NewEthereumHandler(useCase usecase.EthereumUseCase, validator *validator.EthereumValidator) *EthereumHandler {
	return &EthereumHandler{
		useCase:   useCase,
		validator: validator,
	}
}

// GetAddressInfo handles the request to get Ethereum data for a specific address
func (h *EthereumHandler) GetAddressInfo(c *gin.Context) {
	address := c.Param("address")

	// Validate Ethereum address
	if !h.validator.IsValidAddress(address) {
		response.BadRequest(c, "Invalid Ethereum address format", nil)
		return
	}

	// Format address (ensures proper casing, etc.)
	address = h.validator.FormatAddress(address)

	// Get address information from use case
	addressInfo, err := h.useCase.GetAddressInfo(c.Request.Context(), address)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	// Format and return successful response
	formattedResponse := response.FormatAddressInfo(addressInfo)
	response.Success(c, formattedResponse)
}

// HealthCheck handles health check requests
func (h *EthereumHandler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
