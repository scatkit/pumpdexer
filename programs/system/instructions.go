package system

import (
	"bytes"
	"fmt"
  "encoding/binary"

	gg_binary "github.com/gagliardetto/binary"
	scatsol "github.com/scatkit/pumpdexer/solana"
)

// Turning variables into IDs
const (
	// Create a new account
	Instruction_CreateAccount uint32 = iota

	// Assign account to a program
	Instruction_Assign

	// Transfer lamports
	Instruction_Transfer

	// Create a new account at an address derived from a base pubkey and a seed
	Instruction_CreateAccountWithSeed

	// Consumes a stored nonce, replacing it with a successor
	Instruction_AdvanceNonceAccount

	// Withdraw funds from a nonce account
	Instruction_WithdrawNonceAccount

	// Drive state of Uninitalized nonce account to Initialized, setting the nonce value
	Instruction_InitializeNonceAccount

	// Change the entity authorized to execute nonce instructions on the account
	Instruction_AuthorizeNonceAccount

	// Allocate space in a (possibly new) account without funding
	Instruction_Allocate

	// Allocate space for and assign an account at an address derived from a base public key and a seed
	Instruction_AllocateWithSeed

	// Assign account to a program based on a seed
	Instruction_AssignWithSeed

	// Transfer lamports from a derived address
	Instruction_TransferWithSeed
)

// Instruction that returns a name of a program instruction by ID
func InstructionIDToName(id uint32) string {
	switch id {
	case Instruction_CreateAccount:
		return "CreateAccount"
	case Instruction_Assign:
		return "Assign"
	case Instruction_Transfer:
		return "Transfer"
	case Instruction_CreateAccountWithSeed:
		return "CreateAccountWithSeed"
	case Instruction_AdvanceNonceAccount:
		return "AdvanceNonceAccount"
	case Instruction_WithdrawNonceAccount:
		return "WithdrawNonceAccount"
	case Instruction_InitializeNonceAccount:
		return "InitializeNonceAccount"
	case Instruction_AuthorizeNonceAccount:
		return "AuthorizeNonceAccount"
	case Instruction_Allocate:
		return "Allocate"
	case Instruction_AllocateWithSeed:
		return "AllocateWithSeed"
	case Instruction_AssignWithSeed:
		return "AssignWithSeed"
	case Instruction_TransferWithSeed:
		return "TransferWithSeed"
	default:
		return ""
	}
}

type Instruction struct {
  gg_binary.BaseVariant
}

var ProgramID scatsol.PublicKey = scatsol.SystemProgramID

func SetProgramID(pubkey scatsol.PublicKey){
  ProgramID = pubkey
  scatsol.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "System"

func init() {
	scatsol.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
} 

func registryDecodeInstruction(accounts []*scatsol.AccountMeta, data []byte) (interface{}, error) {
  inst, err := DecodeInstruction(accounts, data)
  if err != nil {
    return nil, err
  }
  return inst, nil
}

// Instruction DECODER itself
func DecodeInstruction(accounts []*scatsol.AccountMeta, data []byte) (*Instruction, error) {
	inst := new(Instruction)
	if err := gg_binary.NewBinDecoder(data).Decode(inst); err != nil {
		return nil, fmt.Errorf("unable to decode instruction: %w", err)
	}
	if v, ok := inst.Impl.(scatsol.AccountsSettable); ok {
		err := v.SetAccounts(accounts)
		if err != nil {
			return nil, fmt.Errorf("unable to set accounts for instruction: %w", err)
		}
	}
	return inst, nil
}

// Instructions methods ---
func (inst *Instruction) Accounts() (out []*scatsol.AccountMeta) { 
	return inst.Impl.(scatsol.AccountsGettable).GetAccounts() // AccountsGettable is an interface
}

func (inst *Instruction) Data() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := gg_binary.NewBinEncoder(buf).Encode(inst); err != nil {
		return nil, fmt.Errorf("unable to encode instruction: %w", err)
	}
	return buf.Bytes(), nil
}

func (inst *Instruction) ProgramID() scatsol.PublicKey {
	return ProgramID
}
// ---

var InstructionImplDef = gg_binary.NewVariantDefinition(
	gg_binary.Uint32TypeIDEncoding,
	[]gg_binary.VariantType{ {
			"Transfer", (*Transfer)(nil),
		},
	},
)

func (inst *Instruction) UnmarshalWithDecoder(decoder *gg_binary.Decoder) error {
	return inst.BaseVariant.UnmarshalBinaryVariant(decoder, InstructionImplDef)
}

func (inst Instruction) MarshalWithEncoder(encoder *gg_binary.Encoder) error {
	err := encoder.WriteUint32(inst.TypeID.Uint32(), binary.LittleEndian)
	if err != nil {
		return fmt.Errorf("unable to write variant type: %w", err)
	}
	return encoder.Encode(inst.Impl)
}


