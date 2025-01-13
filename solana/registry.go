package solana
import(
  "sync"
  "fmt"
  "reflect"
)

// InstructionDecoder receives the AccountMeta FOR THAT INSTRUCTION,
// and not the accounts of the *Message object. Resolve with
// CompiledInstruction.ResolveInstructionAccounts(message) beforehand.
type InstructionDecoder func(instructionAccounts []*AccountMeta, data []byte) (interface{}, error)

var instructionDecoderRegistry = newInstructionDecoderRegistry()

type decoderRegistry struct {
	mu       *sync.RWMutex
	decoders map[PublicKey]InstructionDecoder
}

func newInstructionDecoderRegistry() *decoderRegistry {
	return &decoderRegistry{
		mu:       &sync.RWMutex{},
		decoders: make(map[PublicKey]InstructionDecoder),
	}
}

//func (reg *decoderRegistry) Has(programID PublicKey) bool {
//	reg.mu.RLock()
//	defer reg.mu.RUnlock()
//
//	_, ok := reg.decoders[programID]
//	return ok
//}


func RegisterInstructionDecoder(programID PublicKey, decoder InstructionDecoder) {
	prev, has := instructionDecoderRegistry.Get(programID)
	if has {
		// If it's the same function, then OK (tollerate multiple calls with same params).
		if isSameFunction(prev, decoder) {
			return
		}
		// If it's another decoder for the same pubkey, then panic.
		panic(fmt.Sprintf("unable to re-register instruction decoder for program %s", programID))
	}
	instructionDecoderRegistry.RegisterIfNew(programID, decoder)
}

func (reg *decoderRegistry) Get(programID PublicKey) (InstructionDecoder, bool) {
  reg.mu.RLock()
  defer reg.mu.RUnlock()

  decoder, ok := reg.decoders[programID]
  return decoder, ok
}

func isSameFunction(f1 interface{}, f2 interface{}) bool {
	return reflect.ValueOf(f1).Pointer() == reflect.ValueOf(f2).Pointer()
}

// RegisterIfNew registers the provided decoder for the provided programID ONLY if there isn't
// already a registered decoder for the programID.
// Returns true if was successfully registered right now (non-previously registered);
// returns false if there already was a decoder registered.
func (reg *decoderRegistry) RegisterIfNew(programID PublicKey, decoder InstructionDecoder) bool {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	_, ok := reg.decoders[programID]
	if ok {
		return false
	}
	reg.decoders[programID] = decoder
	return true
}
