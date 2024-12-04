package rpc
import (
  "fmt"
  stdjson "encoding/json"
  "github.com/scatkit/pumpdexer/solana"  
  "math/big"
)

type Context struct{
  Slot uint64 `json:"slot"`
}
 
type RPCContext struct{
  Context Context `json:"context,omitempty"`
}

type AccountInfoResult struct{
  RPCContext
  Value *Account `json:"value"`
}

type Account struct{
  Lamports uint64 `json:"lamports"`
  Owner solana.PublicKey `json:"owner"` 
  // Data associated with the account. Format: json/binary -> depends on the encoding parameter
  Data *DataBytesOrJSON `json:"data"`
  Executable bool `json:"executable"`
  RentEpoch *big.Int `json:"rentEpoch"` 
}

type DataBytesOrJSON struct{
  rawDataEncoding solana.EncodingType // string
  asDecodedBinary solana.Data
  asJSON          stdjson.RawMessage
}

func (wrap *DataBytesOrJSON) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || (len(data) == 4 && string(data) == "null") {
		// TODO: is this an error?
		return nil
	}

	firstChar := data[0]

	switch firstChar {
	// Check if first character is `[`, standing for a JSON array.
	case '[':
		// It's base64 (or similar)
		{
			err := wrap.asDecodedBinary.UnmarshalJSON(data)
			if err != nil {
				return err
			}
			wrap.rawDataEncoding = wrap.asDecodedBinary.Encoding
		}
	case '{':
		// It's JSON, most likely.
		// TODO: is it always JSON???
		{
			// Store raw bytes, and unmarshal on request.
			wrap.asJSON = data
			wrap.rawDataEncoding = solana.EncodingJsonParsed
		}
	default:
		return fmt.Errorf("unknown kind: %v", data)
	}

	return nil
}

func (dt *DataBytesOrJSON) GetBinary() []byte{
  return dt.asDecodedBinary.Content
}

type TransactionSignature struct{
  Err interface{} `json:"err"` // err if failed else nil if succeeded
  Memo *string    `json:"signature"` 
  Signature *solana.Signature `json:"signature"`
  Slot uint64  `json:"slot,omitempty"`
  BlockTime *solana.UnixTimeSeconds `json:"blockTime,omitempty"`
}
 
type GetBalanceResult struct {
	RPCContext
	Value uint64 `json:"value"`
 }

type UiTokenAmount struct{
  Amount    string      `json:"amount"`
  Decimals  uint8       `json:"decimals"`
  UiAmount  *float64    `json:"uiAmount"`
  UiAmountString string `json:"uiAmountString"`
}

type CommitmentType string

const (
  // The node will query the most recent block confirmed by supermajority
  CommitmentFinalized CommitmentType = "finalized"
  // The node will query the most recent block that has been voted on by supermajority of the cluster
  CommitmentConfirmed CommitmentType = "confirmed"
  // The node will qurty its most recent block (might still be skipped by the cluster)
  CommitmentProcessed CommitmentType = "processed"
)

type DataSlice struct{
  Offset *uint64 `json:"offset,omitempty"`
  Length *uint64 `json:"length,omitempty"`
}
