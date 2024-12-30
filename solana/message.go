package solana

import (
	"encoding/base64"
	"fmt"

	bin "github.com/gagliardetto/binary"
)

type MessageVersion int

const (
	MessageVersionLegacy MessageVersion = 0
	MessageVersionV0     MessageVersion = 1
)

type MessageAddressTableLookupSlice []MessageAddressTableLookup

// Number of accounts from all the lookups
func (lookups MessageAddressTableLookupSlice) NumLookups() int {
	count := 0
	for _, lookup := range lookups {
		count += len(lookup.WritableIndexes)
		count += len(lookup.ReadonlyIndexes)
	}
	return count
}

func (lookups MessageAddressTableLookupSlice) NumWritableLookups() int {
	count := 0
	for _, lookup := range lookups {
		count += len(lookup.WritableIndexes)
	}
	return count
}

func (lookups MessageAddressTableLookupSlice) GetTableIDs() PublicKeySlice {
	if lookups == nil {
		return nil
	}
	ids := make(PublicKeySlice, 0)
	for _, lookup := range lookups {
		ids.UniqueAppend(lookup.AccountKey)
	}
	return ids
}

type MessageAddressTableLookup struct {
	AccountKey      PublicKey       `json:"accountKey"` // the account key of the address table
	WritableIndexes Uint8SliceAsNum `json:"writableIndexes"`
	ReadonlyIndexes Uint8SliceAsNum `json:"readonlyIndexes"`
}

type Uint8SliceAsNum []uint8

func (slice Uint8SliceAsNum) MarshalJSON() ([]byte, error) {
	out := make([]uint16, len(slice))
	for i, idx := range slice {
		out[i] = uint16(idx)
	}
	return json.Marshal(out)
}

type MessageHeader struct {
	// Specifies how many valid signatures are required for the transaction to be valid
	// The signatures must match the first `numRequiredSignatures` of `message.accountKeys`.
	NumRequiredSignatures uint8 `json:"numRequiredSignatures"`

	// The last numReadonlySignedAccounts of the signed keys are read-only accounts.
	// Programs may process multiple transactions that load read-only accounts within
	// a single PoH entry (in parallel), but are not permitted to credit or debit lamports or modify
	// account data.
	// Transactions targeting the same writable account are evaluated sequentially to avoid conflicts.
	NumReadonlySignedAccounts uint8 `json:"numReadonlySignedAccounts"`

	//(e.g TokenMint acc) the last `numReadonlyUnsignedAccounts` of the unsigned keys are read-only accounts.
	NumReadonlyUnsignedAccounts uint8 `json:"numReadOnlyUnsignedAccounts"`
}
type Message struct {
	version MessageVersion
	// List of base-58 encoded public keys used by the transaction,
	// including by the instructions and for signatures.
	// The first `message.header.numRequiredSignatures` public keys must sign the transaction.
	AccountKeys PublicKeySlice `json:"accountKeys"`
	// Account types and signatures required by the transaction
	Header MessageHeader `json:"header"`
	// A base-58 encoded hash of a recent block in the ledger used to
	// prevent transaction duplication and to give transactions lifetimes.
	RecentBlockhash Hash `json:"recentBlockhash"`
	// List of program instructions that will be executed in a sequence
	// and committed in one atomic transaction if all succeed.
	Instructions []CompiledInstruction `json:"instructions"`

	// List of address table lookups to laod additional accounts for this transaction
	AddressTableLookups MessageAddressTableLookupSlice `json:"addressTableLookups"`

	// The actual tables that contain the list of account pubkeys.
	// NOTE: you need to fetch these from the chain, and then call `SetAddressTables`
	// before you use this transaction -- otherwise, you will get a panic.
	addressTables map[PublicKey]PublicKeySlice

	resolved bool //if true, the lookups have been resolved, and the `AccountKeys` slice contains all the accounts (static + dynamic)
}

func (msg Message) signerKeys() PublicKeySlice{
  return msg.AccountKeys[0:msg.Header.NumRequiredSignatures]
}

func (msg *Message) MarshalBinary() ([]byte, error) {
	switch msg.version {
	case MessageVersionV0:
		return msg.MarshalV0()
	case MessageVersionLegacy:
		return msg.MarshalLegacy()
	default:
		return nil, fmt.Errorf("Invalid message version: %d", msg.version)
	}
}

