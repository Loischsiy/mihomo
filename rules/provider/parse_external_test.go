package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/metacubex/mihomo/component/resource"
	C "github.com/metacubex/mihomo/constant"
	P "github.com/metacubex/mihomo/constant/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ruleVehicleGetter interface {
	Vehicle() P.Vehicle
}

func parseExternalRule(t *testing.T, mapping map[string]any) P.RuleProvider {
	t.Helper()
	old := C.Path.HomeDir()
	C.SetHomeDir(t.TempDir())
	t.Cleanup(func() { C.SetHomeDir(old) })
	rp, err := ParseRuleProvider("test", mapping, nil, func(string) resource.BundleFile { return nil })
	require.NoError(t, err)
	t.Cleanup(func() {
		if c, ok := rp.(interface{ Close() error }); ok {
			_ = c.Close()
		}
	})
	return rp
}

func TestExternalRuleVehicleUsesHashPath(t *testing.T) {
	url := "http://127.0.0.1:1/rules"
	rp := parseExternalRule(t, map[string]any{"type": "http", "behavior": "domain", "url": url, "meowclash-external": true})
	fv, ok := rp.(ruleVehicleGetter).Vehicle().(*resource.FileVehicle)
	require.True(t, ok)
	assert.Equal(t, C.Path.GetPathByHash("rules", url), fv.Path())
	assert.True(t, fv.IsExternal())
	assert.Equal(t, P.File, rp.VehicleType())
}

func TestExternalRuleRespectsPathOverride(t *testing.T) {
	rp := parseExternalRule(t, map[string]any{"type": "http", "behavior": "domain", "url": "http://127.0.0.1:1/rules", "path": "./custom/r.yaml", "meowclash-external": true})
	fv := rp.(ruleVehicleGetter).Vehicle().(*resource.FileVehicle)
	assert.Equal(t, C.Path.Resolve("./custom/r.yaml"), fv.Path())
}

func TestNonExternalRuleKeepsHTTPVehicle(t *testing.T) {
	rp := parseExternalRule(t, map[string]any{"type": "http", "behavior": "domain", "url": "http://127.0.0.1:1/rules"})
	_, ok := rp.(ruleVehicleGetter).Vehicle().(*resource.HTTPVehicle)
	assert.True(t, ok)
	assert.Equal(t, P.HTTP, rp.VehicleType())
}

func TestExternalRuleMissingFile(t *testing.T) {
	rp := parseExternalRule(t, map[string]any{"type": "http", "behavior": "domain", "url": "http://127.0.0.1:1/missing", "meowclash-external": true})
	err := rp.Initial()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "external provider file not found")
}

func TestExternalRuleFlagIgnoredForFile(t *testing.T) {
	rp := parseExternalRule(t, map[string]any{"type": "file", "behavior": "domain", "path": "./r.yaml", "meowclash-external": true})
	fv, ok := rp.(ruleVehicleGetter).Vehicle().(*resource.FileVehicle)
	require.True(t, ok)
	assert.False(t, fv.IsExternal())
	_ = os.MkdirAll(filepath.Dir(fv.Path()), 0o755)
}
