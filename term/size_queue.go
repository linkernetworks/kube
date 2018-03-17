package term

import "k8s.io/client-go/tools/remotecommand"

type SocketIoSizeQueue struct {
	C chan *remotecommand.TerminalSize
}

func (q *SocketIoSizeQueue) Push(cols uint16, rows uint16) {
	q.C <- &remotecommand.TerminalSize{Width: cols, Height: rows}
}

func (q *SocketIoSizeQueue) Next() *remotecommand.TerminalSize {
	return <-q.C
}
