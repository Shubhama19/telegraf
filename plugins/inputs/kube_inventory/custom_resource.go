package kube_inventory

import (
	"context"

	"github.com/influxdata/telegraf"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func collectCustomResource(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	for _, resource := range ki.customResources {
		list, err := ki.client.getCustomResource(ctx, resource)
		if err != nil {
			acc.AddError(err)
			return
		}
		for i := range list.Items {
			ki.gatherCRD(&list.Items[i], acc)
		}
	}
}

func (ki *KubernetesInventory) gatherCRD(d *unstructured.Unstructured, acc telegraf.Accumulator) {
	fields := map[string]interface{}{
		"created": d.GetCreationTimestamp().UnixNano(),
	}
	tags := map[string]string{
		"name":      d.GetName(),
		"namespace": d.GetNamespace(),
	}

	acc.AddFields(customResourceMeasurement, fields, tags)
}
