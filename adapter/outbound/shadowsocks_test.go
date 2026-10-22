package outbound

import (
	"testing"

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
