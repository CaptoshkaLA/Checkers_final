package keeper

import (
	"context"
	"fmt"
	"strings"

	rules "github.com/captoshkala/checkers/x/checkers/rules"
	"github.com/captoshkala/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) ForfeitExpiredGames(goCtx context.Context) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	opponents := map[string]string{
		rules.BLACK_PLAYER.Color: rules.RED_PLAYER.Color,
		rules.RED_PLAYER.Color:   rules.BLACK_PLAYER.Color,
	}
	nextGame, found := k.GetNextGame(ctx)
	if !found {
		panic("Next game not found")
	}
	storedGameId := nextGame.FifoHead
	var storedGame types.StoredGame
	for {
		if strings.Compare(storedGameId, types.NoFifoIdKey) == 0 {
			break
		}
		storedGame, found = k.GetStoredGame(ctx, storedGameId)
		if !found {
			panic("fifo head game not found" + nextGame.FifoHead)
		}
		deadline, err := storedGame.GetDeadlineAsTime()
		if err != nil {
			panic(err)
		}
		if deadline.Before(ctx.BlockTime()) {
			k.RemoveFromFifo(ctx, &storedGame, &nextGame)
			if storedGame.MoveCount == 0 {
				storedGame.Winner = rules.NO_PLAYER.Color
				k.RemoveStoredGame(ctx, storedGameId)
			} else {
				storedGame.Winner, found = opponents[storedGame.Turn]
				if !found {
					panic(fmt.Sprintf(types.ErrCannotFindWinnerByColor.Error(), storedGame.Turn))
				}
				if storedGame.MoveCount <= 1 {
					k.MustRefundWager(ctx, &storedGame)
				} else {
					k.MustPayWinnings(ctx, &storedGame)
				}
				k.SetStoredGame(ctx, storedGame)
			}
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(sdk.EventTypeMessage,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(sdk.AttributeKeyAction, types.ForfeitGameEventKey),
					sdk.NewAttribute(types.ForfeitGameEventIdValue, storedGameId),
					sdk.NewAttribute(types.ForfeitGameEventWinner, storedGame.Winner),
				),
			)
		} else {
			break
		}
		storedGameId = nextGame.FifoHead
	}
	k.SetNextGame(ctx, nextGame)
}
