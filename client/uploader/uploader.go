package uploader

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/m1m0ry/golang/ftp/client/common"
)

func UploadFile(filePath string) error {
	targetUrl := common.BaseUrl + "upload"
	if !common.IsFile(filePath) {
		fmt.Printf("filePath:%s is not exist", filePath)
		return errors.New(filePath + "文件不存在")
	}
	filename := filepath.Base(filePath)
	//先定义http体
	bodyBuf := &bytes.Buffer{}//http body
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("filename", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}
	//打开文件句柄操(文件句柄对于打开的文件是唯一的识别依据)
	fh, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("error opening filePath: %s\n", filePath)
		return err
	}
	hasher := &common.Hasher{//指针
		Reader: fh,    //文件名
		Hash:   sha1.New(),
		Size:   0,
	}
	_, err = io.Copy(fileWriter, hasher)//写http body
	if err != nil {
		return err
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	//创建http请求
	request, err := http.NewRequest(http.MethodPost, targetUrl, bodyBuf)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", contentType)
	request.Header.Add("file-md5", hasher.Sum())
	fmt.Println(hasher.Sum())
	
	resp, err := http.DefaultClient.Do(request) //发送并接收响应
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%s文件上传失败\n", filename)
		return errors.New("上传文件失败")
	}
	fmt.Printf("上传文件%s成功\n", filename)
	return nil
}
