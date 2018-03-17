package term

import (
	"io"
	"log"

	socketio "github.com/c9s/go-socket.io"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/remotecommand"
)

type SocketIoReader struct {
	Event  string
	Socket socketio.Socket
	Buffer chan []byte
}

func (r *SocketIoReader) Write(p []byte) (n int, err error) {
	log.Println(r.Event, p)
	r.Buffer <- p
	return len(p), nil
}

func (r *SocketIoReader) Read(p []byte) (n int, err error) {
	data := <-r.Buffer
	n = copy(p, data)
	return n, err
}

type SocketIoWriter struct {
	Event  string
	Socket socketio.Socket
}

func (w *SocketIoWriter) Write(p []byte) (n int, err error) {
	data := string(p)
	if err := w.Socket.Emit(w.Event, data); err != nil {
		if err != io.EOF {
			log.Println("emit error:", err)
		}
		return 0, err
	}
	log.Printf("emit %s -> '%v'", w.Event, p)
	return len(p), err
}

type SocketIoTermSession struct {
	Stdout    *SocketIoWriter
	Stderr    *SocketIoWriter
	Stdin     *SocketIoReader
	SizeQueue *SocketIoSizeQueue
	TTY       bool
	Pod       *corev1.Pod
}

func NewSession(socket socketio.Socket, pod *corev1.Pod) *SocketIoTermSession {
	return &SocketIoTermSession{
		Stdin:     &SocketIoReader{Event: "term:stdin", Socket: socket, Buffer: make(chan []byte, 30)},
		Stdout:    &SocketIoWriter{Event: "term:stdout", Socket: socket},
		Stderr:    &SocketIoWriter{Event: "term:stderr", Socket: socket},
		SizeQueue: &SocketIoSizeQueue{C: make(chan *remotecommand.TerminalSize, 10)},
		TTY:       true,
		Pod:       pod,
	}
}

func (t *SocketIoTermSession) Terminate() {
	// ^U (21) and ^D (4)
	t.Stdin.Write([]byte{21, 4})
}
