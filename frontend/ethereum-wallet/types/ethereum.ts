export interface GasPrice {
  wei: string
  gwei: number
}

export interface Balance {
  wei: string
  ether: number
}

export interface AddressInfo {
  address: string
  gasPrice: GasPrice
  currentBlock: number
  balance: Balance
  timestamp: string
}

// Extend the Window interface to include ethereum
declare global {
  interface Window {
    ethereum?: any
  }
}

