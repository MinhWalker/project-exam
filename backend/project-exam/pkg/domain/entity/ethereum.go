package entity

import (
	"math/big"
	"time"
)

// AddressInfo represents the core data structure for Ethereum address information
type AddressInfo struct {
	Address      string
	GasPrice     GasPrice
	CurrentBlock uint64
	Balance      Balance
	Timestamp    time.Time
}

// GasPrice represents gas price information
type GasPrice struct {
	Wei  *big.Int
	Gwei float64
}

// Balance represents balance information for an Ethereum address
type Balance struct {
	Wei   *big.Int
	Ether float64
}
