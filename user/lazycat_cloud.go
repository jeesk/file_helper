//go:build lazycat_cloud
// +build lazycat_cloud

package tools

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
)

const BoxHomeRootDir = "/lzcapp/run/mnt/home/"

func MustGetUID(c *http.Request) string {
	uid, ok := GetUID(c)
	if !ok {
		panic("uid not found")
	}
	return uid
}

func GetUID(c *http.Request) (string, bool) {
	s := c.Header.Get("X-Hc-User-Id")
	return s, s != ""
}

// 假设UID必须存在
func GetUserHome(c *http.Request) string {
	uid := c.Header.Get("X-Hc-User-Id")
	return filepath.Join(BoxHomeRootDir, uid)
}
func GetUser(c *gin.Context) string {
	return c.GetHeader("X-Hc-User-Id")
}
func PrintUserVersion() {
	fmt.Println("用户auth版本信息: lazycat_cloud")
}
