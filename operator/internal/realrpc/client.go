// Package realrpc reads ACTUAL CometBFT RPC responses and Cosmos SDK REST data.
// It does not implement consensus, generate transactions, or invent chain state.
package realrpc

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const RequiredChainID = "kyrelis-devnet-1"

// Only explicit loopback targets are supported; never proxy to an arbitrary hostname.
func ValidateLocalURL(value string) (string, error) {
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", errors.New("povolena je jen lokalni http://127.0.0.1:PORT adresa")
	}
	host := u.Hostname()
	if host != "127.0.0.1" && host != "::1" {
		return "", errors.New("RPC musi byt na 127.0.0.1 nebo ::1")
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil || port < 1024 || port > 65535 {
		return "", errors.New("neplatny port")
	}
	return strings.TrimRight(u.String(), "/"), nil
}

type Client struct {
	RPC  string
	REST string
	HTTP *http.Client
}

func New(rpc, rest string) (*Client, error) {
	var err error
	if rpc, err = ValidateLocalURL(rpc); err != nil {
		return nil, fmt.Errorf("RPC: %w", err)
	}
	if rest, err = ValidateLocalURL(rest); err != nil {
		return nil, fmt.Errorf("REST: %w", err)
	}
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.Proxy = nil
	t.DialContext = (&net.Dialer{Timeout: 2 * time.Second}).DialContext
	return &Client{RPC: rpc, REST: rest, HTTP: &http.Client{Timeout: 3 * time.Second, Transport: t}}, nil
}
func (c *Client) get(ctx context.Context, base, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", base+path, nil)
	if err != nil {
		return err
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("RPC/REST vratilo HTTP %d", res.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("neplatna odpoved uzlu: %w", err)
	}
	return nil
}

type RPCStatus struct {
	Result struct {
		NodeInfo struct {
			ID      string `json:"id"`
			Network string `json:"network"`
			Version string `json:"version"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHash   string `json:"latest_block_hash"`
			LatestBlockHeight string `json:"latest_block_height"`
			LatestBlockTime   string `json:"latest_block_time"`
			CatchingUp        bool   `json:"catching_up"`
		} `json:"sync_info"`
		ValidatorInfo struct {
			Address     string `json:"address"`
			VotingPower string `json:"voting_power"`
		} `json:"validator_info"`
	} `json:"result"`
	Error any `json:"error"`
}
type Status struct {
	Connected   bool   `json:"connected"`
	ChainID     string `json:"chain_id"`
	NodeID      string `json:"node_id"`
	Height      string `json:"height"`
	LatestHash  string `json:"latest_hash"`
	LatestTime  string `json:"latest_time"`
	CatchingUp  bool   `json:"catching_up"`
	Peers       int    `json:"peers"`
	VotingPower string `json:"own_voting_power"`
	Error       string `json:"error,omitempty"`
}

func (c *Client) Status(ctx context.Context) Status {
	var s RPCStatus
	if err := c.get(ctx, c.RPC, "/status", &s); err != nil {
		return Status{Error: err.Error()}
	}
	if s.Error != nil {
		return Status{Error: "CometBFT status returned an error"}
	}
	if s.Result.NodeInfo.Network != RequiredChainID {
		return Status{Error: fmt.Sprintf("Jiny blockchain: %s (ocekavan %s)", s.Result.NodeInfo.Network, RequiredChainID), ChainID: s.Result.NodeInfo.Network}
	}
	out := Status{Connected: true, ChainID: s.Result.NodeInfo.Network, NodeID: s.Result.NodeInfo.ID, Height: s.Result.SyncInfo.LatestBlockHeight, LatestHash: s.Result.SyncInfo.LatestBlockHash, LatestTime: s.Result.SyncInfo.LatestBlockTime, CatchingUp: s.Result.SyncInfo.CatchingUp, VotingPower: s.Result.ValidatorInfo.VotingPower}
	var ni struct {
		Result struct {
			NPeers string `json:"n_peers"`
		} `json:"result"`
	}
	if c.get(ctx, c.RPC, "/net_info", &ni) == nil {
		out.Peers, _ = strconv.Atoi(ni.Result.NPeers)
	}
	return out
}

type Validator struct {
	Address          string `json:"address"`
	VotingPower      string `json:"voting_power"`
	ProposerPriority string `json:"proposer_priority"`
}

func (c *Client) Validators(ctx context.Context) ([]Validator, error) {
	var response struct {
		Result struct {
			Validators []Validator `json:"validators"`
			Total      string      `json:"total"`
		} `json:"result"`
	}
	if err := c.get(ctx, c.RPC, "/validators?per_page=100", &response); err != nil {
		return nil, err
	}
	total, _ := strconv.Atoi(response.Result.Total)
	if total > 100 {
		return nil, fmt.Errorf("%d validatoru: zobrazuji pouze prvnich 100; pro dalsi je nutna strankovani", total)
	}
	if response.Result.Validators == nil {
		return []Validator{}, nil
	}
	return response.Result.Validators, nil
}

type Block struct {
	Height       string   `json:"height"`
	Hash         string   `json:"hash"`
	Time         string   `json:"time"`
	PreviousHash string   `json:"previous_hash"`
	Transactions []string `json:"transactions"`
	TxCount      int      `json:"tx_count"`
}

func (c *Client) Block(ctx context.Context, height int64) (Block, error) {
	if height < 1 {
		return Block{}, errors.New("vyska bloku musi byt kladna")
	}
	var response struct {
		Result struct {
			BlockID struct {
				Hash string `json:"hash"`
			} `json:"block_id"`
			Block struct {
				Header struct {
					Height      string `json:"height"`
					Time        string `json:"time"`
					LastBlockID struct {
						Hash string `json:"hash"`
					} `json:"last_block_id"`
				} `json:"header"`
				Data struct {
					Txs []string `json:"txs"`
				} `json:"data"`
			} `json:"block"`
		} `json:"result"`
	}
	if err := c.get(ctx, c.RPC, fmt.Sprintf("/block?height=%d", height), &response); err != nil {
		return Block{}, err
	}
	x := Block{Height: response.Result.Block.Header.Height, Hash: response.Result.BlockID.Hash, Time: response.Result.Block.Header.Time, PreviousHash: response.Result.Block.Header.LastBlockID.Hash}
	for _, base64tx := range response.Result.Block.Data.Txs {
		raw, err := base64.StdEncoding.DecodeString(base64tx)
		if err != nil {
			return Block{}, fmt.Errorf("neplatne zakodovana transakce: %w", err)
		}
		hash := sha256.Sum256(raw)
		x.Transactions = append(x.Transactions, strings.ToUpper(hex.EncodeToString(hash[:])))
	}
	x.TxCount = len(x.Transactions)
	return x, nil
}
func (c *Client) Balance(ctx context.Context, address string) (string, error) {
	if len(address) < 10 || len(address) > 100 || !strings.HasPrefix(address, "kyr1") {
		return "", errors.New("adresa musi zacinat kyr1")
	}
	for _, r := range address {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyz0123456789", r) {
			return "", errors.New("neplatne znaky adresy")
		}
	}
	var resp struct {
		Balances []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"balances"`
	}
	err := c.get(ctx, c.REST, "/cosmos/bank/v1beta1/balances/"+url.PathEscape(address), &resp)
	if err != nil {
		return "", err
	}
	for _, b := range resp.Balances {
		if b.Denom == "ukyrl" {
			return b.Amount, nil
		}
	}
	return "0", nil
}
func (c *Client) Supply(ctx context.Context) (string, error) {
	var resp struct {
		Amount struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"amount"`
	}
	if err := c.get(ctx, c.REST, "/cosmos/bank/v1beta1/supply/by_denom?denom=ukyrl", &resp); err != nil {
		return "", err
	}
	if resp.Amount.Denom != "ukyrl" {
		return "", errors.New("uzel nevratil token ukyrl")
	}
	return resp.Amount.Amount, nil
}

// BroadcastTx forwards complete, externally signed TxRaw bytes to the VERIFIED
// local KYRELIS devnet only. Never handles private keys, seed phrases or arbitrary RPC hosts.
type BroadcastResponse struct {
	TxHash string `json:"txhash"`
	Code   uint32 `json:"code"`
	RawLog string `json:"raw_log"`
}

func (c *Client) BroadcastTx(ctx context.Context, encoded string) (BroadcastResponse, error) {
	if len(encoded) == 0 || len(encoded) > 700000 {
		return BroadcastResponse{}, errors.New("invalid signed transaction size")
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || len(decoded) < 10 || len(decoded) > 512*1024 {
		return BroadcastResponse{}, errors.New("signed transaction must be a valid base64 protobuf payload (10B–512KiB)")
	}
	status := c.Status(ctx)
	if !status.Connected || status.ChainID != RequiredChainID {
		return BroadcastResponse{}, errors.New("cannot broadcast: local node is offline or connected to another chain")
	}
	body, _ := json.Marshal(map[string]string{"tx_bytes": encoded, "mode": "BROADCAST_MODE_SYNC"})
	req, err := http.NewRequestWithContext(ctx, "POST", c.REST+"/cosmos/tx/v1beta1/txs", strings.NewReader(string(body)))
	if err != nil {
		return BroadcastResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return BroadcastResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return BroadcastResponse{}, fmt.Errorf("transaction broadcast returned HTTP %d", resp.StatusCode)
	}
	var raw struct {
		TxResponse struct {
			TxHash string `json:"txhash"`
			Code   uint32 `json:"code"`
			RawLog string `json:"raw_log"`
		} `json:"tx_response"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&raw); err != nil {
		return BroadcastResponse{}, fmt.Errorf("invalid node broadcast response: %w", err)
	}
	if raw.TxResponse.Code != 0 {
		return BroadcastResponse{}, fmt.Errorf("node rejected transaction (code %d): %s", raw.TxResponse.Code, raw.TxResponse.RawLog)
	}
	if raw.TxResponse.TxHash == "" {
		return BroadcastResponse{}, errors.New("node did not return a transaction hash")
	}
	return BroadcastResponse{TxHash: raw.TxResponse.TxHash, Code: raw.TxResponse.Code, RawLog: raw.TxResponse.RawLog}, nil
}
