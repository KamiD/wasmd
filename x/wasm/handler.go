package wasm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	AttributeKeyContract = "contract_address"
	AttributeKeyCodeID   = "code_id"
)

// NewHandler returns a handler for "bank" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case MsgStoreCode:
			return handleStoreCode(ctx, k, &msg)
		case *MsgStoreCode:
			return handleStoreCode(ctx, k, msg)

		case MsgInstantiateContract:
			return handleInstantiate(ctx, k, &msg)
		case *MsgInstantiateContract:
			return handleInstantiate(ctx, k, msg)

		case MsgExecuteContract:
			return handleExecute(ctx, k, &msg)
		case *MsgExecuteContract:
			return handleExecute(ctx, k, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized wasm message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleStoreCode(ctx sdk.Context, k Keeper, msg *MsgStoreCode) sdk.Result {
	codeID, err := k.Create(ctx, msg.Sender, msg.WASMByteCode)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, "store-code"),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(AttributeKeyCodeID, fmt.Sprintf("%d", codeID)),
		),
	)

	return sdk.Result{
		Data:   []byte(fmt.Sprintf("%d", codeID)),
		Events: ctx.EventManager().Events(),
	}
}

func handleInstantiate(ctx sdk.Context, k Keeper, msg *MsgInstantiateContract) sdk.Result {
	contractAddr, err := k.Instantiate(ctx, msg.Sender, msg.Code, msg.InitMsg, msg.InitFunds)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, "instantiate"),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(AttributeKeyCodeID, fmt.Sprintf("%d", msg.Code)),
			sdk.NewAttribute(AttributeKeyContract, contractAddr.String()),
		),
	)

	return sdk.Result{
		Data:   contractAddr,
		Events: ctx.EventManager().Events(),
	}
}

func handleExecute(ctx sdk.Context, k Keeper, msg *MsgExecuteContract) sdk.Result {
	res, err := k.Execute(ctx, msg.Contract, msg.Sender, msg.SentFunds, msg.Msg)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, "execute"),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(AttributeKeyContract, msg.Contract.String()),
		),
	)

	res.Events = append(res.Events, ctx.EventManager().Events()...)
	return res
}
