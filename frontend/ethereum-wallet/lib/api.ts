import type { AddressInfo } from "@/types/ethereum"

export async function getAddressInfo(address: string): Promise<AddressInfo> {
  try {
    const response = await fetch(`/api/ethereum/${address}`)

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.message || "Failed to fetch address information")
    }

    const data = await response.json()
    return data.data
  } catch (error) {
    console.error("API Error:", error)
    throw error
  }
}

