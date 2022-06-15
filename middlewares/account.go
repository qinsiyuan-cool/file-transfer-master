package middlewares

import (
	"FileTransfer/pkg/setting"
	"github.com/patrickmn/go-cache"
	"time"
)

var (
	Cache        *cache.Cache
	RefreshToken string
	AccessToken  string
	Status       string
)

func init() {
	Status = "work"
	Cache = cache.New(time.Duration(60)*time.Minute, time.Duration(120)*time.Minute)
}
func SetRefreshToken(refreshToken string) {
	Cache.Set("refresh_token", refreshToken, cache.DefaultExpiration)
}
func SetAccessToken(accessToken string) {
	Cache.Set("access_token", accessToken, cache.DefaultExpiration)
}
func GetRefreshToken() string {
	re, ok := Cache.Get("refresh_token")
	if ok {
		return re.(string)
	}
	return setting.RefreshToken

}
func GetAccessToken() string {
	re, ok := Cache.Get("access_token")
	if ok {
		return re.(string)
	}
	return setting.RefreshToken

}
