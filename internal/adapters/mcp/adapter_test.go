package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/royisme/bobamixer/internal/adapters"
)

type fakeTransport struct {
	payload []byte
	resp    responseEnvelope
}

func (f *fakeTransport) Call(ctx context.Context, payload []byte) ([]byte, error) {
	f.payload = payload
	return json.Marshal(f.resp)
}

func TestAdapterExecute(t *testing.T) {
	ft := &fakeTransport{resp: responseEnvelope{Output: "ok", InputTokens: 10, OutputTokens: 5}}
	adapter := New(ft, "default")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := adapter.Execute(ctx, adapters.Request{SessionID: "s", Profile: "p", Tool: "tool"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if string(res.Output) != "ok" {
		t.Fatalf("output=%s", res.Output)
	}
	if ft.payload == nil {
		t.Fatal("payload not sent")
	}
}
