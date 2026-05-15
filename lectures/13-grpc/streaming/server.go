// Package streaming — эскизы обработчиков всех 4 типов RPC.
// Конкретные сигнатуры берутся из сгенерированного *_grpc.pb.go.
package streaming

import "io"

type Msg struct{ Text string }

type RepeatServer interface {
	Send(*Msg) error
}

type CollectServer interface {
	Recv() (*Msg, error)
	SendAndClose(*Msg) error
}

type ChatServer interface {
	Recv() (*Msg, error)
	Send(*Msg) error
}

type Server struct{}

// Unary: один запрос — один ответ.
func (Server) Say(req *Msg) (*Msg, error) {
	return &Msg{Text: req.Text}, nil
}

// Server streaming: один Recv (req пришёл аргументом), много Send.
func (Server) Repeat(req *Msg, stream RepeatServer) error {
	for i := 0; i < 3; i++ {
		if err := stream.Send(&Msg{Text: req.Text}); err != nil {
			return err
		}
	}
	return nil
}

// Client streaming: много Recv до io.EOF, один SendAndClose.
func (Server) Collect(stream CollectServer) error {
	var joined string
	for {
		m, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&Msg{Text: joined})
		}
		if err != nil {
			return err
		}
		joined += m.Text
	}
}

// Bidi: Recv и Send независимы, обычно в разных горутинах.
func (Server) Chat(stream ChatServer) error {
	for {
		m, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&Msg{Text: "echo: " + m.Text}); err != nil {
			return err
		}
	}
}
