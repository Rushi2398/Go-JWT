package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func CheckUserType(ctx *gin.Context, role string) (err error) {
	userType := ctx.GetString("user_type")
	if userType != role {
		return errors.New("unauthorised to access this resource")
	}
	return nil
}

func MatchUserTypeToUid(ctx *gin.Context, userId string) (err error) {
	userType := ctx.GetString("user_type")
	uid := ctx.GetString("uid")

	if userType == "USER" && uid != userId {
		return errors.New("unauthorised to access this resource")
	}
	return CheckUserType(ctx, userType)
}
