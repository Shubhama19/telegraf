package kube_inventory

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCustomResource(t *testing.T) {
	cli := &client{}
	selectInclude := []string{}
	selectExclude := []string{}
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())

	tests := []struct {
		name     string
		handler  *mockHandler
		output   []telegraf.Metric
		hasError bool
	}{
		{
			name: "no custom resource set",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/customresource/": &unstructured.UnstructuredList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect custom resource",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/customresource/": &unstructured.UnstructuredList{
						Items: []unstructured.Unstructured{
							{
								Object: map[string]interface{}{
									"metadata": map[string]interface{}{
										"name":              "testObj",
										"namespace":         "testns",
										"generation":        int64(123),
										"creationTimestamp": now.Format("2006-01-02T15:04:05Z07:00"),
									},
								},
							},
						},
					},
				},
			},
			output: []telegraf.Metric{
				testutil.MustMetric(
					"kubernetes_custom_resource",
					map[string]string{
						"name":      "testObj",
						"namespace": "testns",
					},
					map[string]interface{}{
						"generation": int64(123),
						"created":    now.UnixNano(),
					},
					time.Unix(0, 0),
				),
			},
			hasError: false,
		},
	}

	for _, v := range tests {
		ks := &KubernetesInventory{
			client:          cli,
			SelectorInclude: selectInclude,
			SelectorExclude: selectExclude,
		}
		require.NoError(t, ks.createSelectorFilters())
		acc := new(testutil.Accumulator)
		items := ((v.handler.responseMap["/customresource/"]).(*unstructured.UnstructuredList)).Items
		for i := range items {
			ks.gatherCustomResource(&items[i], acc)
		}

		err := acc.FirstError()
		if v.hasError {
			require.Errorf(t, err, "%s failed, should have error", v.name)
			continue
		}

		// No error case
		require.NoErrorf(t, err, "%s failed, err: %v", v.name, err)

		require.Len(t, acc.Metrics, len(v.output))
		testutil.RequireMetricsEqual(t, acc.GetTelegrafMetrics(), v.output, testutil.IgnoreTime())
	}
}
