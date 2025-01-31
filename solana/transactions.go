package solana

import (
	"encoding/base64"
	"fmt"
	"sort"

	bin "github.com/gagliardetto/binary"
	"go.uber.org/zap"
	//"github.com/davecgh/go-spew/spew"
)

type Transaction struct {
	// A compact array of base-58 encoded signatures applied to the transaction.
	// The signature at index `I` corresponds to the public key at index 
	Signatures []Signature `json:"signatures"` // 64 bytes * num of signatures
	// Content of the message
	Message Message `json:"message"`
}

func (tx *Transaction) UnmarshalBase64(b64 string) error { // <- accepts base64 string
	fmt.Println("ub64")
	b, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return err
	}
	return tx.UnmarshalWithDecoder(bin.NewBinDecoder(b))
}

func (tx *Transaction) MarshalBinary() ([]byte, error) {
	messageContent, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to encode tx.Message to binary: %w", err)
	}

	var signatureCount []byte
	bin.EncodeCompactU16Length(&signatureCount, len(tx.Signatures))
	output := make([]byte, 0, len(signatureCount)+len(signatureCount)*64+len(messageContent))
	output = append(output, signatureCount...)
	for _, sig := range tx.Signatures {
		output = append(output, sig[:]...)
	}
	output = append(output, messageContent...)

	return output, nil
}

func (tx Transaction) MarshalWithEncoder(encoder *bin.Encoder) error {
	out, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	return encoder.WriteBytes(out, false)
}

