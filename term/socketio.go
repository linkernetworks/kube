package term

import (
	"bytes"

	socketio "github.com/c9s/go-socket.io"
)

type SocketIoReader struct {
	Event  string
	Socket socketio.Socket
	Buffer chan []byte
}

func (r *SocketIoReader) Write(p []byte) (n int, err error) {
	r.Buffer <- p
	return len(p), nil
}

func (r *SocketIoReader) Read(p []byte) (n int, err error) {
	data := <-r.Buffer
	n = copy(p, data)
	return n, nil
}

type SocketIoWriter struct {
	Event  string
	Socket socketio.Socket
	Buffer *bytes.Buffer
}

func (w *SocketIoWriter) Write(p []byte) (n int, err error) {
	n, err = w.Buffer.Write(p)
	if werr := w.Socket.Emit(w.Event, w.Buffer.String()); werr == nil {
		w.Buffer.Reset()
	}
	return n, err
}
