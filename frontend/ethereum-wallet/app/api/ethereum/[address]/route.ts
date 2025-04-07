import { type NextRequest, NextResponse } from "next/server"
import { ethers } from "ethers"

export async function GET(request: NextRequest, { params }: { params: { address: string } }) {
  try {
    const address = params.address

    // Validate Ethereum address
    if (!ethers.isAddress(address)) {
      return NextResponse.json({ message: "Invalid Ethereum address format" }, { status: 400 })
    }

    // In a real application, you would fetch this data from a blockchain provider
    // For this example, we'll use mock data that matches the example response
    const mockData = {
      address: address,
      gasPrice: {
        wei: "12000000000",
        gwei: 12.0,
      },
      currentBlock: 18782549,
      balance: {
        wei: "2500000000000000000",
        ether: 2.5,
      },
      timestamp: new Date().toISOString(),
    }

    return NextResponse.json(
      {
        status: "success",
        data: mockData,
      },
      { status: 200 },
    )
  } catch (error) {
    console.error("Error in API route:", error)
    return NextResponse.json({ status: "error", message: "Internal server error" }, { status: 500 })
  }
}

