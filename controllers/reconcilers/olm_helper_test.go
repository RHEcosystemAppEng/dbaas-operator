package reconcilers

import (
	"testing"
)

type fields struct {
	namespace string
	name      string
}

func Test_GetClusterServiceVersion(t *testing.T) {

	test := struct {
		name   string
		fields fields
		want   string
	}{
		name: "Get ClusterServiceVersion",
		fields: fields{
			namespace: "test-namespace",
			name:      "provider-csv",
		},
		want: "provider-csv",
	}
	t.Run(test.name, func(t *testing.T) {
		got := GetClusterServiceVersion(test.fields.namespace, test.fields.name)
		if got.ObjectMeta.Name != test.want {
			t.Errorf("GetClusterServiceVersion() got = %+v, want %+v", got.ObjectMeta.Name, test.want)
		}
	})
}

func Test_GetGetSubscription(t *testing.T) {

	test := struct {
		name   string
		fields fields
		want   string
	}{
		name: "Get Subscription",
		fields: fields{
			namespace: "test-namespace",
			name:      "provider-subscription",
		},
		want: "provider-subscription",
	}
	t.Run(test.name, func(t *testing.T) {
		got := GetSubscription(test.fields.namespace, test.fields.name)
		if got.ObjectMeta.Name != test.want {
			t.Errorf("Test_GetGetSubscription() got = %+v, want %+v", got.ObjectMeta.Name, test.want)
		}
	})

}

func Test_GetOperatorGroup(t *testing.T) {

	test := struct {
		name   string
		fields fields
		want   string
	}{
		name: "Get OperatorGroup",
		fields: fields{
			namespace: "test-namespace",
			name:      "test-operatorgroup",
		},
		want: "test-operatorgroup",
	}
	t.Run(test.name, func(t *testing.T) {
		got := GetOperatorGroup(test.fields.namespace, test.fields.name)
		if got.ObjectMeta.Name != test.want {
			t.Errorf("Test_GetOperatorGroup() got = %+v, want %+v", got.ObjectMeta.Name, test.want)
		}
	})

}

func Test_GetCatalogSource(t *testing.T) {

	test := struct {
		name   string
		fields fields
		want   string
	}{
		name: "Get GetCatalogSource",
		fields: fields{
			namespace: "test-namespace",
			name:      "provider-catalogsource",
		},
		want: "provider-catalogsource",
	}
	t.Run(test.name, func(t *testing.T) {
		got := GetCatalogSource(test.fields.namespace, test.fields.name)
		if got.ObjectMeta.Name != test.want {
			t.Errorf("Test_GetCatalogSource() got = %+v, want %+v", got.ObjectMeta.Name, test.want)
		}
	})

}
