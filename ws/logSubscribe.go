package ws

import (
	"context"
	"errors"

	"github.com/scatkit/pumpdexer/rpc"
	"github.com/scatkit/pumpdexer/solana"
)

type LogResult struct {
	Context struct {
		Slot uint64
	} `json:"context"`
	Value struct {
		Signature solana.Signature `json:"signature"`
		Err       interface{}      `json:"err"`
		// Array of log messages the transaction instructions output
		// during execution, null if simulation failed before the transaction
		// was able to execute (for example due to an invalid blockhash
		// or signature verification failure)
		Logs []string `json:"logs"`
	} `json:"value"`
}

type LogsSubscribeFilterType string

const (
	LogsSubcribeFilterAll          LogsSubscribeFilterType = "all"
	LogsSubcribeFilterAllWithVotes LogsSubscribeFilterType = "allWithVotes"
)

func (cl *Client) LogSubscribe(filter LogsSubscribeFilterType, commitment rpc.CommitmentType) (*LogSubscription, error) {
	return cl.logSubscribe(filter, commitment)
}

func (cl *Client) LogSubscribeToAddress(mentions solana.PublicKey, commitment rpc.CommitmentType) (*LogSubscription, error) {
	return cl.logSubscribe(
		map[string]interface{}{
			"mentions": []string{mentions.String()}, // mentions is an array of a signle pubkey put in the object
		},
		commitment,
	)
}

func (cl *Client) logSubscribe(filter interface{}, commitment rpc.CommitmentType) (*LogSubscription, error) {
	params := []interface{}{filter}
	conf := map[string]interface{}{}
	if commitment != "" {
		conf["commitment"] = commitment
	}

	genSub, err := cl.subscribe(
		params,
		conf,
		"logsSubscribe",
		"logsUnsubscribe",
		func(msg []byte) (interface{}, error) {
			var res LogResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}

	return &LogSubscription{
		sub: genSub,
	}, nil

}

type LogSubscription struct {
	sub *Subscription
}

func (ls *LogSubscription) Recv(ctx context.Context) (*LogResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-ls.sub.stream:
		if !ok {
			return nil, errors.New("Subscription is closed")
		}
		return d.(*LogResult), nil
	case err := <-ls.sub.err:
		return nil, err
	}
}

func (ls *LogSubscription) Err() <-chan error {
	return ls.sub.err
}

func (ls *LogSubscription) Response() <-chan *LogResult {
	typedChan := make(chan *LogResult, 1)
	go func(ch chan *LogResult) {
		d, ok := <-ls.sub.stream
		if !ok {
			return
		}
		ch <- d.(*LogResult)
	}(typedChan)

	return typedChan
}

func (ls *LogSubscription) Unsubscribe() {
	ls.sub.Unsubscribe()
}
