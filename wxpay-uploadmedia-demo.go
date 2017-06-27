// A demo of Weixin Pay upload media api.
package main

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

var config Configuration
var cert tls.Certificate
var caCert []byte

// UploadMediaRequest 上传图片数据结构
type UploadMediaRequest struct {
	// XMLName xml根节点标识
	XMLName xml.Name `xml:"xml"`
	// MchID 商户号
	MchID string `xml:"mch_id"`
	// Media 媒体文件
	Media string `xml:"media"`
	// MediaHash 媒体文件hash
	MediaHash string `xml:"media_hash"`
	// Sign 签名
	Sign string `xml:"sign"`
}

// Configuration 配置类
type Configuration struct {
	// AppID appid
	// AppID string `json:"appid"`
	// MchID 商户号
	MchID string `json:"mch_id"`
	// Key 秘钥，在商户平台的API安全里设置
	Key string
	// Cert 证书配置
	Cert struct {
		// CertFile 证书文件路径
		CertFile string
		// KeyFile 秘钥文件路径
		KeyFile string
		// Ca CA证书路径
		Ca string
	}
	// ImgFile 图片文件名
	ImgFile string `json:"img_file"`
	// IsLoad 判断是否已经载入配置
	IsLoad bool
}

func init() {
	var err error

	config.IsLoad = false
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal("error:", err)
	}
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)

	err = decoder.Decode(&config)

	if err != nil {
		log.Println("error:", err)
	}

	config.IsLoad = true

	// Load client cert
	cert, err = tls.LoadX509KeyPair(config.Cert.CertFile, config.Cert.KeyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Load CA cert
	caCert, err = ioutil.ReadFile(config.Cert.Ca)
	if err != nil {
		log.Fatal(err)
	}

}

func main() {
	log.Println("===start===")
	var uploadMediaRequest UploadMediaRequest
	uploadMediaRequest.MchID = config.MchID
	//multipart post form

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add your image file
	file := config.ImgFile
	f, err := os.Open(file)
	if err != nil {
		log.Fatal("open img file error:", err)
	}

	// calculate hash first
	md5h := md5.New()
	io.Copy(md5h, f)
	f.Seek(0, os.SEEK_SET)

	defer f.Close()

	//file part
	fw, err := w.CreateFormFile("media", file)
	if err != nil {
		log.Fatal(err)
		return
	}
	if _, err = io.Copy(fw, f); err != nil {
		log.Fatal(err)
		return
	}
	uploadMediaRequest.MediaHash = strings.ToUpper(hex.EncodeToString(md5h.Sum(nil)))
	// calculate sign
	var hashString = "mch_id=" + uploadMediaRequest.MchID + "&media_hash=" + uploadMediaRequest.MediaHash + "&key=" + config.Key
	hasher := md5.New()
	hasher.Write([]byte(hashString))
	uploadMediaRequest.Sign = strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))

	output, err := xml.MarshalIndent(uploadMediaRequest, "  ", "    ")
	if err != nil {
		log.Printf("error: %v\n", err)
	}

	log.Println("Request in XML:\n", string(output))

	// Add the other fields
	// mch_id
	if fw, err = w.CreateFormField("mch_id"); err != nil {
		return
	}
	if _, err = fw.Write([]byte(uploadMediaRequest.MchID)); err != nil {
		return
	}

	// media_hash
	if fw, err = w.CreateFormField("media_hash"); err != nil {
		return
	}
	if _, err = fw.Write([]byte(uploadMediaRequest.MediaHash)); err != nil {
		return
	}

	// sign
	if fw, err = w.CreateFormField("sign"); err != nil {
		return
	}
	if _, err = fw.Write([]byte(uploadMediaRequest.Sign)); err != nil {
		return
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	b.WriteTo(os.Stdout)

	newReq, err := http.NewRequest("POST", "https://api.mch.weixin.qq.com/secapi/mch/uploadmedia", &b)
	//newReq.Header = req.Header
	newReq.URL.Host = "api.mch.weixin.qq.com"
	newReq.Host = newReq.URL.Host
	newReq.URL.Scheme = "https"
	newReq.Header.Set("Content-Type", w.FormDataContentType())

	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequestOut(newReq, true)
	if err != nil {
		fmt.Println(err)
	}
	log.Println(string(requestDump))

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
