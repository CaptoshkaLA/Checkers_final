package keeper

import (
	"context"
	"strconv"

	rules "github.com/captoshkala/checkers/x/checkers/rules"
	"github.com/captoshkala/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CreateGame(goCtx context.Context, msg *types.MsgCreateGame) (*types.MsgCreateGameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	nextGame, found := k.Keeper.GetNextGame(ctx)
	if !found {
		panic("next game not found")
	}
	newIndex := strconv.FormatUint(nextGame.IdValue, 10)
	storedGame := types.StoredGame{
		Creator:   msg.Creator,
		Index:     newIndex,
		Game:      rules.New().String(),
		Red:       msg.Red,
		Black:     msg.Black,
		MoveCount: 0,
		Deadline:  types.FormatDeadline(types.GetNextDeadline(ctx)),
		Winner:    rules.NO_PLAYER.Color,
		Wager:     msg.Wager,
		Token:     msg.Token,
	}
	err := storedGame.Validate()
	if err != nil {
		return nil, err
	}
	k.Keeper.sendToFifoTail(ctx, &storedGame, &nextGame)
	k.Keeper.SetStoredGame(ctx, storedGame)
	nextGame.IdValue++
	k.Keeper.SetNextGame(ctx, nextGame)
	ctx.GasMeter().ConsumeGas(types.CreateGameGas, "create game")
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, "checkers"),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.StoredGameEventKey),
			sdk.NewAttribute(types.StoredGameEventCreator, msg.Creator),
			sdk.NewAttribute(types.StoredGameEventIndex, newIndex),
			sdk.NewAttribute(types.StoredGameEventRed, msg.Red),
			sdk.NewAttribute(types.StoredGameEventBlack, msg.Black),
			sdk.NewAttribute(types.StoredGameEventWager, strconv.FormatUint(msg.Wager, 10)),
			sdk.NewAttribute(types.StoredGameEventToken, msg.Token),
		),
	)
	return &types.MsgCreateGameResponse{IdValue: newIndex}, nil
}