func (msg *Message) MarshalLegacy() ([]byte, error) {
	buf := []byte{
		msg.Header.NumRequiredSignatures,
		msg.Header.NumReadonlySignedAccounts,
		msg.Header.NumReadonlyUnsignedAccounts,
	}

	bin.EncodeCompactU16Length(&buf, len(msg.AccountKeys))
	for _, key := range msg.AccountKeys {
		buf = append(buf, key[:]...)
	}

	buf = append(buf, msg.RecentBlockhash[:]...)

	bin.EncodeCompactU16Length(&buf, len(msg.Instructions))
	for _, instruction := range msg.Instructions {
		buf = append(buf, byte(instruction.ProgramIDIndex))
		bin.EncodeCompactU16Length(&buf, len(instruction.Accounts))
		for _, accountIdx := range instruction.Accounts {
			buf = append(buf, byte(accountIdx))
		}
		bin.EncodeCompactU16Length(&buf, len(instruction.Data))
		buf = append(buf, instruction.Data...)
	}
	return buf, nil
}

func (msg Message) getStaticKeys() (keys PublicKeySlice) {
	if msg.resolved {
		// If the message has been resolved, then the account keys have already
		// been appended to the `AccountKeys` field of the message.
		return msg.AccountKeys[:msg.numStaticAccounts()] // excluding lookups
	}
	return msg.AccountKeys
}

// Number of accounts from all the lookups
func (m Message) NumLookups() int {
	if m.AddressTableLookups == nil {
		return 0
	}
	return m.AddressTableLookups.NumLookups()
}

func (mx Message) NumWritableLookups() int {
	if mx.AddressTableLookups == nil {
		return 0
	}
	return mx.AddressTableLookups.NumWritableLookups()
}

// numStaticAccounts returns the number of accounts that are always present in the
// account keys list (i.e. all the accounts that are NOT in the lookup table).
func (m Message) numStaticAccounts() int {
	if !m.resolved {
		return len(m.AccountKeys)
	}
	return len(m.AccountKeys) - m.NumLookups()
}

func (msg *Message) MarshalV0() ([]byte, error) {
	buf := []byte{
		msg.Header.NumRequiredSignatures,
		msg.Header.NumReadonlySignedAccounts,
		msg.Header.NumReadonlyUnsignedAccounts,
	}
	{
		// Encode only the keys that are not in the address table lookups.
		staticAccountKeys := msg.getStaticKeys()
		bin.EncodeCompactU16Length(&buf, len(staticAccountKeys))
		for _, key := range staticAccountKeys {
			buf = append(buf, key[:]...)
		}

		buf = append(buf, msg.RecentBlockhash[:]...)

		bin.EncodeCompactU16Length(&buf, len(msg.Instructions))
		for _, instruction := range msg.Instructions {
			buf = append(buf, byte(instruction.ProgramIDIndex))
			bin.EncodeCompactU16Length(&buf, len(instruction.Accounts))
			for _, accountIdx := range instruction.Accounts {
				buf = append(buf, byte(accountIdx))
			}

			bin.EncodeCompactU16Length(&buf, len(instruction.Data))
			buf = append(buf, instruction.Data...)
		}
	}
	versionNum := byte(msg.version) // TODO: what number is this?
	if versionNum > 127 {
		return nil, fmt.Errorf("invalid message version: %d", msg.version)
	}
	buf = append([]byte{byte(versionNum + 127)}, buf...)

	if msg.AddressTableLookups != nil && len(msg.AddressTableLookups) > 0 {
		// wite length of address table lookups as u8
		buf = append(buf, byte(len(msg.AddressTableLookups)))
		for _, lookup := range msg.AddressTableLookups {
			// write account pubkey
			buf = append(buf, lookup.AccountKey[:]...)
			// write writable indexes
			bin.EncodeCompactU16Length(&buf, len(lookup.WritableIndexes))
			buf = append(buf, lookup.WritableIndexes...)
			// write readonly indexes
			bin.EncodeCompactU16Length(&buf, len(lookup.ReadonlyIndexes))
			buf = append(buf, lookup.ReadonlyIndexes...)
		}
	} else {
		buf = append(buf, 0)
	}
	return buf, nil
}

func (msg *Message) UnmarshalWithDecoder(decoder *bin.Decoder) (err error) {
	// peek the first byte
	versionNum, err := decoder.Peek(1)
	if err != nil {
		return err
	}
	if versionNum[0] < 127 {
		msg.version = MessageVersionLegacy
	} else {
		msg.version = MessageVersionV0
	}
	switch msg.version {
	case MessageVersionV0:
		return msg.UnmarshalV0(decoder)
	case MessageVersionLegacy:
		return msg.UnmarshalLegacy(decoder)
	default:
		return fmt.Errorf("invalid message version: %d", msg.version)
	}
}

