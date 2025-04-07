package ethereum

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/project-exam/pkg/infrastructure/config"
)

// Client wraps the Ethereum client with application-specific configuration
type Client struct {
	EthClient  *ethclient.Client
	Config     *config.EthereumConfig
	TimeoutCtx func(context.Context) (context.Context, context.CancelFunc)
}

// NewClient creates a new Ethereum client
func NewClient(cfg *config.EthereumConfig) (*Client, error) {
	// Create a timeout context for connecting to the Ethereum node
	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	// Connect to the Ethereum node
	client, err := ethclient.DialContext(ctx, cfg.RPCURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		EthClient: client,
		Config:    cfg,
		TimeoutCtx: func(ctx context.Context) (context.Context, context.CancelFunc) {
			return context.WithTimeout(ctx, cfg.RequestTimeout)
		},
	}, nil
}

// Close closes the connection to the Ethereum client
func (c *Client) Close() {
	if c.EthClient != nil {
		c.EthClient.Close()
	}
}
