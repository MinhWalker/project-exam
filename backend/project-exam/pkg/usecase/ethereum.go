package usecase

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/project-exam/pkg/domain/entity"
	"github.com/project-exam/pkg/domain/repository"
)

// EthereumUseCase defines the interface for Ethereum application business rules
type EthereumUseCase interface {
	GetAddressInfo(ctx context.Context, address string) (*entity.AddressInfo, error)
}

// ethereumUseCase implements the EthereumUseCase interface
type ethereumUseCase struct {
	repo repository.EthereumRepository
}

// NewEthereumUseCase creates a new EthereumUseCase
func NewEthereumUseCase(repo repository.EthereumRepository) EthereumUseCase {
	return &ethereumUseCase{
		repo: repo,
	}
}

// GetAddressInfo retrieves Ethereum data for a specific address
func (uc *ethereumUseCase) GetAddressInfo(ctx context.Context, address string) (*entity.AddressInfo, error) {
	// Create channels for concurrent operations
	gasPriceCh := make(chan *big.Int)
	blockNumberCh := make(chan uint64)
	balanceCh := make(chan *big.Int)
	errCh := make(chan error, 3)

	// Get gas price concurrently
	go func() {
		gasPrice, err := uc.repo.GetGasPrice(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to get gas price: %w", err)
			return
		}
		gasPriceCh <- gasPrice
	}()

	// Get block number concurrently
	go func() {
		blockNumber, err := uc.repo.GetCurrentBlock(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to get block number: %w", err)
			return
		}
		blockNumberCh <- blockNumber
	}()

	// Get balance concurrently
	go func() {
		balance, err := uc.repo.GetAddressBalance(ctx, address)
		if err != nil {
			errCh <- fmt.Errorf("failed to get balance: %w", err)
			return
		}
		balanceCh <- balance
	}()

	// Wait for results or errors
	var gasPrice *big.Int
	var blockNumber uint64
	var balance *big.Int

	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		case err := <-errCh:
			return nil, err
		case gasPrice = <-gasPriceCh:
			continue
		case blockNumber = <-blockNumberCh:
			continue
		case balance = <-balanceCh:
			continue
		}
	}

	// Convert Wei to Gwei for gas price (1 Gwei = 10^9 Wei)
	gwei := new(big.Float).Quo(
		new(big.Float).SetInt(gasPrice),
		new(big.Float).SetInt(big.NewInt(1e9)),
	)

	// Convert Wei to Ether (1 Ether = 10^18 Wei)
	ether := new(big.Float).Quo(
		new(big.Float).SetInt(balance),
		new(big.Float).SetInt(big.NewInt(1e18)),
	)

	// Convert to float64 for easier JSON serialization
	gweiFloat, _ := gwei.Float64()
	etherFloat, _ := ether.Float64()

	// Create and return the AddressInfo entity
	return &entity.AddressInfo{
		Address: address,
		GasPrice: entity.GasPrice{
			Wei:  gasPrice,
			Gwei: gweiFloat,
		},
		CurrentBlock: blockNumber,
		Balance: entity.Balance{
			Wei:   balance,
			Ether: etherFloat,
		},
		Timestamp: time.Now(),
	}, nil
}
