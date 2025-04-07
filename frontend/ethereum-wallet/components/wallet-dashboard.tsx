"use client"

import { useState, useEffect } from "react"
import { ethers } from "ethers"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { WalletBalance } from "@/components/wallet-balance"
import { BlockchainInfo } from "@/components/blockchain-info"
import { AlertCircle } from "lucide-react"
import { getAddressInfo } from "@/lib/api"
import type { AddressInfo } from "@/types/ethereum"

export function WalletDashboard() {
  const [account, setAccount] = useState<string | null>(null)
  const [provider, setProvider] = useState<ethers.BrowserProvider | null>(null)
  const [addressInfo, setAddressInfo] = useState<AddressInfo | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Check if MetaMask is installed
  const isMetaMaskInstalled = typeof window !== "undefined" && window.ethereum !== undefined

  // Connect wallet function
  const connectWallet = async () => {
    setError(null)
    setIsLoading(true)

    try {
      if (!isMetaMaskInstalled) {
        throw new Error("MetaMask is not installed")
      }

      // Create a new provider
      const browserProvider = new ethers.BrowserProvider(window.ethereum)
      setProvider(browserProvider)

      // Request account access
      const accounts = await browserProvider.send("eth_requestAccounts", [])
      const address = accounts[0]
      setAccount(address)

      // Fetch address info from API
      await fetchAddressInfo(address)
    } catch (err) {
      console.error("Error connecting wallet:", err)
      setError(err instanceof Error ? err.message : "Failed to connect wallet")
    } finally {
      setIsLoading(false)
    }
  }

  // Fetch address info from API
  const fetchAddressInfo = async (address: string) => {
    try {
      const data = await getAddressInfo(address)
      setAddressInfo(data)
    } catch (err) {
      console.error("Error fetching address info:", err)
      setError(err instanceof Error ? err.message : "Failed to fetch address information")
    }
  }

  // Disconnect wallet function
  const disconnectWallet = () => {
    setAccount(null)
    setProvider(null)
    setAddressInfo(null)
  }

  // Listen for account changes
  useEffect(() => {
    if (isMetaMaskInstalled && provider) {
      const handleAccountsChanged = async (accounts: string[]) => {
        if (accounts.length === 0) {
          // User disconnected their wallet
          disconnectWallet()
        } else if (accounts[0] !== account) {
          // User switched accounts
          setAccount(accounts[0])
          await fetchAddressInfo(accounts[0])
        }
      }

      window.ethereum.on("accountsChanged", handleAccountsChanged)

      return () => {
        window.ethereum.removeListener("accountsChanged", handleAccountsChanged)
      }
    }
  }, [provider, account])

  return (
    <Card className="w-full max-w-3xl mx-auto">
      <CardHeader>
        <CardTitle>Wallet Connection</CardTitle>
        <CardDescription>Connect your Ethereum wallet to view your balance and blockchain information</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {error && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {!account ? (
          <Button onClick={connectWallet} disabled={isLoading || !isMetaMaskInstalled} className="w-full">
            {isLoading ? "Connecting..." : isMetaMaskInstalled ? "Connect Wallet" : "MetaMask Not Installed"}
          </Button>
        ) : (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Connected Account</p>
                <p className="text-xs text-muted-foreground break-all">{account}</p>
              </div>
              <Button variant="outline" onClick={disconnectWallet}>
                Disconnect
              </Button>
            </div>

            {addressInfo && (
              <>
                <div className="grid gap-6 md:grid-cols-2">
                  <WalletBalance balance={addressInfo.balance} />
                  <BlockchainInfo
                    gasPrice={addressInfo.gasPrice}
                    currentBlock={addressInfo.currentBlock}
                    timestamp={addressInfo.timestamp}
                  />
                </div>
              </>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

