package middleware

import (
	"gaming-leaderboard/util"
	"net/url"

	"github.com/gin-gonic/gin"
)

func SanitizeQueryParams() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		queryParams := make(url.Values)
		for k, v := range ctx.Request.URL.Query() {
			queryParams[util.TrimSpace(k)] = util.TrimStrings(v)
		}

		ctx.Request.URL.RawQuery = queryParams.Encode()
		ctx.Next()
	}
}