func (msg *Message) UnmarshalV0(decoder *bin.Decoder) (err error) {
	version, err := decoder.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read message version: %w", err)
	}
	// TODO: check version. `MessageVersion` is int
	msg.version = MessageVersion(version - 127)

	// The middle of the message is the same as the legacy message:
	err = msg.UnmarshalLegacy(decoder)
	if err != nil {
		return err
	}

	// Address Lookup Tables (new)
	// Read address table lookups length:
	addressTableLookupsLen, err := decoder.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read address table lookups length: %w", err)
	}
	if addressTableLookupsLen > 0 {
		msg.AddressTableLookups = make([]MessageAddressTableLookup, addressTableLookupsLen)
		for i := 0; i < int(addressTableLookupsLen); i++ {
			// read account pubkey
			_, err = decoder.Read(msg.AddressTableLookups[i].AccountKey[:])
			if err != nil {
				return fmt.Errorf("failed to read account pubkey: %w", err)
			}

			// read writable indexes
			writableIndexesLen, err := decoder.ReadCompactU16()
			if err != nil {
				return fmt.Errorf("failed to read writable indexes length: %w", err)
			}
			if writableIndexesLen > decoder.Remaining() {
				return fmt.Errorf("writable indexes length is too large: %d", writableIndexesLen)
			}
			msg.AddressTableLookups[i].WritableIndexes = make([]byte, writableIndexesLen)
			_, err = decoder.Read(msg.AddressTableLookups[i].WritableIndexes)
			if err != nil {
				return fmt.Errorf("failed to read writable indexes: %w", err)
			}

			// read readonly indexes
			readonlyIndexesLen, err := decoder.ReadCompactU16()
			if err != nil {
				return fmt.Errorf("failed to read readonly indexes length: %w", err)
			}
			if readonlyIndexesLen > decoder.Remaining() {
				return fmt.Errorf("readonly indexes length is too large: %d", readonlyIndexesLen)
			}
			msg.AddressTableLookups[i].ReadonlyIndexes = make([]byte, readonlyIndexesLen)
			_, err = decoder.Read(msg.AddressTableLookups[i].ReadonlyIndexes)
			if err != nil {
				return fmt.Errorf("failed to read readonly indexes: %w", err)
			}
		}
	}
	return nil
}

func (msg *Message) UnmarshalLegacy(decoder *bin.Decoder) (err error) {
	{
		msg.Header.NumRequiredSignatures, err = decoder.ReadUint8()
		if err != nil {
			return fmt.Errorf("unable to decode msg.Header.NumRequiredSignatures: %w", err)
		}
		msg.Header.NumReadonlySignedAccounts, err = decoder.ReadUint8()
		if err != nil {
			return fmt.Errorf("unable to decode msg.Header.NumReadonlySignedAccounts: %w", err)
		}
		msg.Header.NumReadonlyUnsignedAccounts, err = decoder.ReadUint8()
		if err != nil {
			return fmt.Errorf("unable to decode msg.Header.NumReadonlyUnsignedAccounts: %w", err)
		}
	}

	// Read pubkeys
	{
		numAccountKeys, err := decoder.ReadCompactU16()
		if err != nil {
			return fmt.Errorf("unable to decode numAccountKeys: %w", err)
		}
		if numAccountKeys > decoder.Remaining()/32 {
			return fmt.Errorf("numAccountKeys %d is too large for remaining bytes %d", numAccountKeys, decoder.Remaining())
		}
		msg.AccountKeys = make(PublicKeySlice, numAccountKeys)
		for i := 0; i < numAccountKeys; i++ {
			_, err := decoder.Read(msg.AccountKeys[i][:])
			if err != nil {
				return fmt.Errorf("unable to decode msg.AccountKeys[%d]: %w", i, err)
			}
		}
	}

	{
		_, err := decoder.Read(msg.RecentBlockhash[:])
		if err != nil {
			return fmt.Errorf("unable to decode msg.RecentBlockhash: %w", err)
		}
	}

	// Read the instructions
	{
		numInstructions, err := decoder.ReadCompactU16()
		if err != nil {
			return fmt.Errorf("unable to decode numInstructions: %w", err)
		}
		if numInstructions > decoder.Remaining() {
			return fmt.Errorf("numInstructions %d is greater than remaining bytes %d", numInstructions, decoder.Remaining())
		}
		msg.Instructions = make([]CompiledInstruction, numInstructions)
		// Per instruction:
		for instructionIndex := 0; instructionIndex < numInstructions; instructionIndex++ {
			// Read programIdIndex
			programIDIndex, err := decoder.ReadUint8() //fetching the data and shiting the index of the decoder
			if err != nil {
				return fmt.Errorf("unable to decode msg.Instructions[%d].ProgramIDIndex: %w", instructionIndex, err)
			}
			msg.Instructions[instructionIndex].ProgramIDIndex = uint16(programIDIndex)

			// Read number of accounts
			{
				numAccounts, err := decoder.ReadCompactU16()
				if err != nil {
					return fmt.Errorf("unable to decode numAccounts for ix[%d]: %w", instructionIndex, err)
				}
				if numAccounts > decoder.Remaining() {
					return fmt.Errorf("ix[%v]: numAccounts %d is greater than remaining bytes %d", instructionIndex, numAccounts, decoder.Remaining())
				}
				msg.Instructions[instructionIndex].Accounts = make([]uint16, numAccounts)
				// Per account:
				for i := 0; i < numAccounts; i++ {
					// Read the
					accountIndex, err := decoder.ReadUint8()
					if err != nil {
						return fmt.Errorf("unable to decode accountIndex for ix[%d].Accounts[%d]: %w", instructionIndex, i, err)
					}
					msg.Instructions[instructionIndex].Accounts[i] = uint16(accountIndex)
				}
			}
			// Read dataLen
			{
				dataLen, err := decoder.ReadCompactU16()
				if err != nil {
					return fmt.Errorf("unable to decode dataLen for ix[%d]: %w", instructionIndex, err)
				}
				if dataLen > decoder.Remaining() {
					return fmt.Errorf("ix[%v]: dataLen %d is greater than remaining bytes %d", instructionIndex, dataLen, decoder.Remaining())
				}
				dataBytes, err := decoder.ReadBytes(dataLen) // <-- read bytes on n length
				if err != nil {
					return fmt.Errorf("unable to decode dataBytes for ix[%d]: %w", instructionIndex, err)
				}
				msg.Instructions[instructionIndex].Data = (Base58)(dataBytes)
			}
		}
	}

	return nil
}

