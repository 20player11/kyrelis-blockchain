package realrpc

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, e := New(srv.URL, srv.URL)
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func localhostServer(t *testing.T, h http.Handler) *httptest.Server {
	t.Helper()
	l, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	s := httptest.NewUnstartedServer(h)
	s.Listener = l
	s.Start()
	t.Cleanup(s.Close)
	return s
}
func TestAllowlist(t *testing.T) {
	for _, x := range []string{"https://127.0.0.1:26657", "http://example.com:26657", "http://localhost:26657", "http://127.0.0.1:1", "http://127.0.0.1:26657@bad.tld:99", "http://127.0.0.1:26657/x", "http://127.0.0.1:26657?x=1"} {
		if _, e := ValidateLocalURL(x); e == nil {
			t.Errorf("accepted %s", x)
		}
	}
	for _, x := range []string{"http://127.0.0.1:26657", "http://[::1]:26657"} {
		if _, e := ValidateLocalURL(x); e != nil {
			t.Errorf("rejected %s: %v", x, e)
		}
	}
}
func TestRealRPCParsing(t *testing.T) {
	s := localhostServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status":
			fmt.Fprint(w, `{"result":{"node_info":{"id":"node1","network":"kyrelis-devnet-1"},"sync_info":{"latest_block_height":"9","latest_block_hash":"ABC","catching_up":false},"validator_info":{"voting_power":"10"}}}`)
		case "/net_info":
			fmt.Fprint(w, `{"result":{"n_peers":"3"}}`)
		case "/validators":
			fmt.Fprint(w, `{"result":{"total":"1","validators":[{"address":"001","voting_power":"10"}]}}`)
		case "/block":
			fmt.Fprintf(w, `{"result":{"block_id":{"hash":"ABC"},"block":{"header":{"height":"9","time":"2026-11-22T12:00:00Z","last_block_id":{"hash":"DEF"}},"data":{"txs":[%q]}}}}`, base64.StdEncoding.EncodeToString([]byte("real signed transaction bytes")))
		case "/cosmos/bank/v1beta1/supply/by_denom":
			fmt.Fprint(w, `{"amount":{"denom":"ukyrl","amount":"21000000000000"}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	c := newTestClient(t, s)
	v := c.Status(context.Background())
	if !v.Connected || v.Peers != 3 || v.Height != "9" {
		t.Fatalf("status %+v", v)
	}
	vals, e := c.Validators(context.Background())
	if e != nil || len(vals) != 1 {
		t.Fatalf("validators %v %+v", e, vals)
	}
	b, e := c.Block(context.Background(), 9)
	if e != nil || b.TxCount != 1 || len(b.Transactions[0]) != 64 {
		t.Fatalf("block %v %+v", e, b)
	}
	amount, e := c.Supply(context.Background())
	if e != nil || amount != "21000000000000" {
		t.Fatalf("supply %v %s", e, amount)
	}
}
func TestWrongChainRejected(t *testing.T) {
	s := localhostServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, strings.Replace(`{"result":{"node_info":{"network":"kyrelis-devnet-1"}}}`, "kyrelis-devnet-1", "some-other-network", 1))
	}))
	c := newTestClient(t, s)
	if c.Status(context.Background()).Connected {
		t.Fatal("connected to wrong chain")
	}
}

func TestBroadcastSignedTxToVerifiedLocalChain(t *testing.T) {
	posted := false
	srv := localhostServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status":
			fmt.Fprint(w, `{"result":{"node_info":{"id":"node1","network":"kyrelis-devnet-1"},"sync_info":{"latest_block_height":"10"}}}`)
		case "/net_info":
			fmt.Fprint(w, `{"result":{"n_peers":"1"}}`)
		case "/cosmos/tx/v1beta1/txs":
			if r.Method != "POST" {
				t.Error("expected POST")
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Error("expected JSON content type")
			}
			posted = true
			fmt.Fprint(w, `{"tx_response":{"txhash":"AABBCC","code":0,"raw_log":""}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	c := newTestClient(t, srv)
	result, err := c.BroadcastTx(context.Background(), base64.StdEncoding.EncodeToString([]byte("some signed protocol buffer payload")))
	if err != nil || result.TxHash != "AABBCC" || !posted {
		t.Fatalf("result=%+v, err=%v posted=%t", result, err, posted)
	}
	if _, err := c.BroadcastTx(context.Background(), "THIS-IS-NOT-BASE64"); err == nil {
		t.Fatal("invalid base64 accepted")
	}
}

func TestBroadcastRejectsWrongChain(t *testing.T) {
	posted := false
	srv := localhostServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status":
			fmt.Fprint(w, `{"result":{"node_info":{"network":"not-kyrelis"}}}`)
		case "/cosmos/tx/v1beta1/txs":
			posted = true
		default:
			http.NotFound(w, r)
		}
	}))
	c := newTestClient(t, srv)
	_, err := c.BroadcastTx(context.Background(), base64.StdEncoding.EncodeToString([]byte("some signed protocol buffer payload")))
	if err == nil || posted {
		t.Fatalf("wrong chain not blocked: %v", err)
	}
}
