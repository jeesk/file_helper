//go:build custom
// +build custom

package tools

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"path/filepath"
)

const BoxHomeRootDir = "/lzcapp/run/mnt/home/"

func MustGetUID(c *gin.Context) string {
	uid, ok := GetUID(c)
	if !ok {
		panic("uid not found")
	}
	return uid
}

func GetUID(c *gin.Context) (string, bool) {
	s := c.GetHeader("X-Hc-User-Id")
	return s, s != ""
}

// 假设UID必须存在
func GetUserHome(c *gin.Context) string {
	uid := c.GetHeader("X-Hc-User-Id")
	return filepath.Join(BoxHomeRootDir, uid)
}
func GetUser(c *gin.Context) string {
	return c.GetHeader("X-Hc-User-Id")
}

func PrintUserAuthVersion() {
	fmt.Println("用户auth版本信息: custom")
}
