package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/metacubex/mihomo/component/resource"
	C "github.com/metacubex/mihomo/constant"
	P "github.com/metacubex/mihomo/constant/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const externalProxiesYAML = `proxies:
  - {name: node-a, type: ss, server: 1.1.1.1, port: 8388, cipher: aes-128-gcm, password: p}
  - {name: node-b, type: ss, server: 2.2.2.2, port: 8388, cipher: aes-128-gcm, password: p}
`

const externalProxiesYAMLv2 = `proxies:
  - {name: node-c, type: ss, server: 3.3.3.3, port: 8388, cipher: aes-128-gcm, password: p}
`

type vehicleGetter interface {
	Vehicle() P.Vehicle
}

func setupExternalHome(t *testing.T) {
	t.Helper()
	old := C.Path.HomeDir()
	C.SetHomeDir(t.TempDir())
	t.Cleanup(func() { C.SetHomeDir(old) })
}

func parseExternal(t *testing.T, mapping map[string]any) P.ProxyProvider {
	t.Helper()
	pd, err := ParseProxyProvider("test", mapping, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		if c, ok := pd.(interface{ Close() error }); ok {
			_ = c.Close()
		}
	})
	return pd
}

func proxyNames(pd P.ProxyProvider) []string {
	var names []string
	for _, p := range pd.Proxies() {
		names = append(names, p.Name())
	}
	return names
}

func writeCache(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

// a) schema decoding + vehicle selection
func TestExternalSchemaDecodesFlag(t *testing.T) {
	setupExternalHome(t)
	pd := parseExternal(t, map[string]any{"type": "http", "url": "http://127.0.0.1:1/sub", "meowclash-external": true})
	_, ok := pd.(vehicleGetter)
	require.True(t, ok)
}

func TestExternalVehicleUsesHashPath(t *testing.T) {
	setupExternalHome(t)
	url := "http://127.0.0.1:1/sub"
	pd := parseExternal(t, map[string]any{"type": "http", "url": url, "meowclash-external": true})
	v := pd.(vehicleGetter).Vehicle()
	fv, ok := v.(*resource.FileVehicle)
	require.True(t, ok, "expected *FileVehicle, got %T", v)
	assert.Equal(t, C.Path.GetPathByHash("proxies", url), fv.Path())
	assert.True(t, fv.IsExternal())
}

func TestNonExternalKeepsHTTPVehicle(t *testing.T) {
	setupExternalHome(t)
	pd := parseExternal(t, map[string]any{"type": "http", "url": "http://127.0.0.1:1/sub"})
	_, ok := pd.(vehicleGetter).Vehicle().(*resource.HTTPVehicle)
	assert.True(t, ok)
	assert.Equal(t, P.HTTP, pd.VehicleType())
}

func TestExternalRespectsPathOverride(t *testing.T) {
	setupExternalHome(t)
	pd := parseExternal(t, map[string]any{"type": "http", "url": "http://127.0.0.1:1/sub", "path": "./custom/p.yaml", "meowclash-external": true})
	fv := pd.(vehicleGetter).Vehicle().(*resource.FileVehicle)
	assert.Equal(t, C.Path.Resolve("./custom/p.yaml"), fv.Path())
}

func TestExternalFlagIgnoredForFileAndInline(t *testing.T) {
	setupExternalHome(t)
	p := filepath.Join(C.Path.HomeDir(), "f.yaml")
	writeCache(t, p, externalProxiesYAML)
	pd := parseExternal(t, map[string]any{"type": "file", "path": "./f.yaml", "meowclash-external": true})
	fv, ok := pd.(vehicleGetter).Vehicle().(*resource.FileVehicle)
	require.True(t, ok)
	assert.False(t, fv.IsExternal())

	pi := parseExternal(t, map[string]any{"type": "inline", "payload": []map[string]any{
		{"name": "x", "type": "ss", "server": "1.1.1.1", "port": 1, "cipher": "aes-128-gcm", "password": "p"},
	}, "meowclash-external": true})
	assert.Equal(t, P.Inline, pi.VehicleType())
}

// b) zero network requests
func TestExternalZeroNetworkRequests(t *testing.T) {
	setupExternalHome(t)
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte(externalProxiesYAMLv2))
	}))
	defer srv.Close()

	writeCache(t, C.Path.GetPathByHash("proxies", srv.URL), externalProxiesYAML)
	pd := parseExternal(t, map[string]any{"type": "http", "url": srv.URL, "interval": 1, "meowclash-external": true})

	require.NoError(t, pd.Initial())
	require.NoError(t, pd.Update())
	assert.EqualValues(t, 0, hits.Load())
	assert.ElementsMatch(t, []string{"node-a", "node-b"}, proxyNames(pd))
}

// c) external cache overwrite is picked up by Update()
func TestExternalCacheOverwritePickedUp(t *testing.T) {
	setupExternalHome(t)
	url := "http://127.0.0.1:1/sub"
	path := C.Path.GetPathByHash("proxies", url)
	writeCache(t, path, externalProxiesYAML)
	pd := parseExternal(t, map[string]any{"type": "http", "url": url, "meowclash-external": true})
	require.NoError(t, pd.Initial())
	assert.ElementsMatch(t, []string{"node-a", "node-b"}, proxyNames(pd))

	writeCache(t, path, externalProxiesYAMLv2)
	require.NoError(t, pd.Update())
	assert.ElementsMatch(t, []string{"node-c"}, proxyNames(pd))
}

// d) missing cache file
func TestExternalMissingFile(t *testing.T) {
	setupExternalHome(t)
	pd := parseExternal(t, map[string]any{"type": "http", "url": "http://127.0.0.1:1/missing", "meowclash-external": true})
	err := pd.Initial()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "external provider file not found")
}

// e) VehicleType + JSON API
func TestExternalVehicleTypeAndJSON(t *testing.T) {
	setupExternalHome(t)
	url := "http://127.0.0.1:1/sub"
	writeCache(t, C.Path.GetPathByHash("proxies", url), externalProxiesYAML)
	pd := parseExternal(t, map[string]any{"type": "http", "url": url, "meowclash-external": true})
	require.NoError(t, pd.Initial())
	assert.Equal(t, P.File, pd.VehicleType())

	buf, err := json.Marshal(map[string]any{"providers": map[string]P.ProxyProvider{"test": pd}})
	require.NoError(t, err)
	var out struct {
		Providers map[string]struct {
			VehicleType string `json:"vehicleType"`
		} `json:"providers"`
	}
	require.NoError(t, json.Unmarshal(buf, &out))
	assert.Equal(t, "File", out.Providers["test"].VehicleType)

	plain := parseExternal(t, map[string]any{"type": "http", "url": url})
	assert.Equal(t, P.HTTP, plain.VehicleType())
}
