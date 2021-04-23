package mux

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestHTTP1(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	a := require.New(t)
	path := "/test/"
	method := "GET"
	mockResData := []byte("948 res")

	rootListener := newLocalListener()
	mux := New(rootListener)
	httpListener := mux.Match(HTTP1())
	mux.OnError(func(err error) {
		a.Fail("connection not matched by a http matcher")
	})

	muxServeCh := make(chan error, 1)
	go func() {
		muxServeCh <- mux.Serve()
	}()

	handler := func(writer http.ResponseWriter, request *http.Request) {
		var err error
		a.Equal(method, request.Method)
		n, err := writer.Write(mockResData)
		a.NoError(err)
		a.Len(mockResData, n)
	}
	httpMux := http.NewServeMux()
	httpMux.HandleFunc(path, handler)
	httpServer := httptest.NewUnstartedServer(httpMux)
	httpServer.Listener = httpListener
	httpServer.Start()

	url := httpServer.URL + path
	httpClient := httpServer.Client()
	req, _ := http.NewRequest(method, url, nil)
	res, err := httpClient.Do(req)
	a.NoError(err)
	a.Equal(200, res.StatusCode)
	responseData, err := ioutil.ReadAll(res.Body)
	a.NoError(err)
	a.Equal(mockResData, responseData)
	a.NoError(res.Body.Close())
	httpClient.CloseIdleConnections()

	httpServer.Close()
	a.NoError(mux.Close())
	muxServeErr := <-muxServeCh
	a.NoError(muxServeErr)
}

func TestAny(t *testing.T) {
	a := require.New(t)
	mockReqData := []byte("1 somre")
	requestData := make([]byte, len(mockReqData))

	rootListener := newLocalListener()
	mux := New(rootListener)
	httpListener := mux.Match(HTTP1())
	tcpListener := mux.Match(Any())
	go func() {
		err := mux.Serve()
		log.Println(err)
	}()

	handler := func(writer http.ResponseWriter, request *http.Request) {
		a.Fail("matched http handler")
	}
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", handler)
	httpServer := httptest.NewUnstartedServer(httpMux)
	httpServer.Listener = httpListener
	httpServer.Start()

	hadTCPConn := false
	go func() {
		for {
			conn, err := tcpListener.Accept()
			if hadTCPConn {
				a.Fail("more than one tcp conn")
				return
			}
			hadTCPConn = true
			a.NoError(err)
			_, err = conn.Read(requestData)
			a.NoError(err)
			a.Equal(mockReqData, requestData)
		}
	}()

	tcpAddr := tcpListener.Addr().String()
	conn, err := net.Dial("tcp", tcpAddr)
	a.NoError(err)
	n, err := conn.Write(mockReqData)
	a.Len(mockReqData, n)
	a.NoError(err)

	time.Sleep(10 * time.Millisecond)
	a.Equal(mockReqData, requestData)
	a.True(hadTCPConn)
}

func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("failed to listen on a port: %v", err))
	}
	return l
}
