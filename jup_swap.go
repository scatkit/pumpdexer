package main
import(
  "log"
  "io"
  "fmt"
  stdjson "encoding/json"
  "errors"
  "net/http"
  "github.com/davecgh/go-spew/spew"
)

const apiBaseURL = "https://quote-api.jup.ag/v6"

type swapMode string

const (
  ExactIn   swapMode = "ExactIn"
  ExactOut  swapMode = "ExactOut"
)

func (sm *swapMode) UnmarshalJSON(data []byte) error{
  var output string
  if err := stdjson.Unmarshal(data, &output); err != nil{
    return err
  }

  switch output{
  case string(ExactIn), string(ExactOut):
    *sm = swapMode(output)
    return nil
  default:
    return errors.New("invalid swapMode value")
  }
  
}

type SwapInfo struct{
  AmmKey      string  `json:"ammKey"`
  Label       string  `json:"label"`
  InputMint   string  `json:"inputMint"`
  OutputMint  string  `json:"outputMint"`
  InAmount    string  `json:"inAmount"`
  OutAmount   string  `json:"outAmount"`
  FeeAmount   string  `json:"feeAmount"`
  FeeMint     string  `json:"feeMint"`
}

type Route struct{
  SwapInfo  SwapInfo `json:"swapInfo"`
  Precent   int32     `json:"perecent"`   
}

type RoutePlan []Route


type QuoteResponse struct{
  InputMint           string  `json:"inputMint"`
  InAmount            string  `json:"inAmount"`
  OutputMint          string  `json:"outputMint"`
  OutAmount           string  `json:"outAmount"`
  OtherAmountTreshold string  `json:"otherAmountTreshold"`
  SwapMode            swapMode`json:"swapMode"` 
  SlippageBps         int32   `json:"slippageBps"`
  PriceImpactPct      string  `json:"priceImpactPct"`
  RoutePlan           RoutePlan `json:"routePlan"`
  ContextSlot         uint64  `json:"contextSlot"`
  TimeTaken           float64 `json:"timeTaken"`
}

func main(){
  inputMint := "So11111111111111111111111111111111111111112"
  outputMint := "4Cnk9EPnW5ixfLZatCPJjDB1PUtcRpVVgTQukm9epump"
  amount := "5000900"
  
  url := fmt.Sprintf("%s/quote?inputMint=%s&outputMint=%s&amount=%s", apiBaseURL, inputMint, outputMint, amount)
  
  resp, err := http.Get(url)
  if err != nil{
    log.Fatalf("Failed to fetch data: %v", err)
  }
  defer resp.Body.Close()
  
  body, err := io.ReadAll(resp.Body)
  if err != nil{
    log.Fatalf("Failed to read response: %v",err)
  }
  
  if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: API returned status %d: %s", resp.StatusCode, string(body))
	}
  
  var jupResponse QuoteResponse
  if err := stdjson.Unmarshal(body, &jupResponse); err != nil{
    log.Fatalf("Failed to parse JSON: %v", err)
  }
  
  spew.Dump(jupResponse)
}

