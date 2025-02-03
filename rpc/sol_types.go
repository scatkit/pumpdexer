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


func (a *AccountInfoResult) GetBinary() []byte{
  if a == nil{
    return nil
  }
  if a.Value == nil{
    return nil
  }
  if a.Value.Data == nil{
    return nil
  }
  return a.Value.Data.GetBinary()
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
  Memo *string    `json:"memo"` // meta data to a transaction
  Signature     solana.Signature `json:"signature"`
  Slot uint64  `json:"slot,omitempty"` // the slot number where the transaction was confirmed
  BlockTime *solana.UnixTimeSeconds `json:"blockTime,omitempty"` // eastimated time when transactio was processed
  ConfirmationStatus ConfirmationStatusType `json:"confirmationStatus,omitempty"`
}

type ConfirmationStatusType string

const (
	ConfirmationStatusProcessed ConfirmationStatusType = "processed"
	ConfirmationStatusConfirmed ConfirmationStatusType = "confirmed"
	ConfirmationStatusFinalized ConfirmationStatusType = "finalized"
)
 
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

type TransactionOpts struct{
  Encoding solana.EncodingType        `json:"encoding,omitemprt"`
  SkipPreflight bool                  `json:"skipPreflight,omitempty"`
  PreflightCommitment CommitmentType  `json:"preflightCommitment,omitempty"`
  MaxRetries *uint                    `json:"maxRetries"`
  MinContextSlot *uint64              `json:"minContextSlot"`
}

//instructions that happen within the given insutruction
type InnerInstruction struct{
  Index uint16 `json:"index"`
  Instructions []CompiledInstruction `json:"instructions"`
}

type CompiledInstruction struct {
	// Index into the message.accountKeys array indicating the program account that executes this instruction.
	// NOTE: it is actually a uint8, but using a uint16 because uint8 is treated as a byte everywhere,
	// and that can be an issue.
	ProgramIDIndex uint16 `json:"programIdIndex"`
	// List of ordered indices into the message.accountKeys array indicating which accounts to pass to the program.
	// NOTE: it is actually a []uint8, but using a uint16 because []uint8 is treated as a []byte everywhere,
	// and that can be an issue.
	Accounts []uint16 `json:"accounts"`
	// The program input data encoded in a base-58 string.
	Data solana.Base58 `json:"data"`
	StackHeight uint16 `json:"stackHeight"`
}

type TransactionMeta struct{
  // Error if transaction failed, null if transaction succeeded.
  Err interface{} `json:"err"`
  // Fee this transaction was charged
	Fee uint64 `json:"fee"`
  // Array of uint64 account balances from before the transaction was processed
	PreBalances []uint64 `json:"preBalances"`
  // Array of uint64 account balances after the transaction was processed
	PostBalances []uint64 `json:"postBalances"`
  // List of inner instructions  or omitted if inner instruction recording
	// was not yet enabled during this transaction.
	InnerInstructions []InnerInstruction `json:"innerInstructions"`
  // List of token balances from before the transaction was processed
	// or omitted if token balance recording was not yet enabled during this transaction
	PreTokenBalances []TokenBalance `json:"preTokenBalances"`
  // List of token balances from after the transaction was processed
	// or omitted if token balance recording was not yet enabled during this transaction
	PostTokenBalances []TokenBalance `json:"postTokenBalances"`
  // Array of string log messages or omitted if log message
	// recording was not yet enabled during this transaction
	LogMessages []string `json:"logMessages"`
  Rewards []BlockReward `json:"rewards"`
	LoadedAddresses LoadedAddresses `json:"loadedAddresses"`
	ReturnData ReturnData `json:"returnData"`
	ComputeUnitsConsumed *uint64 `json:"computeUnitsConsumed"`
}

type LoadedAddresses struct {
	ReadOnly solana.PublicKeySlice `json:"readonly"`
	Writable solana.PublicKeySlice `json:"writable"`
}

type ReturnData struct {
	ProgramId solana.PublicKey `json:"programId"`
	Data      solana.Data      `json:"data"`
}

type TokenBalance struct{
  // Index of the account in which the token balance is provided for.
	AccountIndex uint16 `json:"accountIndex"`
  // Pubkey of token balance's owner.
	Owner *solana.PublicKey `json:"owner,omitempty"`
  // Pubkey of token program.
	ProgramId *solana.PublicKey `json:"programId,omitempty"`
  // Pubkey of the token's mint.
	Mint          solana.PublicKey `json:"mint"`
	UiTokenAmount *UiTokenAmount   `json:"uiTokenAmount"`
}

type RewardType string

const (
	RewardTypeFee     RewardType = "Fee"
	RewardTypeRent    RewardType = "Rent"
	RewardTypeVoting  RewardType = "Voting"
	RewardTypeStaking RewardType = "Staking"
)

type BlockReward struct {
	// The public key of the account that received the reward.
	Pubkey solana.PublicKey `json:"pubkey"`
	// Number of reward lamports credited or debited by the account, as a i64.
	Lamports int64 `json:"lamports"`
	// Account balance in lamports after the reward was applied.
	PostBalance uint64 `json:"postBalance"`
	// Type of reward: "Fee", "Rent", "Voting", "Staking".
	RewardType RewardType `json:"rewardType"`
	// Vote account commission when the reward was credited,
	// only present for voting and staking rewards.
	Commission *uint8 `json:"commission,omitempty"`
}
