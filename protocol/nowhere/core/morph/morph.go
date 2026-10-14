// Package morph is the Nowhere 3 keyed transform below TLS and QUIC.
// Key derivation uses HMAC with the root label "nowhere/nw3/morph". It does
// not call the Nowhere 2 Morph KDF.
package morph

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/sagernet/sing-box/protocol/nowhere/core/internal/chacha20"
)

const (
	nonceLen      = 12
	preludeLen    = 64
	streamLimit   = (1 << 38) - 64
	preludeEnv    = "NOW_MORPH_TCP_PRELUDE"
	morphRootSalt = "nowhere/nw3/morph"
)

// Keys are the directional Morph outputs for one shared key.
type Keys struct {
	TCPC2S [32]byte
	TCPS2C [32]byte
	UDPC2S [32]byte
	UDPS2C [32]byte
}

// Derive builds Morph keys. sharedKey is the same raw text used for authentication.
func Derive(sharedKey []byte) Keys {
	root := hmacSHA256([]byte(morphRootSalt), sharedKey)
	return Keys{
		TCPC2S: expand(root, "tcp c2s"),
		TCPS2C: expand(root, "tcp s2c"),
		UDPC2S: expand(root, "udp c2s"),
		UDPS2C: expand(root, "udp s2c"),
	}
}

func hmacSHA256(key, data []byte) [32]byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	var out [32]byte
	copy(out[:], mac.Sum(nil))
	return out
}

func expand(root [32]byte, info string) [32]byte {
	input := make([]byte, len(info)+1)
	copy(input, info)
	input[len(info)] = 0x01
	return hmacSHA256(root[:], input)
}

type preludePolicy uint8

const (
	preludeLow7 preludePolicy = iota
	preludeFull8
)

func preludePolicyFromEnv() (preludePolicy, error) {
	value, ok := os.LookupEnv(preludeEnv)
	if !ok {
		return preludeFull8, nil
	}
	switch value {
	case "low7":
		return preludeLow7, nil
	case "full8":
		return preludeFull8, nil
	default:
		return 0, fmt.Errorf("%s must be low7 or full8", preludeEnv)
	}
}

func generatePrelude(policy preludePolicy) ([preludeLen]byte, error) {
	var prelude [preludeLen]byte
	if _, err := io.ReadFull(rand.Reader, prelude[:]); err != nil {
		return prelude, err
	}
	if policy == preludeLow7 {
		for i := range prelude {
			prelude[i] &= 0x7f
		}
	}
	return prelude, nil
}

// WrapTCPClient writes the 64-byte prelude and 12-byte nonce, then masks TLS bytes.
func WrapTCPClient(conn net.Conn, keys Keys) (net.Conn, error) {
	policy, err := preludePolicyFromEnv()
	if err != nil {
		return nil, err
	}
	prelude, err := generatePrelude(policy)
	if err != nil {
		return nil, err
	}
	var nonce [nonceLen]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}
	pref := make([]byte, 0, preludeLen+nonceLen)
	pref = append(pref, prelude[:]...)
	pref = append(pref, nonce[:]...)
	return &tcpConn{
		Conn:      conn,
		readKey:   keys.TCPS2C,
		writeKey:  keys.TCPC2S,
		nonce:     nonce,
		writePref: pref,
	}, nil
}

// WrapTCPServer consumes the client prelude and nonce without interpreting the prelude.
func WrapTCPServer(conn net.Conn, keys Keys) net.Conn {
	return &tcpConn{
		Conn:        conn,
		readKey:     keys.TCPC2S,
		writeKey:    keys.TCPS2C,
		readDiscard: preludeLen,
		readNeed:    nonceLen,
	}
}

type tcpConn struct {
	net.Conn
	readKey, writeKey [32]byte
	nonce             [nonceLen]byte
	readOff, writeOff uint64
	writePref         []byte
	readDiscard       int
	readNeed          int
	readMu, writeMu   sync.Mutex
}

func (c *tcpConn) Read(p []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()
	if c.readDiscard > 0 {
		scratch := make([]byte, c.readDiscard)
		if _, err := io.ReadFull(c.Conn, scratch); err != nil {
			return 0, err
		}
		c.readDiscard = 0
	}
	if c.readNeed > 0 {
		buf := make([]byte, c.readNeed)
		if _, err := io.ReadFull(c.Conn, buf); err != nil {
			return 0, err
		}
		copy(c.nonce[nonceLen-c.readNeed:], buf)
		c.readNeed = 0
	}
	if len(p) == 0 {
		return 0, nil
	}
	if err := checkLimit(c.readOff, len(p)); err != nil {
		return 0, err
	}
	n, err := c.Conn.Read(p)
	if n > 0 {
		chacha20.XORKeyStreamAt(p[:n], p[:n], &c.readKey, &c.nonce, c.readOff)
		c.readOff += uint64(n)
	}
	return n, err
}

func (c *tcpConn) Write(p []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if len(c.writePref) > 0 {
		n, err := c.Conn.Write(c.writePref)
		c.writePref = c.writePref[n:]
		if err != nil || len(c.writePref) > 0 {
			if err == nil {
				err = io.ErrShortWrite
			}
			return 0, err
		}
	}
	if len(p) == 0 {
		return 0, nil
	}
	if err := checkLimit(c.writeOff, len(p)); err != nil {
		return 0, err
	}
	buf := append([]byte(nil), p...)
	chacha20.XORKeyStreamAt(buf, buf, &c.writeKey, &c.nonce, c.writeOff)
	n, err := c.Conn.Write(buf)
	c.writeOff += uint64(n)
	return n, err
}

func checkLimit(offset uint64, n int) error {
	if n < 0 || offset+uint64(n) > streamLimit {
		return errors.New("nw3: Morph TCP keystream exhausted")
	}
	return nil
}

// WrapPacketConn masks each UDP datagram with a fresh nonce.
// Client sockets send with udp c2s and receive with udp s2c.
func WrapPacketConn(pc net.PacketConn, keys Keys, client bool) net.PacketConn {
	tx, rx := keys.UDPC2S, keys.UDPS2C
	if !client {
		tx, rx = keys.UDPS2C, keys.UDPC2S
	}
	return &packetConn{PacketConn: pc, txKey: tx, rxKey: rx}
}

type packetConn struct {
	net.PacketConn
	txKey [32]byte
	rxKey [32]byte
	mu    sync.Mutex
}

func (c *packetConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	var nonce [nonceLen]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return 0, err
	}
	buf := make([]byte, nonceLen+len(p))
	copy(buf, nonce[:])
	if len(p) > 0 {
		chacha20.XORKeyStreamAt(buf[nonceLen:], p, &c.txKey, &nonce, 0)
	}
	n, err := c.PacketConn.WriteTo(buf, addr)
	if n < nonceLen {
		return 0, err
	}
	return n - nonceLen, err
}

func (c *packetConn) ReadFrom(p []byte) (int, net.Addr, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	buf := make([]byte, nonceLen+len(p))
	n, addr, err := c.PacketConn.ReadFrom(buf)
	if err != nil {
		return 0, addr, err
	}
	if n <= nonceLen {
		return 0, addr, nil
	}
	var nonce [nonceLen]byte
	copy(nonce[:], buf[:nonceLen])
	payload := buf[nonceLen:n]
	if len(p) < len(payload) {
		payload = payload[:len(p)]
	}
	chacha20.XORKeyStreamAt(p, payload, &c.rxKey, &nonce, 0)
	return len(payload), addr, nil
}
