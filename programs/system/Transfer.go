package system

import (
	"encoding/binary"

	gg_binary "github.com/gagliardetto/binary"
	scatsol "github.com/scatkit/pumpdexer/solana"
)

// AccountMeta{
// 	PublicKey:  pubKey,
// 	IsWritable: WRITE,
// 	IsSigner:   SIGNER,
// }

// Transfer lamports
type Transfer struct {
	// Number of lamports to transfer to the new account
	Lamports *uint64
	// [0] = [WRITE, SIGNER] FundingAccount
	// ··········· Funding account
	// [1] = [WRITE] RecipientAccount
	// ··········· Recipient account
	scatsol.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst Transfer) Build() *Instruction {
	return &Instruction{BaseVariant: gg_binary.BaseVariant{
		Impl:   inst,
		TypeID: gg_binary.TypeIDFromUint32(Instruction_Transfer, binary.LittleEndian),
	}}
}

// NewTransferInstruction declares a new Transfer instruction with the provided parameters and accounts.
func NewTransferInstruction(
	// Parameters:
	lamports uint64,
	// Accounts:
	fundingAccount scatsol.PublicKey,
	recipientAccount scatsol.PublicKey) *Transfer {
	return NewTransferInstructionBuilder().
		SetLamports(lamports). // retunrs Transfer
		SetFundingAccount(fundingAccount).
		SetRecipientAccount(recipientAccount)
}

func (inst *Transfer) SetLamports(lamps uint64) *Transfer {
	inst.Lamports = &lamps
	return inst
}

// Sender
func (inst *Transfer) SetFundingAccount(senderAccount scatsol.PublicKey) *Transfer {
	inst.AccountMetaSlice[0] = scatsol.Meta(senderAccount).WRITE().SIGNER()
	return inst
}

// Recepient
func (inst *Transfer) SetRecipientAccount(recepinetAccount scatsol.PublicKey) *Transfer {
	inst.AccountMetaSlice[1] = scatsol.Meta(recepinetAccount).WRITE()
	return inst
}

// NewTransferInstructionBuilder creates a new `Transfer` instruction builder.
func NewTransferInstructionBuilder() *Transfer {
	nd := &Transfer{
		AccountMetaSlice: make(scatsol.AccountMetaSlice, 2),
	}
	return nd
}
