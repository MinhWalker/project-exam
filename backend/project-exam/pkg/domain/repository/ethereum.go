package repository

import (
	"context"
	"math/big"

	"github.com/project-exam/pkg/domain/entity"
)

// EthereumRepository defines the interface for interacting with the Ethereum blockchain
type EthereumRepository interface {
	// GetGasPrice returns the current gas price from the Ethereum network
	GetGasPrice(ctx context.Context) (*big.Int, error)

	// GetCurrentBlock returns the latest block number
	GetCurrentBlock(ctx context.Context) (uint64, error)

	// GetAddressBalance returns the balance for the given address
	GetAddressBalance(ctx context.Context, address string) (*big.Int, error)

	// GetAddressInfo retrieves all required information for an address in a single call
	GetAddressInfo(ctx context.Context, address string) (*entity.AddressInfo, error)

	// Close closes any connections to the Ethereum network
	Close()
}
