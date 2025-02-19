package token

import (
	"errors"

	bin "github.com/gagliardetto/binary"
	solana "github.com/scatkit/pumpdexer/solana"
)

// Given a wrapped / native token account (a token account containing SOL)
// updates its amount field based on the account's underlying `lamports`.
// This is useful if a non-wrapped SOL account uses `system_instruction::transfer`
// to move lamports to a wrapped token account, and needs to have its token
// `amount` field updated.
type SyncNative struct {

	// [0] = [WRITE] tokenAccount
	// ··········· The native token account to sync with its underlying lamports.
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewSyncNativeInstructionBuilder creates a new `SyncNative` instruction builder.
func NewSyncNativeInstructionBuilder() *SyncNative {
	nd := &SyncNative{
		AccountMetaSlice: make(solana.AccountMetaSlice, 1),
	}
	return nd
}

// SetTokenAccount sets the "tokenAccount" account.
// The native token account to sync with its underlying lamports.
func (inst *SyncNative) SetTokenAccount(tokenAccount solana.PublicKey) *SyncNative {
	inst.AccountMetaSlice[0] = solana.Meta(tokenAccount).WRITE()
	return inst
}

// GetTokenAccount gets the "tokenAccount" account.
// The native token account to sync with its underlying lamports.
func (inst *SyncNative) GetTokenAccount() *solana.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst SyncNative) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint8(Instruction_SyncNative),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst SyncNative) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *SyncNative) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.TokenAccount is not set")
		}
	}
	return nil
}

func (obj SyncNative) MarshalWithEncoder(encoder *bin.Encoder) (err error) {
	return nil
}
func (obj *SyncNative) UnmarshalWithDecoder(decoder *bin.Decoder) (err error) {
	return nil
}

// NewSyncNativeInstruction declares a new SyncNative instruction with the provided parameters and accounts.
func NewSyncNativeInstruction(
	// Accounts:
	tokenAccount solana.PublicKey) *SyncNative {
	return NewSyncNativeInstructionBuilder().
		SetTokenAccount(tokenAccount)
}
