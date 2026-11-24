package app

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// IBC may mint vouchers. The native asset is issued only at genesis.
type voucherBank struct{ bankkeeper.Keeper }

func (b voucherBank) MintCoins(ctx context.Context, module string, coins sdk.Coins) error {
	if coins.AmountOf("ukyrl").IsPositive() {
		return fmt.Errorf("native KYRL issuance is closed")
	}
	return b.Keeper.MintCoins(ctx, module, coins)
}
