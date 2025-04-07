package persistence

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/project-exam/pkg/domain/entity"
	"github.com/project-exam/pkg/domain/repository"
	"github.com/project-exam/pkg/infrastructure/ethereum"
)

// ethereumRepository implements the EthereumRepository interface
type ethereumRepository struct {
	client *ethereum.Client
}

// NewEthereumRepository creates a new EthereumRepository
func NewEthereumRepository(client *ethereum.Client) repository.EthereumRepository {
	return &ethereumRepository{
		client: client,
	}
}

// GetGasPrice returns the current gas price from the Ethereum network
func (r *ethereumRepository) GetGasPrice(ctx context.Context) (*big.Int, error) {
	ctx, cancel := r.client.TimeoutCtx(ctx)
	defer cancel()

	return r.client.EthClient.SuggestGasPrice(ctx)
}

// GetCurrentBlock returns the latest block number
func (r *ethereumRepository) GetCurrentBlock(ctx context.Context) (uint64, error) {
	ctx, cancel := r.client.TimeoutCtx(ctx)
	defer cancel()

	return r.client.EthClient.BlockNumber(ctx)
}

// GetAddressBalance returns the balance for the given address
func (r *ethereumRepository) GetAddressBalance(ctx context.Context, address string) (*big.Int, error) {
	ctx, cancel := r.client.TimeoutCtx(ctx)
	defer cancel()

	ethAddress := common.HexToAddress(address)
	return r.client.EthClient.BalanceAt(ctx, ethAddress, nil) // nil = latest block
}

// GetAddressInfo retrieves all required information for an address in a single call
// This is an optimization that can be used instead of making three separate calls
func (r *ethereumRepository) GetAddressInfo(ctx context.Context, address string) (*entity.AddressInfo, error) {
	// Create channels for concurrent operations
	gasPriceCh := make(chan *big.Int)
	blockNumberCh := make(chan uint64)
	balanceCh := make(chan *big.Int)
	errCh := make(chan error, 3)

	// Get gas price concurrently
	go func() {
		gasPrice, err := r.GetGasPrice(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to get gas price: %w", err)
			return
		}
		gasPriceCh <- gasPrice
	}()

	// Get block number concurrently
	go func() {
		blockNumber, err := r.GetCurrentBlock(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to get block number: %w", err)
			return
		}
		blockNumberCh <- blockNumber
	}()

	// Get balance concurrently
	go func() {
		balance, err := r.GetAddressBalance(ctx, address)
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

// Close closes the connection to the Ethereum client
func (r *ethereumRepository) Close() {
	r.client.Close()
}
