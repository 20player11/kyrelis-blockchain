package app

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Charge direct bank transfers and delegated execution, including nested transfers.
func withTransferFee(next sdk.AnteHandler) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		for _, msg := range tx.GetMsgs() {
			switch msg.(type) {
			case *bank.MsgSend, *bank.MsgMultiSend, *authz.MsgExec:
				fee, ok := tx.(sdk.FeeTx)
				if !ok || fee.GetFee().AmountOf("ukyrl").LT(math.NewInt(100)) {
					return ctx, errorsmod.Wrap(sdkerrors.ErrInsufficientFee, "KYRELIS requires at least 100ukyrl for transfers")
				}
			}
		}
		return next(ctx, tx, simulate)
	}
}
