package net

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"github.com/vocdoni/go-dvote/net/epoll"
	"github.com/vocdoni/go-dvote/types"
	"golang.org/x/crypto/acme/autocert"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type WebsocketHandle struct {
	c *types.Connection
	e *epoll.Epoll
}

func (w *WebsocketHandle) upgrader(writer http.ResponseWriter, reader *http.Request) {

	fail := func(err error) {
		writer.Header().Set("Content-Type", "text/plain")
		log.Panicf("%s", err)
		writer.Write([]byte(err.Error()))
	}

	req, err := http.NewRequest("POST", "0.0.0.0:9090", reader.Body)
	if err != nil {
		fail(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fail(err)
		return
	}

	defer resp.Body.Close()

	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	if _, err := io.Copy(writer, resp.Body); err != nil {
		fail(err)
		return
	}

	// Upgrade connection
	conn, _, _, err := ws.UpgradeHTTP(reader, writer)
	if err != nil {
		return
	}
	if err := w.e.Add(conn); err != nil {
		log.Printf("Failed to add connection %v", err)
		conn.Close()
	}
}

func (w *WebsocketHandle) Init(c *types.Connection) error {
	// Increase resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return err
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return err
	}

	// Start epoll
	var err error
	w.e, err = epoll.MkEpoll()
	if err != nil {
		return err
	}

	http.HandleFunc(c.Path, w.upgrader)

	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("127.0.0.1"),
		Cache:      autocert.DirCache("cache-path"),
	}

	server := &http.Server{
		Addr: ":9090",
		TLSConfig: &tls.Config{
			GetCertificate: m.GetCertificate,
		},
	}

	go func() {
		log.Fatal(server.ListenAndServeTLS("", ""))
	}()

	addr := string("wss://" + c.Address + ":" + strconv.Itoa(c.Port))
	log.Printf("Dvote websocket endpoint initialized on %s", addr)

	return nil
}

func (w *WebsocketHandle) Listen(reciever chan<- types.Message) {
	var msg types.Message
	for {
		connections, err := w.e.Wait()
		if err != nil {
			log.Printf("WS recieve error: %s", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			if payload, _, err := wsutil.ReadClientData(conn); err != nil {
				if err := w.e.Remove(conn); err != nil {
					log.Printf("WS recieve error: %s", err)
				}
				conn.Close()
			} else {
				msg.Data = []byte(payload)
				msg.TimeStamp = time.Now()
				ctx := new(types.WebsocketContext)
				ctx.Conn = &conn
				msg.Context = ctx
				reciever <- msg
			}
		}
	}
}

func (w *WebsocketHandle) Send(msg types.Message) {
	wsutil.WriteServerMessage(*msg.Context.(*types.WebsocketContext).Conn, ws.OpBinary, msg.Data)
}
