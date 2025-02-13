package token
import(
  "fmt"
  "bytes"
  
  bin "github.com/gagliardetto/binary"
  "github.com/scatkit/pumpdexer/solana"
)

// Maximum number of multisignature signers (max N)
const MAX_SIGNERS = 11

var ProgramID solana.PublicKey = solana.TokenProgramID

func SetProgramID(pubkey solana.PublicKey) {
	ProgramID = pubkey
	solana.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "Token"

func init() {
	if !ProgramID.IsZero() {
		solana.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
	}
}

const(
  /*
  Initializes a new account to hold tokens. If this account is associated
	with the native mint then the token balance of the initialized account
	will be equal to the amount of SOL in the account. If this account is
	associated with another mint, that mint must be initialized before this
	command can succeed.
	
	The `InitializeAccount` instruction requires no signers and MUST be
	included within the same Transaction as the system program's
	`CreateAccount` instruction that creates the account being initialized.
	Otherwise another party can acquire ownership of the uninitialized
	account. 
  */
	Instruction_InitializeAccount uint8 = iota
  
  // Close an account by transferring all its SOL to the destination account.
	// Non-native accounts may only be closed if its token amount is zero.
  Instruction_CloseAccount
)

type Instruction struct {
	bin.BaseVariant
}

func (inst *Instruction) ProgramID() solana.PublicKey {
	return ProgramID
}

func (inst *Instruction) Accounts() (out []*solana.AccountMeta) {
	return inst.Impl.(solana.AccountsGettable).GetAccounts()
}

func (inst *Instruction) Data() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := bin.NewBinEncoder(buf).Encode(inst); err != nil {
		return nil, fmt.Errorf("unable to encode instruction: %w", err)
	}
	return buf.Bytes(), nil
}

var InstructionImplDef = bin.NewVariantDefinition(
	bin.Uint8TypeIDEncoding,
	[]bin.VariantType{
		{
      Name: "InitializeAccount",
			Type: (*InitializeAccount)(nil),
		},
  },
)

func (inst *Instruction) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	return inst.BaseVariant.UnmarshalBinaryVariant(decoder, InstructionImplDef)
}

func (inst Instruction) MarshalWithEncoder(encoder *bin.Encoder) error {
	err := encoder.WriteUint8(inst.TypeID.Uint8())
	if err != nil {
		return fmt.Errorf("unable to write variant type: %w", err)
	}
	return encoder.Encode(inst.Impl)
}

func registryDecodeInstruction(accounts []*solana.AccountMeta, data []byte) (interface{}, error) {
	inst, err := DecodeInstruction(accounts, data)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func DecodeInstruction(accounts []*solana.AccountMeta, data []byte) (*Instruction, error) {
	inst := new(Instruction)
	if err := bin.NewBinDecoder(data).Decode(inst); err != nil {
		return nil, fmt.Errorf("unable to decode instruction: %w", err)
	}
	if v, ok := inst.Impl.(solana.AccountsSettable); ok {
		err := v.SetAccounts(accounts)
		if err != nil {
			return nil, fmt.Errorf("unable to set accounts for instruction: %w", err)
		}
	}
	return inst, nil
}

