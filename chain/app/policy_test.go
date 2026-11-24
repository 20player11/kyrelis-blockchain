package app

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"testing"
)

type feeProbe struct {
	sdk.FeeTx
	coins sdk.Coins
}

func (f feeProbe) GetFee() sdk.Coins  { return f.coins }
func (f feeProbe) GetMsgs() []sdk.Msg { return []sdk.Msg{&bank.MsgSend{}} }
func TestNativeRules(t *testing.T) {
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return ctx, nil }
	for _, n := range []int64{0, 99, 100, 101} {
		_, err := withTransferFee(next)(sdk.Context{}, feeProbe{coins: sdk.NewCoins(sdk.NewInt64Coin("ukyrl", n))}, false)
		if (err == nil) != (n >= 100) {
			t.Fatalf("fee %d: %v", n, err)
		}
	}
	if err := (voucherBank{}).MintCoins(context.Background(), "transfer", sdk.NewCoins(sdk.NewInt64Coin("ukyrl", 1))); err == nil {
		t.Fatal("native mint allowed")
	}
}
