package token
import (
	"errors"

  bin "github.com/gagliardetto/binary"
  "github.com/scatkit/pumpdexer/solana"
)

type InitializeAccount struct{
  // [0] = [WRITE] Account:         `The account to initalize`
	// [1] = [] Mint:                 `The mint this account will be associated with`
	// [2] = [] Owner:                `The new account's owner/multisignature`
	// [3] = [] $(SysVarRentPubkey):  `Rent sysvar`
  solana.AccountMetaSlice `bin:"-" borsh_skip:"true"` // has to be ananomys for bin to access it
}

// NewInitializeAccountInstructionBuilder creates a new `InitializeAccount` instruction builder.
func NewInitializeAccountInstructionBuilder() *InitializeAccount {
	nd := &InitializeAccount{
		AccountMetaSlice: make(solana.AccountMetaSlice, 4),
	}
	nd.AccountMetaSlice[3] = solana.Meta(solana.SysVarRentPubkey)
	return nd
}

// SetAccount sets the "account" account.
// The account to initialize.
func (inst *InitializeAccount) SetAccount(account solana.PublicKey) *InitializeAccount {
	inst.AccountMetaSlice[0] = solana.Meta(account).WRITE()
	return inst
}
 
// SetMintAccount sets the "mint" account.
// The mint this account will be associated with.
func (inst *InitializeAccount) SetMintAccount(mint solana.PublicKey) *InitializeAccount {
	inst.AccountMetaSlice[1] = solana.Meta(mint)
	return inst
}

// GetMintAccount gets the "mint" account.
// The mint this account will be associated with.
func (inst *InitializeAccount) GetMintAccount() *solana.AccountMeta{
	return inst.AccountMetaSlice[1]
}

// SetOwnerAccount sets the "owner" account.
// The new account's owner/multisignature.
func (inst *InitializeAccount) SetOwnerAccount(owner solana.PublicKey) *InitializeAccount {
	inst.AccountMetaSlice[2] = solana.Meta(owner)
	return inst
}

// GetOwnerAccount gets the "owner" account.
// The new account's owner/multisignature.
func (inst *InitializeAccount) GetOwnerAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[2]
}

// SetSysVarRentPubkeyAccount sets the "$(SysVarRentPubkey)" account.
// Rent sysvar.
func (inst *InitializeAccount) SetSysVarRentPubkeyAccount(SysVarRentPubkey solana.PublicKey) *InitializeAccount {
	inst.AccountMetaSlice[3] = solana.Meta(SysVarRentPubkey)
	return inst
}

// GetSysVarRentPubkeyAccount gets the "$(SysVarRentPubkey)" account.
// Rent sysvar.
func (inst *InitializeAccount) GetSysVarRentPubkeyAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[3]
}

func (inst InitializeAccount) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint8(Instruction_InitializeAccount),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst InitializeAccount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeAccount) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Account is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Mint is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.Owner is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.SysVarRentPubkey is not set")
		}
	}
	return nil
}

// NewInitializeAccountInstruction declares a new InitializeAccount instruction with the provided parameters and accounts
func NewInitializeAccountInstruction(
	// Accounts:
	account solana.PublicKey,
	mint solana.PublicKey,
	owner solana.PublicKey,
	SysVarRentPubkey solana.PublicKey) *InitializeAccount {
	return NewInitializeAccountInstructionBuilder().
		SetAccount(account).
		SetMintAccount(mint).
		SetOwnerAccount(owner).
		SetSysVarRentPubkeyAccount(SysVarRentPubkey)
}

