package outbound

import (
	"io"
	"net"
	"testing"

	M "github.com/metacubex/sing/common/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeOutlinePrefix(t *testing.T) {
	got, err := decodeOutlinePrefix("POST ")
	require.NoError(t, err)
	assert.Equal(t, []byte("POST "), got)

	got, err = decodeOutlinePrefix("\u0013\u0003")
	require.NoError(t, err)
	assert.Equal(t, []byte{0x13, 0x03}, got)

	got, err = decodeOutlinePrefix("\u00ff\u0080")
	require.NoError(t, err)
	assert.Equal(t, []byte{0xff, 0x80}, got)

	got, err = decodeOutlinePrefix("")
	require.NoError(t, err)
	assert.Nil(t, got)

	_, err = decodeOutlinePrefix("a\u0100")
	assert.Error(t, err)
	_, err = decodeOutlinePrefix("中")
	assert.Error(t, err)
}

func TestNewShadowSocksPrefix(t *testing.T) {
	base := ShadowSocksOption{
		Name:     "ss",
		Server:   "127.0.0.1",
		Port:     8388,
		Password: "password",
		Cipher:   "chacha20-ietf-poly1305",
		UDP:      true,
	}

	t.Run("both", func(t *testing.T) {
		opt := base
		opt.Prefix = "POST "
		opt.UDPPrefix = "k{\u0001 "
		ss, err := NewShadowSocks(opt)
		require.NoError(t, err)
		assert.NotNil(t, ss)
	})

	t.Run("tcp only", func(t *testing.T) {
		opt := base
		opt.Prefix = "POST "
		_, err := NewShadowSocks(opt)
		require.NoError(t, err)
	})

	t.Run("udp only", func(t *testing.T) {
		opt := base
		opt.UDPPrefix = "k{\u0001 "
		_, err := NewShadowSocks(opt)
		require.NoError(t, err)
	})

	t.Run("invalid codepoint", func(t *testing.T) {
		opt := base
		opt.UDPPrefix = "\u0100"
		_, err := NewShadowSocks(opt)
		assert.ErrorContains(t, err, "U+0000-U+00FF")
	})

	t.Run("prefix too long", func(t *testing.T) {
		opt := base
		opt.Prefix = "0123456789012345678901234567890123456789"
		_, err := NewShadowSocks(opt)
		assert.Error(t, err)
	})

	t.Run("2022 rejects prefix", func(t *testing.T) {
		for _, tc := range []struct{ tcp, udp string }{{"POST ", ""}, {"", "POST "}, {"POST ", "POST "}} {
			opt := base
			opt.Cipher = "2022-blake3-aes-128-gcm"
			opt.Password = "MDEyMzQ1Njc4OWFiY2RlZg=="
			opt.Prefix = tc.tcp
			opt.UDPPrefix = tc.udp
			_, err := NewShadowSocks(opt)
			assert.ErrorContains(t, err, "prefix is only supported for AEAD ciphers")
		}
	})

	t.Run("2022 without prefix still works", func(t *testing.T) {
		opt := base
		opt.Cipher = "2022-blake3-aes-128-gcm"
		opt.Password = "MDEyMzQ1Njc4OWFiY2RlZg=="
		_, err := NewShadowSocks(opt)
		require.NoError(t, err)
	})
}

func TestNewShadowSocksPrefixOnWire(t *testing.T) {
	const tcpPrefix, udpPrefix = "POST ", "k{\u0001 "
	ss, err := NewShadowSocks(ShadowSocksOption{
		Server:    "127.0.0.1",
		Port:      8388,
		Password:  "password",
		Cipher:    "chacha20-ietf-poly1305",
		Prefix:    tcpPrefix,
		UDPPrefix: udpPrefix,
	})
	require.NoError(t, err)

	captureSalt := func(write func(net.Conn)) []byte {
		client, server := net.Pipe()
		t.Cleanup(func() { _ = client.Close() })
		t.Cleanup(func() { _ = server.Close() })
		done := make(chan struct{})
		go func() {
			defer close(done)
			write(client)
		}()

		salt := make([]byte, 32)
		_, err := io.ReadFull(server, salt)
		_ = server.Close()
		<-done
		require.NoError(t, err)
		return salt
	}

	tcpSalt := captureSalt(func(conn net.Conn) {
		_, _ = ss.method.DialEarlyConn(conn, M.ParseSocksaddr("example.com:80")).Write([]byte("test"))
	})
	assert.Equal(t, tcpPrefix, string(tcpSalt[:len(tcpPrefix)]))

	udpSalt := captureSalt(func(conn net.Conn) {
		_, _ = ss.method.DialPacketConn(conn).WriteTo([]byte("test"), &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 53})
	})
	assert.Equal(t, udpPrefix, string(udpSalt[:len(udpPrefix)]))
}
