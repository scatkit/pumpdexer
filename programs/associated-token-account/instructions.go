package associatedtokenaccount

import(
  "fmt"
  "github.com/scatkit/pumpdexer/solana"
  bin "github.com/gagliardetto/binary"
)

// This program defines the convention and provides the mechanism for mapping
// the user's wallet address to the associated token accounts they hold.
var ProgramID solana.PublicKey = solana.SPLAssociatedTokenAccountProgramID

func SetProgramID(pubkey solana.PublicKey) {
	ProgramID = pubkey
	solana.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

func init() {
	solana.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

func registryDecodeInstruction(accounts []*solana.AccountMeta, data []byte) (interface{}, error){
  inst, err := DecodeInstruction(accounts, data)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

type Instruction struct {
	bin.BaseVariant
}

// the programID instruction acts on
func (inst *Instruction) ProgramID() solana.PublicKey{
  return ProgramID
}

// list of accounts the instructions require
func (inst *Instruction) Accounts() (out []*solana.AccountMeta) {
  return inst.Impl.(solana.AccountsGettable).GetAccounts()
}

// binary encoded transaction
func (inst *Instruction) Data() ([]byte, error){
  return []byte{}, nil
}

var InstructionImplDef = bin.NewVariantDefinition(
	bin.NoTypeIDEncoding, // NOTE: the associated-token-account program has no ID encoding.
	[]bin.VariantType{
    {
      Name: "Create",
      Type: (*Create)(nil), // passing the nil poiner
    },
  },
)

func (inst *Instruction) UnmarshalWithDecoder(decoder *bin.Decoder) error{
  return inst.BaseVariant.UnmarshalBinaryVariant(decoder, InstructionImplDef)
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
