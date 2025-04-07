import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { GasPrice } from "@/types/ethereum"
import { formatDate } from "@/lib/utils"

interface BlockchainInfoProps {
  gasPrice: GasPrice
  currentBlock: number
  timestamp: string
}

export function BlockchainInfo({ gasPrice, currentBlock, timestamp }: BlockchainInfoProps) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-lg">Network Information</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <div>
            <p className="text-sm font-medium">Gas Price</p>
            <p className="text-lg">{gasPrice.gwei} Gwei</p>
            <p className="text-xs text-muted-foreground">{gasPrice.wei} wei</p>
          </div>
          <div>
            <p className="text-sm font-medium">Current Block</p>
            <p className="text-lg">{currentBlock.toLocaleString()}</p>
          </div>
          <div>
            <p className="text-sm font-medium">Last Updated</p>
            <p className="text-sm">{formatDate(timestamp)}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

