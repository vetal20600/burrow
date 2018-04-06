package v0

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hyperledger/burrow/execution/evm/events"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/txs"
)

type V0Client struct {
	url    string
	codec  rpc.Codec
	client *http.Client
}

type RPCResponse struct {
	Result  json.RawMessage `json:"result"`
	Error   *rpc.RPCError   `json:"error"`
	Id      string          `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
}

func NewV0Client(url string) *V0Client {
	return &V0Client{
		url:   url,
		codec: NewTCodec(),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (vc *V0Client) Transact(param TransactParam) (*txs.Receipt, error) {
	receipt := new(txs.Receipt)
	err := vc.Call(TRANSACT, param, receipt)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (vc *V0Client) TransactAndHold(param TransactParam) (*events.EventDataCall, error) {
	eventDataCall := new(events.EventDataCall)
	err := vc.Call(TRANSACT_AND_HOLD, param, eventDataCall)
	if err != nil {
		return nil, err
	}
	return eventDataCall, nil
}

func (vc *V0Client) Call(method string, param interface{}, result interface{}) error {
	// Marhsal into JSONRPC request object
	bs, err := vc.codec.EncodeBytes(param)
	if err != nil {
		return err
	}
	request := rpc.NewRPCRequest("test", method, bs)
	bs, err = json.Marshal(request)
	if err != nil {
		return err
	}
	// Post to JSONService
	resp, err := vc.client.Post(vc.url, "application/json", bytes.NewBuffer(bs))
	if err != nil {
		return err
	}
	// Marshal into JSONRPC response object
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	rpcResponse := new(RPCResponse)
	err = json.Unmarshal(bs, rpcResponse)
	if err != nil {
		return err
	}
	if rpcResponse.Error != nil {
		return rpcResponse.Error
	}
	err = json.Unmarshal(rpcResponse.Result, result)
	if err != nil {
		return err
	}
	return nil
}