// Added:
func (m *Message) SetVersion(version MessageVersion) *Message {
	switch version { // <-- check if the version is valid
	case MessageVersionV0, MessageVersionLegacy:
	default:
		panic(fmt.Errorf("invalid message version: %d", version))
	}
	m.version = version
	return m
}

// GetVersion returns the message version.
func (m *Message) GetVersion() MessageVersion {
	return m.version
}

func (mx Message) MarshalJSON() ([]byte, error) {
	if mx.version == MessageVersionLegacy {
		out := struct {
			AccountKeys     []string              `json:"accountKeys"`
			Header          MessageHeader         `json:"header"`
			RecentBlockhash string                `json:"recentBlockhash"`
			Instructions    []CompiledInstruction `json:"instructions"`
		}{
			AccountKeys:     make([]string, len(mx.AccountKeys)),
			Header:          mx.Header,
			RecentBlockhash: mx.RecentBlockhash.String(),
			Instructions:    mx.Instructions,
		}
		for i, key := range mx.AccountKeys {
			out.AccountKeys[i] = key.String()
		}
		return json.Marshal(out)
	}
	// Versioned message:
	out := struct {
		AccountKeys         []string                    `json:"accountKeys"`
		Header              MessageHeader               `json:"header"`
		RecentBlockhash     string                      `json:"recentBlockhash"`
		Instructions        []CompiledInstruction       `json:"instructions"`
		AddressTableLookups []MessageAddressTableLookup `json:"addressTableLookups"`
	}{
		AccountKeys:         make([]string, len(mx.AccountKeys)),
		Header:              mx.Header,
		RecentBlockhash:     mx.RecentBlockhash.String(),
		Instructions:        mx.Instructions,
		AddressTableLookups: mx.AddressTableLookups,
	}
	for i, key := range mx.AccountKeys {
		out.AccountKeys[i] = key.String()
	}
	if out.AddressTableLookups == nil {
		out.AddressTableLookups = make([]MessageAddressTableLookup, 0)
	}
	return json.Marshal(out)
}

func (mx Message) ToBase64() string {
	out, _ := mx.MarshalBinary()
	return base64.StdEncoding.EncodeToString(out)
}

func (mx *Message) UnmarshalBase64(b64 string) error {
	b, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return err
	}
	return mx.UnmarshalWithDecoder(bin.NewBinDecoder(b))
}

func (msg *Message) SetAddressTables(tables map[PublicKey]PublicKeySlice) error{
  if msg.addressTables != nil{
    return fmt.Errorf("address tables already set")
  }
  msg.addressTables = tables
  return nil
}

func (msg *Message) SetAddressTableLookups(lookups []MessageAddressTableLookup) *Message{
  msg.AddressTableLookups = lookups
  msg.version = MessageVersionV0
  return msg
}
