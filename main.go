package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var mux map[string]func(http.ResponseWriter, *http.Request)

type Handler struct{}

const (
	TemplateDir = "./view/"
	UploadDir   = "./upload/"
)

func (*Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := mux[r.URL.String()]; ok {
		h(w, r)
		return
	}
	if ok, _ := regexp.MatchString("/css/", r.URL.String()); ok {
		http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))).ServeHTTP(w, r)
	} else {
		http.StripPrefix("/", http.FileServer(http.Dir("./upload/"))).ServeHTTP(w, r)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t, _ := template.ParseFiles(TemplateDir + "file.html")
		t.Execute(w, "上传文件")
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Fprintf(w, "%v", "上传错误")
			return
		}
		fileext := filepath.Ext(handler.Filename)
		if check(fileext) == false {
			fmt.Fprintf(w, "%v", "不允许的上传类型")
			return
		}

		filename := handler.Filename
		f, _ := os.OpenFile(UploadDir+handler.Filename, os.O_CREATE|os.O_WRONLY, 0660)
		_, err = io.Copy(f, file)
		if err != nil {
			fmt.Fprintf(w, "%v", "上传失败")
			return
		}
		fileDir, _ := filepath.Abs(UploadDir + filename)
		fmt.Fprintf(w, "%v", filename+"上传完成,服务器地址:"+fileDir)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles(TemplateDir + "index.html")
	t.Execute(w, "首页")
}

func StaticServer(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/file", http.FileServer(http.Dir("./upload/"))).ServeHTTP(w, r)
}

func check(name string) bool {
	ext := []string{".exe", ".js", ".png"}

	for _, v := range ext {
		if v == name {
			return false
		}
	}
	return true
}

func main() {
	server := http.Server{
		Addr:        ":9090",
		Handler:     &Handler{},
		ReadTimeout: 10 * time.Second,
	}
	mux = make(map[string]func(http.ResponseWriter, *http.Request))
	mux["/"] = index
	mux["/upload"] = upload
	mux["/file"] = StaticServer
	fmt.Println("启动成功！")
	server.ListenAndServe()
}
