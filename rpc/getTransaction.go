package rpc
import(
  "context"
  "errors"
  "fmt"
  "encoding/json"
  
  "github.com/scatkit/pumpdexer/solana"
  bin "github.com/gagliardetto/binary"
)

// TransactionResultEnvelope will contain a *solana.Transaction if the requested encoding is `solana.EncodingJSON`
// (which is also the default when the encoding is not specified),
// or a `solana.Data` in case of EncodingBase58, EncodingBase64.
type TransactionResultEnvelope struct {
	asDecodedBinary     solana.Data
	asParsedTransaction *solana.Transaction
}

func (wrap TransactionResultEnvelope) MarshalJSON() ([]byte, error){
  if wrap.asParsedTransaction != nil{
    return json.Marshal(wrap.asParsedTransaction)
  }
  return json.Marshal(wrap.asDecodedBinary)
}

func (wrap *TransactionResultEnvelope) UnmarshalJSON(data []byte) error{
  if len(data) == 0 || (len(data) == 4 && string(data) == "null") {
		// TODO: is this an error?
		return nil
	}
  firstChar := data[0]
  
  switch firstChar{
    case '[': // <- base64(or similar)
    err := wrap.asDecodedBinary.UnmarshalJSON(data)
			if err != nil {
				return err
			}
    case '{': // <- likely JSON
      {
        return json.Unmarshal(data, &wrap.asParsedTransaction)
      }
    default:
      return fmt.Errorf("Unknown encoding: %v", data)
  }
  return nil 
}

// GetBinary returns the decoded bytes if the encoding is
// "base58", "base64".
func (dt *TransactionResultEnvelope) GetBinary() []byte {
	return dt.asDecodedBinary.Content
}

func (dt *TransactionResultEnvelope) GetData() solana.Data {
	return dt.asDecodedBinary
}

type GetTransactionOpts struct{
  Encoding    solana.EncodingType `json:"encoding,omitempty"`
  Commitment  CommitmentType      `json:"commitment,omitempty"`
  // Max transaction version to return in responses.
	// If the requested block contains a transaction with a higher version, an error will be returned.
  MaxSupportedTransactionVersion *uint64
}

type GetTransactionResult struct{
  Slot        uint64 `json:"slot"`
  BlockTime   *solana.UnixTimeSeconds `json:"blockTime" bin:"optional"`
  Transaction *TransactionResultEnvelope  `json:"transaction" bin:"optional"`
  Meta        *TransactionMeta           `json:"meta,omitempty" bin:"optional"`
	Version     TransactionVersion         `json:"version"`
}

func (cl *Client) GetTransaction(ctx context.Context, txSig solana.Signature, opts *GetTransactionOpts,
) (out *GetTransactionResult, err error){
  params := []interface{}{txSig}
  if opts != nil{
    obj := map[string]interface{}{}
    if opts.Encoding != ""{
      if !solana.IsAnyOfEncodingType(
				opts.Encoding,
				// Valid encodings:
				// solana.EncodingJSON, // TODO
				solana.EncodingJsonParsed, // TODO
				solana.EncodingBase58,
				solana.EncodingBase64,
				solana.EncodingBase64Zstd,
			) {
				return nil, fmt.Errorf("provided encoding is not supported: %s", opts.Encoding)
			}
			obj["encoding"] = opts.Encoding
    }
    if opts.Commitment != "" {
			obj["commitment"] = opts.Commitment
		}
    if opts.MaxSupportedTransactionVersion != nil {
			obj["maxSupportedTransactionVersion"] = *opts.MaxSupportedTransactionVersion
		}
    if len(obj) > 0 {
			params = append(params, obj)
		}
  }
  err = cl.rpcClient.CallForInfo(ctx, &out, "getTransaction", params)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, errors.New("not found")
	}
	return out, err
}

// GetRawJSON returns a *solana.Transaction when the data
// encoding is EncodingJSON.
func (dt *TransactionResultEnvelope) GetTransaction() (*solana.Transaction, error) {
	if dt.asDecodedBinary.Content != nil {
		tx := new(solana.Transaction)
		err := tx.UnmarshalWithDecoder(bin.NewBinDecoder(dt.asDecodedBinary.Content))
		if err != nil {
			return nil, err
		}
		return tx, nil
	}
	return dt.asParsedTransaction, nil
}

func (obj TransactionResultEnvelope) MarshalWithEncoder(encoder *bin.Encoder) (err error) {
	return encoder.Encode(obj.asDecodedBinary)
}

func (obj *TransactionResultEnvelope) UnmarshalWithDecoder(decoder *bin.Decoder) (err error) {
	return decoder.Decode(&obj.asDecodedBinary)
}

func (obj GetTransactionResult) MarshalWithEncoder(encoder *bin.Encoder) (err error) {
	err = encoder.WriteUint64(obj.Slot, bin.LE)
	if err != nil {
		return err
	}
	{
		if obj.BlockTime == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.WriteInt64(int64(*obj.BlockTime), bin.LE)
			if err != nil {
				return err
			}
		}
	}
	{
		if obj.Transaction == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = obj.Transaction.MarshalWithEncoder(encoder)
			if err != nil {
				return err
			}
		}
	}
	{
		if obj.Meta == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			// NOTE: storing as JSON bytes:
			buf, err := json.Marshal(obj.Meta)
			if err != nil {
				return err
			}
			err = encoder.WriteBytes(buf, true)
			if err != nil {
				return err
			}
		}
	}
	{
		buf, err := json.Marshal(obj.Version)
		if err != nil {
			return err
		}
		err = encoder.WriteBytes(buf, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (obj *GetTransactionResult) UnmarshalWithDecoder(decoder *bin.Decoder) (err error) {
	// Deserialize `Slot`:
	obj.Slot, err = decoder.ReadUint64(bin.LE)
	if err != nil {
		return err
	}
	// Deserialize `BlockTime` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.BlockTime)
			if err != nil {
				return err
			}
		}
	}
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			obj.Transaction = new(TransactionResultEnvelope)
			err = obj.Transaction.UnmarshalWithDecoder(decoder)
			if err != nil {
				return err
			}
		}
	}
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			// NOTE: storing as JSON bytes:
			buf, err := decoder.ReadByteSlice()
			if err != nil {
				return err
			}
			err = json.Unmarshal(buf, &obj.Meta)
			if err != nil {
				return err
			}
		}
	}
	{
		buf, err := decoder.ReadByteSlice()
		if err != nil {
			return err
		}
		err = json.Unmarshal(buf, &obj.Version)
		if err != nil {
			return err
		}
	}
	return nil
}
