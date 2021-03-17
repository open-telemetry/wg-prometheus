package configparser

import (
	"testing"

	"github.com/dyweb/gommon/util/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractPrometheusK8sSD(t *testing.T) {
	b := testutil.ReadFixture(t, "testdata/otel_one_prom.yaml")
	extracted, err := ExtractPrometheusK8sSD(b)
	require.Nil(t, err)
	//t.Logf("extracted \n%s", string(extracted))
	//testutil.WriteFixture(t, "testdata/otel_one_prom_extracted.yaml", extracted)
	expected := testutil.ReadFixture(t, "testdata/otel_one_prom_extracted.yaml")
	assert.Equal(t, string(expected), string(extracted))
}

func TestReplacePrometheusK8sSD(t *testing.T) {
	b := testutil.ReadFixture(t, "testdata/otel_one_prom.yaml")
	replaced, err := ReplacePrometheusK8sSD(b, "/etc/clb")
	require.Nil(t, err)
	//t.Logf("replaced \n%s", string(replaced))
	//testutil.WriteFixture(t, "testdata/otel_one_prom_replaced.yaml", replaced)
	expected := testutil.ReadFixture(t, "testdata/otel_one_prom_replaced.yaml")
	assert.Equal(t, string(expected), string(replaced))
}
