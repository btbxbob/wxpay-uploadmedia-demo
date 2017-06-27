package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

var cert tls.Certificate
var caCert []byte

func init() {
	var err error
	// Load client cert
	cert, err = tls.LoadX509KeyPair("", "")
	if err != nil {
		log.Fatal(err)
	}

	// Load CA cert
	caCert, err = ioutil.ReadFile("")
	if err != nil {
		log.Fatal(err)
	}

}

//UploadMediaRequest 上传图片数据结构
type UploadMediaRequest struct {
	//XMLName xml根节点标识
	XMLName xml.Name `xml:"xml"`
	//mchID 商户号
	MchID string `xml:"mch_id,cdata"`
	//media 媒体文件
	Media string `xml:"media,cdata"`
	//mediaHash 媒体文件hash
	MediaHash string `xml:"media_hash,cdata"`
	//sign 签名
	Sign string `xml:"sign,cdata"`
}

func main() {

	//multipart post form

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add your image file
	file = "test.jpg"
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()
	fw, err := w.CreateFormFile("image", file)
	if err != nil {
		return
	}
	if _, err = io.Copy(fw, f); err != nil {
		return
	}
	// Add the other fields
	if fw, err = w.CreateFormField("key"); err != nil {
		return
	}
	if _, err = fw.Write([]byte("KEY")); err != nil {
		return
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	newReq, err := http.NewRequest("POST", "123", strings.NewReader("name=cjb"))
	//newReq.Header = req.Header
	newReq.URL.Host = "api.mch.weixin.qq.com"
	newReq.Host = newReq.URL.Host
	newReq.URL.Scheme = "https"

	var client *http.Client
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client = &http.Client{Transport: transport}

	resp, err := client.Do(newReq)

	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println(resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	log.Printf("[INFO][IN] Respond content:%v", string(respBody))
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err.Error())
	}

	return
}
