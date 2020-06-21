package cobraprompt

import "github.com/c-bata/go-prompt"

// RawWriter is an instance of prompt.ConsoleWriter interface
// where Write() and WriteStr() does not remove control sequences.
// This allows more colorful prompt prefix and suggestion items.
type RawWriter struct {
	prompt.PosixWriter
}

func (w *RawWriter) Write(data []byte) {
	w.WriteRaw(data)
}

func (w *RawWriter) WriteStr(data string) {
	w.WriteRawStr(data)
}
