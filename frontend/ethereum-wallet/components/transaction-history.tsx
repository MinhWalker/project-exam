import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import type { Transaction } from "@/types/ethereum"
import { formatEther, shortenAddress, formatTimestamp } from "@/lib/utils"
import { ExternalLink } from "lucide-react"
import Link from "next/link"

interface TransactionHistoryProps {
  transactions: Transaction[]
}

export function TransactionHistory({ transactions }: TransactionHistoryProps) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-lg">Recent Transactions</CardTitle>
      </CardHeader>
      <CardContent>
        {transactions.length === 0 ? (
          <p className="text-center text-muted-foreground py-4">No transactions found</p>
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Hash</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Value</TableHead>
                  <TableHead className="hidden md:table-cell">From/To</TableHead>
                  <TableHead className="hidden md:table-cell">Time</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {transactions.map((tx) => (
                  <TableRow key={tx.hash}>
                    <TableCell className="font-medium">
                      <Link
                        href={`https://etherscan.io/tx/${tx.hash}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center hover:underline"
                      >
                        {shortenAddress(tx.hash)}
                        <ExternalLink className="h-3 w-3 ml-1" />
                      </Link>
                    </TableCell>
                    <TableCell>{tx.direction}</TableCell>
                    <TableCell>{formatEther(tx.value)} ETH</TableCell>
                    <TableCell className="hidden md:table-cell">
                      {tx.direction === "in" ? shortenAddress(tx.from) : shortenAddress(tx.to)}
                    </TableCell>
                    <TableCell className="hidden md:table-cell">{formatTimestamp(tx.timestamp)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

