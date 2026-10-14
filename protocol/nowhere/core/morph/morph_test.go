package morph

import (
	"bytes"
	"encoding/hex"
	"io"
	"net"
	"testing"
)

func TestDeriveMatchesRustVector(t *testing.T) {
	keys := Derive([]byte("test portal key"))
	assertHex(t, keys.TCPC2S[:], "f3d2fabc455a59c400d09d400b5a0b9b221e2b157589fc87674b902111ea1014")
	assertHex(t, keys.TCPS2C[:], "190cfc4c689b24d4bfc8431f46d9f2d1bcbf6222eed612a3ea12d04ff4e4af8c")
	assertHex(t, keys.UDPC2S[:], "20221bde5d70b44211a22c7e62967945d7d85bcde0d17a3b28d1fa686f8e5c1f")
	assertHex(t, keys.UDPS2C[:], "78938ee9245028e7de63ac27b0ab64c68df2961a2cf39da1394f0f242bf8b389")
}

func TestTCPRoundTrip(t *testing.T) {
	clientSide, serverSide := net.Pipe()
	keys := Derive([]byte("0123456789abcdef0123456789abcdef"))
	client, err := WrapTCPClient(clientSide, keys)
	if err != nil {
		t.Fatal(err)
	}
	server := WrapTCPServer(serverSide, keys)
	payload := []byte("tls-record-bytes")
	errCh := make(chan error, 1)
	go func() {
		buf := make([]byte, len(payload))
		_, err := io.ReadFull(server, buf)
		if err != nil {
			errCh <- err
			return
		}
		if !bytes.Equal(buf, payload) {
			errCh <- io.ErrUnexpectedEOF
			return
		}
		_, err = server.Write(payload)
		errCh <- err
	}()
	if _, err := client.Write(payload); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, len(payload))
	if _, err := io.ReadFull(client, buf); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, payload) {
		t.Fatalf("client read %q", buf)
	}
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
}

func assertHex(t *testing.T, got []byte, want string) {
	t.Helper()
	if hex.EncodeToString(got) != want {
		t.Fatalf("got %x want %s", got, want)
	}
}
