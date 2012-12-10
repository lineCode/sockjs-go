package sockjs

import (
	"bytes"
	"code.google.com/p/gorilla/mux"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type xhrStreamingProtocol struct{}

func (this *context) XhrStreamingHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	sessid := vars["sessionid"]

	httpTx := &httpTransaction{
		protocolHelper: xhrStreamingProtocol{},
		req:            req,
		rw:             rw,
		sessionId:      sessid,
		done:           make(chan bool),
	}
	this.baseHandler(httpTx)
}

func (xhrStreamingProtocol) isStreaming() bool   { return true }
func (xhrStreamingProtocol) contentType() string { return "application/javascript; charset=UTF-8" }

func (xhrStreamingProtocol) writeOpenFrame(w io.Writer) (int, error) {
	return fmt.Fprintln(w, "o")
}
func (xhrStreamingProtocol) writeHeartbeat(w io.Writer) (int, error) {
	return fmt.Fprintln(w, "h")
}
func (xhrStreamingProtocol) writePrelude(w io.Writer) (int, error) {
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "%s\n", strings.Repeat("h", 2048))
	n, err := b.WriteTo(w)
	return int(n), err
}
func (xhrStreamingProtocol) writeClose(w io.Writer, code int, msg string) (int, error) {
	return fmt.Fprintf(w, "c[%d,\"%s\"]\n", code, msg)
}

// ****** following code was taken from https://github.com/mrlauer/gosockjs
var re = regexp.MustCompile("[\x00-\x1f\u200c-\u200f\u2028-\u202f\u2060-\u206f\ufff0-\uffff]")

// ****** end

func (xhrStreamingProtocol) writeData(w io.Writer, frames ...[]byte) (int, error) {
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "a[")
	for n, frame := range frames {
		if n > 0 {
			b.Write([]byte(","))
		}
		sesc := re.ReplaceAllFunc(frame, func(s []byte) []byte {
			return []byte(fmt.Sprintf(`\u%04x`, []rune(string(s))[0]))
		})
		b.Write(sesc)
	}
	fmt.Fprintf(b, "]\n")
	n, err := b.WriteTo(w)
	return int(n), err
}