// Accepts a decoder built from base 64 encoded transaction string
func TransactionFromDecoder(decoder *bin.Decoder) (*Transaction, error) {
	var output *Transaction
	err := decoder.Decode(&output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (tx *Transaction) UnmarshalWithDecoder(decoder *bin.Decoder) (err error) {
	{
		numSignatures, err := decoder.ReadCompactU16()
		if err != nil {
			return fmt.Errorf("unable to read numSignatures: %w", err)
		}
		if numSignatures < 0 {
			return fmt.Errorf("numSignatures is negative")
		}
		if numSignatures > decoder.Remaining()/64 {
			return fmt.Errorf("numSignatures %d is too large for remaining bytes %d", numSignatures, decoder.Remaining())
		}

		tx.Signatures = make([]Signature, numSignatures)
		for i := 0; i < numSignatures; i++ {
			_, err := decoder.Read(tx.Signatures[i][:])
			if err != nil {
				return fmt.Errorf("unable to read tx.Signatures[%d]: %w", i, err)
			}
		}
	}

	{
		err := tx.Message.UnmarshalWithDecoder(decoder)
		if err != nil {
			return fmt.Errorf("unable to decode tx.Message: %w", err)
		}
	}
	return nil
}

type CompiledInstruction struct {
	// Index into the message.accountKeys array indicating the (program account) that executes this instruction.
	// NOTE: it is actually uint8, but using a uint16 because uint8 is treated as a byte everywhere, and that can be an issue.
	ProgramIDIndex uint16 `json:"programIdIndex"`
	// List of ordered indices into the message.accountKeys array indicating which accounts to pass to the program.
	// NOTE: it is actually a []uint8, but using a uint16 because []uint8 is treated as a []byte everywhere, and that can be an issue.
	Accounts []uint16 `json:"accounts"`
	// The program input data encoded in base58 string.
	Data Base58 `json:"data"`
}

type Instruction interface {
	ProgramID() PublicKey     // <-- the programID the instruction acts on
	Accounts() []*AccountMeta // <-- returns the list of accounts the instructions require
	Data() ([]byte, error)    // <-- the bianry encoded instructions
}

type TransactionOption interface {
	apply(opts *transactionOptions)
}

type transactionOptions struct {
	payer         PublicKey
	addressTables map[PublicKey]PublicKeySlice
}

type transactionOptionFunc func(opts *transactionOptions)

func (f transactionOptionFunc) apply(opts *transactionOptions) {
	f(opts)
}

func TransactionPayer(payer PublicKey) TransactionOption {
	return transactionOptionFunc(func(opts *transactionOptions) { opts.payer = payer })
}

func TransactionAddressTables(tables map[PublicKey]PublicKeySlice) TransactionOption {
	return transactionOptionFunc(func(opts *transactionOptions) { opts.addressTables = tables })
}

var DebugNewTransaction = false

type addressTablePubkeyWithIndex struct {
	addressTable PublicKey
	index        uint8
}

func NewTransaction(instructions []Instruction, recentBlockHash Hash, opts ...TransactionOption) (*Transaction, error) {
	if len(instructions) < 0 {
		return nil, fmt.Errorf("requires at least one instruction to create a transaction")
	}

	// OPTS
	options := transactionOptions{}
	for _, opt := range opts { // range over the opts (TransactionPayer,)
		opt.apply(&options) // modifies options var by including opts from `transactionOptions` (payer or addressTables)
	}

	feePayer := options.payer // --> `PubicKey`
	if feePayer.IsZero() {
		found := false
		for _, acc := range instructions[0].Accounts() {
			if acc.IsSigner {
				if DebugNewTransaction {
					zlog.Info("Found a fee payer", zap.Stringer("account_pub_key", acc.PublicKey))
				}
				feePayer = acc.PublicKey
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("cannot determine a fee payer. You can either pass the fee payer via the 'TransactionWithInstructions' option parameter or it falls back to the first instruction's first signer")
		}
	}

	addressLookupKeysMap := make(map[PublicKey]addressTablePubkeyWithIndex) // all accounts from tables as map
	for addressTablePubKey, addressTable := range options.addressTables {   // map[addressTablePubKey] = PublicKeySlice
		if len(addressTable) > 256 {
			return nil, fmt.Errorf("max lookup table index exceeded for %s table", addressTablePubKey)
		}

		for i, address := range addressTable {
			_, ok := addressLookupKeysMap[address]
			if ok {
				continue
			}
			addressLookupKeysMap[address] = addressTablePubkeyWithIndex{addressTable: address, index: uint8(i)}
		}
	}

	// INSTRUCTIONS
	programIDs := make(PublicKeySlice, 0)
	accounts := []*AccountMeta{}

	for _, instruction := range instructions {
		accounts = append(accounts, instruction.Accounts()...)
		programIDs.UniqueAppend(instruction.ProgramID())
	}

	if DebugNewTransaction {
		zlog.Info("instruction accounts:", zap.Int("num_accounts", len(accounts)))
		zlog.Info("instruction programs:", zap.Int("num_programdIDs", len(programIDs)))
	}
	programIDsMap := make(map[PublicKey]struct{}, len(programIDs)) // for IsInvoke check

	// Add programID to account list
	for _, programID := range programIDs {
		accounts = append(accounts, &AccountMeta{
			PublicKey:  programID,
			IsSigner:   false,
			IsWritable: false,
		})

		programIDsMap[programID] = struct{}{}
	}

	// Sort. Prioritizing first by Signer then by writable
	sort.SliceStable(accounts, func(i, j int) bool {
		return accounts[i].less(accounts[j])
	})

	uniqAccountsMap := map[PublicKey]uint64{} // map[PubKey]index
	uniqAccounts := []*AccountMeta{}
	for _, acc := range accounts {
		if index, found := uniqAccountsMap[acc.PublicKey]; found {
			uniqAccounts[index].IsWritable = uniqAccounts[index].IsWritable || acc.IsWritable // if any of similar accs are writables make them writable
			continue
		}

		uniqAccounts = append(uniqAccounts, acc)
		uniqAccountsMap[acc.PublicKey] = uint64(len(uniqAccounts) - 1)
	}

	if DebugNewTransaction {
		zlog.Debug("unqiue accounts sorted", zap.Int("account_count", len(uniqAccounts)))
	}

	// Move the payer to the front
	feePayerIndex := -1
	for idx, acc := range uniqAccounts {
		if acc.PublicKey.Equals(feePayer) {
			feePayerIndex = idx
		}
	}

	if DebugNewTransaction {
		zlog.Debug("current fee payer index", zap.Int("fee_payer_index", feePayerIndex))
	}

	accountCount := len(uniqAccounts)
	if feePayerIndex < 0 {
		// fee payer isnt part of accounts, we want to add it ???
		accountCount++
	}

	allKeys := make([]*AccountMeta, accountCount)

	itr := 1 // reserve 0 for the FeePayer
	for idx, uniqAccount := range uniqAccounts {
		if idx == feePayerIndex {
			uniqAccount.IsSigner = true
			uniqAccount.IsWritable = true
			allKeys[0] = uniqAccount
			continue
		}
		allKeys[itr] = uniqAccount
		itr++
	}

	if feePayerIndex < 0 {
		feePayerAccount := &AccountMeta{PublicKey: feePayer, IsSigner: true, IsWritable: true}
		allKeys[0] = feePayerAccount
	}

	message := Message{RecentBlockhash: recentBlockHash}

	lookupsMap := make(map[PublicKey]struct { // extended MessageAddressTableLookup
		AccountKey      PublicKey // the account key of address table (what accesses the table)
		WritableIndexes []uint8
		Writable        []PublicKey
		ReadonlyIndexes []uint8
		Readonly        []PublicKey
	})

	for idx, acc := range allKeys {
		if DebugNewTransaction {
			zlog.Debug("transaction account", zap.Int("account_index", idx), zap.Stringer("account_pub_key", acc.PublicKey))
		}

		addressLookupKeyEntry, isPresentInTables := addressLookupKeysMap[acc.PublicKey] // comma-ok syntax* | Returns (addressTablePubkey + index)
		_, IsInvoke := programIDsMap[acc.PublicKey]                                     // comma-ok synax* | IsInvoke is a bool in this case

		if isPresentInTables && idx != 0 && !acc.IsSigner && !IsInvoke { // Present - not feePayer - not signer - no programID
			lookup := lookupsMap[addressLookupKeyEntry.addressTable]
			if acc.IsWritable {
				lookup.WritableIndexes = append(lookup.WritableIndexes, addressLookupKeyEntry.index)
				lookup.Writable = append(lookup.Writable, acc.PublicKey)
			} else {
				lookup.ReadonlyIndexes = append(lookup.ReadonlyIndexes, addressLookupKeyEntry.index)
				lookup.Writable = append(lookup.Readonly, acc.PublicKey)
			}
			lookupsMap[addressLookupKeyEntry.addressTable] = lookup
			continue
		}

		message.AccountKeys = append(message.AccountKeys, acc.PublicKey)
		if acc.IsSigner {
			message.Header.NumRequiredSignatures++ // Number of valid signatures required fo the transaction to be valid
			if !acc.IsWritable {
				message.Header.NumReadonlySignedAccounts++
			}
			continue
		}
		if !acc.IsWritable {
			message.Header.NumReadonlyUnsignedAccounts++
		}
	}

	var lookupsWritableKeys []PublicKey
	var lookupsReadonlyKeys []PublicKey

	if len(lookupsMap) > 0 {
		lookups := make([]MessageAddressTableLookup, 0, len(lookupsMap))
		for tablePublicKey, lkp := range lookupsMap {
			lookupsWritableKeys = append(lookupsWritableKeys, lkp.Writable...)
			lookupsReadonlyKeys = append(lookupsReadonlyKeys, lkp.Readonly...)

			lookups = append(lookups, MessageAddressTableLookup{
				AccountKey:      tablePublicKey,
				WritableIndexes: lkp.WritableIndexes,
				ReadonlyIndexes: lkp.ReadonlyIndexes,
			})
			if DebugNewTransaction {
				zlog.Debug("filled lookupsWritableKeys", zap.Int("writable_keys", len(lookupsWritableKeys)))
				zlog.Debug("filled lookupsWritableKeys", zap.Int("readonly_keys", len(lookupsReadonlyKeys)))
			}
		}

		err := message.SetAddressTables(options.addressTables)
		if err != nil {
			return nil, fmt.Errorf("message.SetAddressTables: %w", err)
		}
		message.SetAddressTableLookups(lookups)
		if DebugNewTransaction {
			zlog.Debug("set msg address tables from lookups", zap.Int("msg_alt_length", len(message.AddressTableLookups)))
		}
	}

	var idx uint16
	accountKeyIndex := make(map[string]uint16, len(message.AccountKeys)+len(lookupsWritableKeys)+len(lookupsReadonlyKeys))
	for _, acc := range message.AccountKeys {
		accountKeyIndex[acc.String()] = idx
		idx++
	}
	for _, acc := range lookupsWritableKeys {
		accountKeyIndex[acc.String()] = idx
		idx++
	}
	for _, acc := range lookupsReadonlyKeys {
		accountKeyIndex[acc.String()] = idx
		idx++
	}

	if DebugNewTransaction {
		zlog.Debug("message header compiled",
			zap.Uint8("num_required_signatures", message.Header.NumRequiredSignatures),
			zap.Uint8("num_readonly_signed_accounts", message.Header.NumReadonlySignedAccounts),
			zap.Uint8("num_readonly_unsigned_accounts", message.Header.NumReadonlyUnsignedAccounts),
		)
	}

	for txIdx, instruction := range instructions {
		accounts := instruction.Accounts()
		accountIndex := make([]uint16, len(accounts))
		if DebugNewTransaction {
			zlog.Debug("processing instruction:", zap.Int("transaction_id", txIdx))
		}
		for idx, acc := range accounts {
			accountIndex[idx] = accountKeyIndex[acc.PublicKey.String()]
			if DebugNewTransaction {
				zlog.Debug(fmt.Sprintf("set accountIndex[%d] with key", idx), zap.String("account_pub_key", acc.PublicKey.String()))
			}
		}
		data, err := instruction.Data()
		if err != nil {
			return nil, fmt.Errorf("Unable to encode instructions[%d]: %w", txIdx, err)
		}
		message.Instructions = append(message.Instructions,
			CompiledInstruction{
				ProgramIDIndex: accountKeyIndex[instruction.ProgramID().String()],
				Accounts:       accountIndex,
				Data:           data,
			})
	}

	return &Transaction{Message: message}, nil
}

type privateKeyGetter func(key PublicKey) *PrivateKey

func (tx *Transaction) PartialSign(getter privateKeyGetter) (out []Signature, err error) {
	DebugNewTransaction = true
	messageContent, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("unable to encode message for signing: %w", err)
	}
	signerKeys := tx.Message.signerKeys()

	if len(tx.Signatures) == 0 {
		tx.Signatures = make([]Signature, len(signerKeys))
	} else if len(tx.Signatures) != len(signerKeys) {
		return nil, fmt.Errorf("invalid signatures length, expected %d, actual %d", len(signerKeys), len(tx.Signatures))
	}

	for i, key := range signerKeys {
		privateKey := getter(key)
		if privateKey != nil {
			sig, err := privateKey.Sign(messageContent)
			if err != nil {
				return nil, fmt.Errorf("failed to sign with key %q: %w", key.String(), err)
			}
			// Directly assing a signature to the corresponding position the transaction's signaute slice
			tx.Signatures[i] = sig
		}
	}

	return tx.Signatures, nil
}

func (tx *Transaction) Sign(getter privateKeyGetter) (out []Signature, err error) {
	signerKeys := tx.Message.signerKeys() // returns all signed acc keys
	for _, key := range signerKeys {
		if getter(key) == nil { // check if signer's `pub key` matches the `key` from signerKeys
			return nil, fmt.Errorf("singer key %q not found. Ensure all the singer keys are in the vault", key.String())
		}
	}
	return tx.PartialSign(getter)
}

func (tx Transaction) ToBase64() (string, error) {
	txs_bytes, err := tx.MarshalBinary()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(txs_bytes), nil
}

func (tx Transaction) MustToBase64() string {
	out, err := tx.ToBase64()
	if err != nil {
		panic(err)
	}
	return out
}
