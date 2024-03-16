package main

import (
	"embed"
	"file-helper/db"
	user "file-helper/user"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lxzan/gws"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func startServer(mb db.MsgBackend, staticDic string) {
	fmt.Println("代码版本:" + Version)
	if !strings.HasSuffix(staticDic, "/") {
		staticDic = staticDic + "/"
	}
	if checkDir(staticDic) {
		fmt.Println("该文件夹不存在，并且创建失败")
		return
	}

	db.InitMsgBackend(mb)
	user.PrintUserVersion()

	r := mux.NewRouter()
	hanlder := FpHanlder{&Handler{sessions: gws.NewConcurrentMap[string, *gws.Conn](16)}}
	r.PathPrefix("/fp/ws").Handler(hanlder)
	r.PathPrefix("/fp/all").Handler(hanlder)
	r.PathPrefix("/fp/upload").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 解析表单
		err := r.ParseMultipartForm(50 << 20) // 设置最大内存限制为10MB
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 获取文件句柄、头信息和大小
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		// 创建目标文件
		uuidStr := uuid.New().String()
		index := strings.LastIndex(handler.Filename, ".")
		fileExt := handler.Filename[index:]
		fileName := uuidStr + fileExt
		dst, err := os.Create(staticDic + fileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		// 复制文件内容到目标文件
		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, fmt.Sprintf("{\"filename\": \"%s\"}", fileName))
	})
	r.PathPrefix("/fp/static").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileName := strings.Replace(r.URL.Path, "/fp/static/", "", 1)
		filePath := filepath.Join(staticDic, fileName)
		_, err := http.Dir(staticDic).Open(fileName)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filePath)
	})
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

func checkDir(staticDic string) bool {
	if _, err := os.Stat(staticDic); os.IsNotExist(err) {
		err := os.MkdirAll(staticDic, os.ModePerm)
		if err != nil {
			fmt.Println("无法创建目录:", err)
			return true
		}
		fmt.Println("目录创建成功")
	} else {
		fmt.Println("目录已存在")
	}
	return false
}
