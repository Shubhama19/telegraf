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

		ki.gatherCustomResourceCount(len(list.Items), acc)

		for i := range list.Items {
			ki.gatherCustomResource(&list.Items[i], acc)
		}
	}
}

func (ki KubernetesInventory) gatherCustomResourceCount(count int, acc telegraf.Accumulator) {
	fields := map[string]interface{}{"count": count}
	tags := map[string]string{}

	acc.AddFields(customResourceMeasurement, fields, tags)
}

func (ki *KubernetesInventory) gatherCustomResource(cr *unstructured.Unstructured, acc telegraf.Accumulator) {
	fields := map[string]interface{}{
		"created":    cr.GetCreationTimestamp().UnixNano(),
		"generation": cr.GetGeneration(),
	}
	tags := map[string]string{
		"name":      cr.GetName(),
		"namespace": cr.GetNamespace(),
	}

	acc.AddFields(customResourceMeasurement, fields, tags)
}
