package main

import (
	"embed"
	"file-helper/db"
	user "file-helper/user"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lxzan/gws"
	"io/fs"
	"net"
	"net/http"
)

var Version = ""

func createUIHanlder() http.Handler {
	fileFs, err := fs.Sub(ui, "ui/dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(fileFs))
	return fileServer
}

//go:embed ui/dist
var ui embed.FS

const port = 7855

func redirectToUIPath(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

const preferRootPath = true

// 显式声明实现 MsgBackend 接口

func startServer(mb db.MsgBackend) {
	fmt.Println("代码版本:" + Version)
	db.InitMsgBackend(mb)
	user.PrintUserVersion()

	r := mux.NewRouter()
	hanlder := FpHanlder{&Handler{sessions: gws.NewConcurrentMap[string, *gws.Conn](16)}}
	r.PathPrefix("/fp/ws").Handler(hanlder)
	r.PathPrefix("/fp/all").Handler(hanlder)
	uiHanlder := createUIHanlder()
	if preferRootPath {
		r.PathPrefix("/").Handler(uiHanlder)
	} else {
		r.PathPrefix("/ui/").Handler(http.StripPrefix("/ui", uiHanlder))
		r.PathPrefix("/ui").HandlerFunc(redirectToUIPath)
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	go func() {
		hs := http.Server{
			Handler: r,
		}
		err = hs.Serve(lis)
		if err != nil {
			fmt.Println("服务启动失败", err.Error())
		} else {
			fmt.Println("服务启动成功")
		}
	}()
}
