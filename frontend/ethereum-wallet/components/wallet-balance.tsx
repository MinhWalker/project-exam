import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { Balance } from "@/types/ethereum"

interface WalletBalanceProps {
  balance: Balance
}

export function WalletBalance({ balance }: WalletBalanceProps) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-lg">Wallet Balance</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{balance.ether} ETH</div>
        <p className="text-xs text-muted-foreground mt-1">{balance.wei} wei</p>
      </CardContent>
    </Card>
  )
}

