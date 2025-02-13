package token
import(
  "fmt"
  "errors"
  "github.com/scatkit/pumpdexer/solana"
  bin "github.com/gagliardetto/binary"
)

// Close an account by transferring all of its SOL to the destination account
// Non-native accounts may only be closed if its token amount is zero.
type CloseAccount struct{
  // [0] = [WRITE] Account: `The account to close`
	// [1] = [WRITe] Destination: `The destination account`
	// [2] = [] Owner: `The account's owner`
	// [3...] = [SIGNER] signers: `signer accounts``
  Accounts solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
  Signers solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewCloseAccountBuilder() *CloseAccount{
  return &CloseAccount{
    Accounts: make(solana.AccountMetaSlice, 3),
    Signers: make(solana.AccountMetaSlice, 0),
  }
}

// SetAccount sets the "account" account.
// The account to close.
func (inst *CloseAccount) SetAccount(account solana.PublicKey) *CloseAccount{
  inst.Accounts[0] = solana.Meta(account).WRITE()
  return inst
}

// GetAccount gets the "account" account.
// The account to close.
func (inst *CloseAccount) GetAccount() *solana.AccountMeta {
	return inst.Accounts[0]
}

// SetDestinationAccount sets the "destination" account.
// The destination account.
func (inst *CloseAccount) SetDestinationAccount(destination solana.PublicKey) *CloseAccount {
	inst.Accounts[1] = solana.Meta(destination).WRITE()
	return inst
}

// GetDestinationAccount gets the "destination" account.
// The destination account.
func (inst *CloseAccount) GetDestinationAccount() *solana.AccountMeta {
	return inst.Accounts[1]
}

// SetOwnerAccount sets the "owner" account.
// The account's owner.
func (inst *CloseAccount) SetOwnerAccount(owner solana.PublicKey, multisigSigners ...solana.PublicKey) *CloseAccount {
	inst.Accounts[2] = solana.Meta(owner)
	if len(multisigSigners) == 0 {
		inst.Accounts[2].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, solana.Meta(signer).SIGNER())
	}
	return inst
}

// GetOwnerAccount gets the "owner" account.
// The account's owner.
func (inst *CloseAccount) GetOwnerAccount() *solana.AccountMeta {
	return inst.Accounts[2]
}

func (inst CloseAccount) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint8(Instruction_CloseAccount),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CloseAccount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CloseAccount) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.Accounts[0] == nil {
			return errors.New("accounts.Account is not set")
		}
		if inst.Accounts[1] == nil {
			return errors.New("accounts.Destination is not set")
		}
		if inst.Accounts[2] == nil {
			return errors.New("accounts.Owner is not set")
		}
		if !inst.Accounts[2].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
		if len(inst.Signers) > MAX_SIGNERS {
			return fmt.Errorf("too many signers; got %v, but max is 11", len(inst.Signers))
		}
	}
	return nil
}

func (obj CloseAccount) MarshalWithEncoder(encoder *bin.Encoder) (err error) {
	return nil
}
func (obj *CloseAccount) UnmarshalWithDecoder(decoder *bin.Decoder) (err error) {
	return nil
}

// Declares a new CloseAccount instruction with the provided parameters and accounts.
func NewCloseAccountInstruction(
  // Accounts:
  account solana.PublicKey,
  destination solana.PublicKey,
  owner solana.PublicKey,
  multisigSigners []solana.PublicKey,
) *CloseAccount{
  return NewCloseAccountBuilder().
    SetAccount(account).
    SetDestinationAccount(destination).
    SetOwnerAccount(owner, multisigSigners...)
}

