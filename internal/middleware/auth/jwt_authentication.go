package auth_middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"jank.com/jank_blog/internal/global"
	"jank.com/jank_blog/internal/utils"
	"jank.com/jank_blog/pkg/serve/controller/account"
)

// JWTConfig 用于配置 JWT 中间件
type JWTConfig struct {
	Authorization   string
	TokenPrefix     string
	RefreshToken    string
	RedisPrefix     string
	LocalsUserIdKey string
}

// DefaultJWTConfig 提供默认的 JWT 配置
var DefaultJWTConfig = JWTConfig{
	Authorization:   "Authorization",
	TokenPrefix:     "Bearer ",
	RefreshToken:    "Refresh_Token",
	RedisPrefix:     "ACC_AUTH_TOKEN_CACHE_PREFIX",
	LocalsUserIdKey: "Locals_User_Id",
}

// JWTMiddleware 用于处理请求的 JWT 验证和 Token 刷新
func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(DefaultJWTConfig.Authorization)
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "缺少 Authorization Header")
			}

			tokenString := strings.TrimPrefix(authHeader, DefaultJWTConfig.TokenPrefix)
			token, err := utils.ValidateJWTToken(tokenString, false)
			if err != nil {
				refreshHeader := c.Request().Header.Get(DefaultJWTConfig.RefreshToken)
				if refreshHeader == "" {
					return echo.NewHTTPError(http.StatusUnauthorized, "无效 Access Token，请重新登录")
				}

				refreshTokenString := strings.TrimPrefix(refreshHeader, DefaultJWTConfig.TokenPrefix)
				newTokens, refreshErr := refreshTokenLogic(refreshTokenString)
				if refreshErr != nil {
					return echo.NewHTTPError(http.StatusUnauthorized, "无效 Access 和 Refresh Token，请重新登录")
				}

				c.Response().Header().Set(DefaultJWTConfig.Authorization, DefaultJWTConfig.TokenPrefix+newTokens["accessToken"])
				c.Response().Header().Set(DefaultJWTConfig.RefreshToken, DefaultJWTConfig.TokenPrefix+newTokens["refreshToken"])
				return next(c)
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				userId := int64(claims[DefaultJWTConfig.LocalsUserIdKey].(float64))
				c.Set(account.LocalsUserIdKey, userId)

				redisKey := DefaultJWTConfig.RedisPrefix + strconv.FormatInt(userId, 10)
				exp := claims["exp"].(float64)
				expireTime := time.Until(time.Unix(int64(exp), 0))

				err := global.RedisClient.Set(context.Background(), redisKey, tokenString, expireTime).Err()
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "无法存储 Token 到 Redis")
				}
			} else {
				return echo.NewHTTPError(http.StatusUnauthorized, "Access Token 无效，请重新登录")
			}

			return next(c)
		}
	}
}

// refreshTokenLogic 负责刷新 Token
func refreshTokenLogic(refreshTokenString string) (map[string]string, error) {
	token, err := utils.ValidateJWTToken(refreshTokenString, true)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId := uint(claims[DefaultJWTConfig.LocalsUserIdKey].(float64))

		newAccessToken, newRefreshToken, err := utils.GenerateJWT(userId)
		if err != nil {
			return nil, err
		}

		return map[string]string{
			"accessToken":  newAccessToken,
			"refreshToken": newRefreshToken,
		}, nil
	}

	return nil, fmt.Errorf("token 验证失败")
}
