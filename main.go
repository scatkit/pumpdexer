package main

import (
	"context"
	"fmt"
  "encoding/binary"
  "bytes"
	//"github.com/davecgh/go-spew/spew"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)
 

type RaydiumLiquidityV4Structure struct{
  Status               uint64
  Nonce                uint64
	MaxOrder             uint64
	Depth                uint64
	BaseDecimal          uint64
	QuoteDecimal         uint64
	State                uint64
	ResetFlag            uint64
	MinSize              uint64
	VolMaxCutRatio       uint64
	AmountWaveRatio      uint64
	BaseLotSize          uint64
	QuoteLotSize         uint64
	MinPriceMultiplier   uint64
	MaxPriceMultiplier   uint64
	SystemDecimalValue   uint64
	MinSeparateNumerator uint64
	MinSeparateDenominator uint64
	TradeFeeNumerator    uint64
	TradeFeeDenominator  uint64
	PnlNumerator         uint64
	PnlDenominator       uint64
	SwapFeeNumerator     uint64
	SwapFeeDenominator   uint64
	BaseNeedTakePnl      uint64
	QuoteNeedTakePnl     uint64
	QuoteTotalPnl        uint64
	BaseTotalPnl         uint64
	PoolOpenTime         uint64
	PunishPcAmount       uint64
	PunishCoinAmount     uint64
	OrderbookToInitTime  uint64
	SwapBaseInAmount     [16]byte // u128 as a 16-byte array
	SwapQuoteOutAmount   [16]byte // u128 as a 16-byte array
	SwapBase2QuoteFee    uint64
	SwapQuoteInAmount    [16]byte // u128 as a 16-byte array
	SwapBaseOutAmount    [16]byte // u128 as a 16-byte array
	SwapQuote2BaseFee    uint64
	BaseVault            solana.PublicKey // 32-byte public key
	QuoteVault           solana.PublicKey // 32-byte public key
	BaseMint             solana.PublicKey // 32-byte public key
	QuoteMint            solana.PublicKey // 32-byte public key
	LpMint               solana.PublicKey // 32-byte public key
	OpenOrders           solana.PublicKey // 32-byte public key
	MarketId             solana.PublicKey // 32-byte public key
	MarketProgramId      solana.PublicKey // 32-byte public key
	TargetOrders         solana.PublicKey // 32-byte public key
	WithdrawQueue        solana.PublicKey // 32-byte public key
	LpVault              solana.PublicKey // 32-byte public key
	Owner                solana.PublicKey // 32-byte public key
	LpReserve            uint64
	Padding              [3]uint64 // Padding for alignment
}

func getPoolInfo(data []byte){
  var liqState RaydiumLiquidityV4Structure
  reader := bytes.NewReader(data)
  if err := binary.Read(reader, binary.LittleEndian, &liqState); err != nil{
    fmt.Errorf("cant read binary data: %v\n",err)
  }
  
  fmt.Printf("Status: %v\n", liqState.Status)
  //fmt.Printf("Nonce:", liqState.nonce.());
  //fmt.Printf("Max Order:", liqState.maxOrder.());
  //fmt.Printf("Depth:", liqState.depth.());
  //fmt.Printf("Base Decimal:", liqState.baseDecimal.());
  //fmt.Printf("Quote Decimal:", liqState.quoteDecimal.());
  //fmt.Printf("State:", liqState.state.());
  //fmt.Printf("Reset Flag:", liqState.resetFlag.());
  //fmt.Printf("Min Size:", liqState.minSize.());
  //fmt.Printf("Vol Max Cut Ratio:", liqState.volMaxCutRatio.());
  //fmt.Printf("Amount Wave Ratio:", liqState.amountWaveRatio.());
  //fmt.Printf("Base Lot Size:", liqState.baseLotSize.());
  //fmt.Printf("Quote Lot Size:", liqState.quoteLotSize.());
  //fmt.Printf("Min Price Multiplier:", liqState.minPriceMultiplier.());
  //fmt.Printf("Max Price Multiplier:", liqState.maxPriceMultiplier.());
  //fmt.Printf("System Decimal Value:", liqState.systemDecimalValue.());
  //fmt.Printf("Min Separate Numerator:", liqState.minSeparateNumerator.());
  //fmt.Printf("Min Separate Denominator:", liqState.minSeparateDenominator.());
  //fmt.Printf("Trade Fee Numerator:", liqState.tradeFeeNumerator.());
  //fmt.Printf("Trade Fee Denominator:", liqState.tradeFeeDenominator.());
  //fmt.Printf("PnL Numerator:", liqState.pnlNumerator.());
  //fmt.Printf("PnL Denominator:", liqState.pnlDenominator.());
  //fmt.Printf("Swap Fee Numerator:", liqState.swapFeeNumerator.());
  //fmt.Printf("Swap Fee Denominator:", liqState.swapFeeDenominator.());
  //fmt.Printf("Base Need Take PnL:", liqState.baseNeedTakePnl.());
  //fmt.Printf("Quote Need Take PnL:", liqState.quoteNeedTakePnl.());
  //fmt.Printf("Quote Total PnL:", liqState.quoteTotalPnl.());
  //fmt.Printf("Base Total PnL:", liqState.baseTotalPnl.());
  //fmt.Printf("Pool Open Time:", liqState.poolOpenTime.());
  //fmt.Printf("Punish Pc Amount:", liqState.punishPcAmount.());
  //fmt.Printf("Punish Coin Amount:", liqState.punishCoinAmount.());
  //fmt.Printf("Orderbook To Init Time:", liqState.orderbookToInitTime.())
  //fmt.Printf("Swap Base In Amount:", liqState.swapBaseInAmount)
  //fmt.Printf("Swap Quote Out Amount:", liqState.swapQuoteOutAmount)
  //fmt.Printf("Swap Base2Quote Fee:", liqState.swapBase2QuoteFee)
  //fmt.Printf("Swap Quote In Amount:", liqState.swapQuoteInAmount)
  //fmt.Printf("Swap Base Out Amount:", liqState.swapBaseOutAmount)
  //fmt.Printf("Swap Quote2Base Fee:", liqState.swapQuote2BaseFee)

  //// For PubicKey fields, you can directly  them as they are:
  //fmt.Printf("Base Vault:", liqState.baseVault.toString());
  //fmt.Printf("Quote Vault:", liqState.quoteVault.toString());
  //fmt.Printf("Base Mint:", liqState.baseMint.toString());
  //fmt.Printf("Quote Mint:", liqState.quoteMint.toString());
  //fmt.Printf("LP Mint:", liqState.lpMint.toString());
  //fmt.Printf("Open Orders:", liqState.openOrders.toString());
  //fmt.Printf("Market ID:", liqState.marketId.toString());
  //fmt.Printf("Market Program ID:", liqState.marketProgramId.toString());
  //fmt.Printf("Target Orders:", liqState.targetOrders.toString());
  //fmt.Printf("Withdraw Queue:", liqState.withdrawQueue.toString());
  //fmt.Printf("LP Vault:", liqState.lpVault.toString());
  //fmt.Printf("Owner:", liqState.owner.toString());
  //fmt.Printf("LP Reserve:", liqState.lpReserve.());
}

func main() {
	client := rpc.New("https://api.mainnet-beta.solana.com")
	poolID := solana.MustPublicKeyFromBase58("zcdAw3jpcqEY8JYVxNVMqs2cU35cyDdy4ot7V8edNhz")
	resp, err := client.GetAccountInfo(context.TODO(), poolID)
	if err != nil {
		fmt.Errorf("Error: %v", err)
	}
  
  out := resp.GetBinary()
  getPoolInfo(out) 
}
