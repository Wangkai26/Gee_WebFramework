package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"gee"
)

type student struct {
	Name string
	Age  int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	r := gee.Default()

	// 验证 实现静态资源服务，支持HTML渲染
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets","./static")

	stu1 := &student{"zhangsan",18}
	stu2 := &student{"lisi",19}
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK,"css.tmpl",nil)
	})
	r.GET("/students", func(c *gee.Context) {
		c.HTML(http.StatusOK,"arr.tmpl",gee.H{
			"title":"kai",
			"stuArr":[2]*student{stu1,stu2},
		})
	})

	r.GET("/date", func(c *gee.Context) {
		c.HTML(http.StatusOK,"custom_func.tmpl",gee.H{
			"title":"kai",
			"now":time.Date(2022,10,2,22,9,0,0,time.UTC),
		})
	})

	// 验证错误处理机制
	//r.GET("/", func(c *gee.Context) {
	//	c.String(http.StatusOK, "Hello Geektutu\n")
	//})
	// index out of range for testing Recovery()
	r.GET("/panic", func(c *gee.Context) {
		names := []string{"geektutu"}
		c.String(http.StatusOK, names[100])
	})

	r.Run(":9999")
}
