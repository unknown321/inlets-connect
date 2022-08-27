package handler

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func pipe(from net.Conn, to net.Conn) error {
	defer from.Close()
	n, err := io.Copy(from, to)
	log.Printf("Wrote: %d bytes", n)
	if err != nil && strings.Contains(err.Error(), "closed network") {
		return nil
	}
	return err
}

func Handle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodConnect {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Printf("%s", r.Host)

		defer r.Body.Close()

		conn, err := net.DialTimeout("tcp", r.Host, time.Second*5)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to dial %s, error: %s", r.Host, err.Error()), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)

		log.Printf("Dialed upstream: %s %s", conn.RemoteAddr(), conn.LocalAddr())

		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Unable to hijack connection", http.StatusInternalServerError)
			return
		}

		reqConn, wbuf, err := hj.Hijack()
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to hijack connection %s", err), http.StatusInternalServerError)
			return
		}
		defer reqConn.Close()
		defer wbuf.Flush()

		g := new(errgroup.Group)
		g.Go(func() error {
			return pipe(reqConn, conn)
		})
		g.Go(func() error {
			return pipe(conn, reqConn)
		})

		if err := g.Wait(); err != nil {
			log.Println("Error", err.Error())
		}

		log.Printf("Connection %s done.", conn.RemoteAddr())
	})
}
