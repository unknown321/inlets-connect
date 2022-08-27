package handler

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	httpTimeout = time.Second * 5
)

type Bucket struct {
	Quota         int64
	Value         int64
	LastAccess    time.Time
	LimitDuration time.Duration
}

type Buckets map[string]Bucket

func (b *Bucket) ResetQuota() {
	b.Value = 0
}

var buckets = Buckets{"google.com:443": {
	Quota:         2 * 1024 * 1024,
	Value:         0,
	LastAccess:    time.Time{},
	LimitDuration: time.Second * 10,
}}

func pipe(from net.Conn, to net.Conn) (b int64, err error) {
	defer from.Close()
	n, err := io.Copy(from, to)
	log.Printf("Wrote: %d bytes", n)

	if err != nil && strings.Contains(err.Error(), "closed network") {
		return n, nil
	}

	if err != nil {
		return 0, fmt.Errorf("cannot pipe: %w", err)
	}

	return n, nil
}

//nolint:funlen
func Handle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodConnect {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

			return
		}

		if v, ok := buckets[r.Host]; ok {
			if time.Now().Add(v.LimitDuration * -1).After(v.LastAccess) {
				log.Printf("Quota reset for %s", r.Host)
				v.ResetQuota()
			}

			if v.Value > v.Quota {
				http.Error(w, "Quota reached", http.StatusPaymentRequired)

				return
			}
		}

		defer r.Body.Close()

		conn, err := net.DialTimeout("tcp", r.Host, httpTimeout)
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
			return pipeWithBucket(reqConn, conn, r.Host)
		})
		g.Go(func() error {
			_, errReq := pipe(conn, reqConn)

			return errReq
		})

		if err = g.Wait(); err != nil {
			log.Println("Error", err.Error())
		}

		log.Printf("Connection %s done.", conn.RemoteAddr())
	})
}

func pipeWithBucket(reqConn net.Conn, conn net.Conn, host string) error {
	b, errResp := pipe(reqConn, conn)
	if b != 0 {
		if v, ok := buckets[host]; ok {
			v.Value += b
			v.LastAccess = time.Now()
			buckets[host] = v

			log.Printf("host %s, value %d, quota %d", host, buckets[host].Value, buckets[host].Quota)
		}
	}

	return errResp
}
