package exector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type ExectorHttp struct {
	Url string ""
	Method string ""
	Data map[string]interface {}
	ContentType string ""
}

func NewExectorHttp(url string, method string, data map[string]interface {}) *ExectorHttp {
	return &ExectorHttp{
		Url: url,
		Method: method,
		Data: data,
		ContentType: "application/json",
	}
}
func (e *ExectorHttp) Execute() error {
	if e.Method == "post" {
		Post(e.Url, e.Data, e.ContentType)
	} else if e.Method == "get" {
		Get(e.Url)
	}
	return nil
}

func Get(url string) {
	// 超时时间：5秒
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}

	fmt.Println(result.String())
}

func Post(url string, data map[string]interface {}, contentType string)  {
	// 超时时间：5秒
	client := &http.Client{Timeout: 5 * time.Second}
	jsonStr, _ := json.Marshal(data)
	fmt.Println(jsonStr)
	resp, err := client.Post(url, contentType, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(result))
}