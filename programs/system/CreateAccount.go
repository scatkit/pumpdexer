package system
 
import(
  "errors"
  "encoding/binary"
  "fmt"
  
  "github.com/scatkit/pumpdexer/solana"
  bin "github.com/gagliardetto/binary"
)

type CreateAccount struct{
  Lamports *uint64 // num of lamprots transfer to the new account
  Space *uint64 // num of bytes of memory to allocate
  Owner *solana.PublicKey // address of the program that will own the account
  // [0] = [WRITE, SIGNER] FundingAccount: `Funding account`
	// [1] = [WRITE, SIGNER] NewAccount: `New account`
	AccountMetaSlice solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewCreateAccountInstructionBuilder creates a new `CreateAccount` instruction builder.
func NewCreateAccountInstructionBuilder() *CreateAccount {
	nd := &CreateAccount{
		AccountMetaSlice: make(solana.AccountMetaSlice, 2),
	}
	return nd
}

// Number of lamports to transfer to the new account
func (inst *CreateAccount) SetLamports(amount uint64) *CreateAccount{
  inst.Lamports = &amount
  return inst
}

// Number of bytes of memory to allocate
func (inst *CreateAccount) SetSpace(space uint64) *CreateAccount {
	inst.Space = &space
	return inst
}

// Address of program that will own the new account
func (inst *CreateAccount) SetOwner(owner solana.PublicKey) *CreateAccount {
	inst.Owner = &owner
	return inst
}

// Funding account
func (inst *CreateAccount) SetFundingAccount(fundingAccount solana.PublicKey) *CreateAccount {
	inst.AccountMetaSlice[0] = solana.Meta(fundingAccount).WRITE().SIGNER()
	return inst
}

// Setting new account
func (inst *CreateAccount) SetNewAccount(newAccount solana.PublicKey) *CreateAccount { 
	inst.AccountMetaSlice[1] = solana.Meta(newAccount).WRITE().SIGNER()
	return inst
}
 
// NewCreateAccountInstruction declares a new CreateAccount instruction with the provided parameters and accounts.
func NewCreateAccountInstruction(
	// Parameters:
	lamports uint64,
  space uint64, 
  owner solana.PublicKey,
	// Accounts:
	fundingAccount solana.PublicKey,
	newAccount solana.PublicKey) *CreateAccount {
	return NewCreateAccountInstructionBuilder().
		SetLamports(lamports).
		SetSpace(space).
		SetOwner(owner).
		SetFundingAccount(fundingAccount).
		SetNewAccount(newAccount)
}

func (inst *CreateAccount) Validate() error {
	// Check whether all (required) parameters are set:
	if inst.Lamports == nil {
		return errors.New("Lamports parameter is not set")
	}
	if inst.Space == nil {
		return errors.New("Space parameter is not set")
	}
	if inst.Owner == nil {
		return errors.New("Owner parameter is not set")
	}

	// Check whether all accounts are set:
	for accIndex, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", accIndex)
		}
	}
	return nil
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CreateAccount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst CreateAccount) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_CreateAccount, binary.LittleEndian),
	}}
}

func (inst CreateAccount) MarshalWithEncoder(encoder *bin.Encoder) error {
	// Serialize `Lamports` param:
	{
		err := encoder.Encode(*inst.Lamports)
		if err != nil {
			return err
		}
	}
	// Serialize `Space` param:
	{
		err := encoder.Encode(*inst.Space)
		if err != nil {
			return err
		}
	}
	// Serialize `Owner` param:
	{
		err := encoder.Encode(*inst.Owner)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst *CreateAccount) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	// Deserialize `Lamports` param:
	{
		err := decoder.Decode(&inst.Lamports)
		if err != nil {
			return err
		}
	}
	// Deserialize `Space` param:
	{
		err := decoder.Decode(&inst.Space)
		if err != nil {
			return err
		}
	}
	// Deserialize `Owner` param:
	{
		err := decoder.Decode(&inst.Owner)
		if err != nil {
			return err
		}
	}
	return nil
}
