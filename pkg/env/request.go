package env

import (
	"context"
	"gaming-leaderboard/constants"

	"github.com/google/uuid"
)

func GetRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value(constants.RequestID).(string)
	if !ok {
		return uuid.New().String()
	}

	return requestID
}
