package sidecars

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	networking_v1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/tests/data"
	"github.com/kiali/kiali/tests/testutils/validations"
)

func TestEgressHostFormatCorrect(t *testing.T) {
	assert := assert.New(t)

	vals, valid := EgressHostChecker{
		Services:       fakeServices([]string{"details", "reviews"}),
		ServiceEntries: kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{data.CreateExternalServiceEntry()}),
		Sidecar: *sidecarWithHosts([]string{
			"*/*",
			"~/*",
			"./*",
			"./reviews.bookinfo.svc.cluster.local",
			"./*.bookinfo.svc.cluster.local",
			"./wikipedia.org",
			"bookinfo/*",
			"bookinfo/*.bookinfo.svc.cluster.local",
			"bookinfo/reviews.bookinfo.svc.cluster.local",
			"bookinfo/wikipedia.org",
		}),
	}.Check()

	assert.Empty(vals)
	assert.True(valid)
}

func TestEgressExportedInternalServiceEntryPresent(t *testing.T) {
	assert := assert.New(t)

	vals, valid := EgressHostChecker{
		Services:       []core_v1.Service{},
		ServiceEntries: kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*data.CreateEmptyMeshInternalServiceEntry("details-se", "bookinfo3", []string{"details.bookinfo2.svc.cluster.local"})}),
		Sidecar: *sidecarWithHosts([]string{
			"bookinfo/details.bookinfo2.svc.cluster.local",
		}),
	}.Check()

	assert.Empty(vals)
	assert.True(valid)
}

func TestEgressExportedExternalServiceEntryPresent(t *testing.T) {
	assert := assert.New(t)

	vals, valid := EgressHostChecker{
		Services:       []core_v1.Service{},
		ServiceEntries: kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*data.CreateEmptyMeshExternalServiceEntry("details-se", "bookinfo3", []string{"www.myhost.com"})}),
		Sidecar: *sidecarWithHosts([]string{
			"bookinfo/www.myhost.com",
		}),
	}.Check()

	assert.Empty(vals)
	assert.True(valid)
}

func TestWildcardHostEgressExportedExternalServiceEntryNotPresent(t *testing.T) {
	assert := assert.New(t)

	vals, valid := EgressHostChecker{
		Services:       []core_v1.Service{},
		ServiceEntries: kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*data.CreateEmptyMeshExternalServiceEntry("details-se", "bookinfo3", []string{"www.myhost.com"})}),
		Sidecar: *sidecarWithHosts([]string{
			"bookinfo/*.myhost.com",
		}),
	}.Check()

	assert.NotEmpty(vals)
	assert.True(valid)
	assert.Equal(models.WarningSeverity, vals[0].Severity)
	assert.Equal("spec/egress[0]/hosts[0]", vals[0].Path)
	assert.NoError(validations.ConfirmIstioCheckMessage("sidecar.egress.servicenotfound", vals[0]))
}

func TestEgressExportedExternalWildcardServiceEntryPresent(t *testing.T) {
	assert := assert.New(t)

	vals, valid := EgressHostChecker{
		Services:       []core_v1.Service{},
		ServiceEntries: kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*data.CreateEmptyMeshExternalServiceEntry("details-se", "bookinfo3", []string{"*.myhost.com"})}),
		Sidecar: *sidecarWithHosts([]string{
			"bookinfo/www.myhost.com",
		}),
	}.Check()

	assert.Empty(vals)
	assert.True(valid)
}

func TestEgressExportedInternalServiceEntryNotPresent(t *testing.T) {
	assert := assert.New(t)

	vals, valid := EgressHostChecker{
		Services:       []core_v1.Service{},
		ServiceEntries: kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*data.CreateEmptyMeshInternalServiceEntry("details-se", "bookinfo3", []string{"details.bookinfo2.svc.cluster.local"})}),
		Sidecar: *sidecarWithHosts([]string{
			"bookinfo/details.bookinfo.svc.cluster.local",
		}),
	}.Check()

	assert.NotEmpty(vals)
	assert.True(valid)
	assert.Equal(models.WarningSeverity, vals[0].Severity)
	assert.Equal("spec/egress[0]/hosts[0]", vals[0].Path)
	assert.NoError(validations.ConfirmIstioCheckMessage("sidecar.egress.servicenotfound", vals[0]))
}

func TestEgressExportedExternalServiceEntryNotPresent(t *testing.T) {
	assert := assert.New(t)

	vals, valid := EgressHostChecker{
		Services:       []core_v1.Service{},
		ServiceEntries: kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*data.CreateEmptyMeshExternalServiceEntry("details-se", "bookinfo3", []string{"www.myhost.com"})}),
		Sidecar: *sidecarWithHosts([]string{
			"bookinfo/www.wrong.com",
		}),
	}.Check()

	assert.NotEmpty(vals)
	assert.True(valid)
	assert.Equal(models.WarningSeverity, vals[0].Severity)
	assert.Equal("spec/egress[0]/hosts[0]", vals[0].Path)
	assert.NoError(validations.ConfirmIstioCheckMessage("sidecar.egress.servicenotfound", vals[0]))
}

func TestEgressExportedWildcardInternalServiceEntryPresent(t *testing.T) {
	assert := assert.New(t)

	vals, valid := EgressHostChecker{
		Services:       []core_v1.Service{},
		ServiceEntries: kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*data.CreateEmptyMeshInternalServiceEntry("details-se", "bookinfo3", []string{"*.bookinfo2.svc.cluster.local"})}),
		Sidecar: *sidecarWithHosts([]string{
			"bookinfo/details.bookinfo2.svc.cluster.local",
		}),
	}.Check()

	assert.Empty(vals)
	assert.True(valid)
}

func TestEgressExportedWildcardInternalServiceEntryNotPresent(t *testing.T) {
	assert := assert.New(t)

	vals, valid := EgressHostChecker{
		Services:       []core_v1.Service{},
		ServiceEntries: kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*data.CreateEmptyMeshInternalServiceEntry("details-se", "bookinfo3", []string{"*.bookinfo3.svc.cluster.local"})}),
		Sidecar: *sidecarWithHosts([]string{
			"bookinfo/*.bookinfo2.svc.cluster.local",
		}),
	}.Check()

	assert.NotEmpty(vals)
	assert.True(valid)
	assert.Equal(models.WarningSeverity, vals[0].Severity)
	assert.Equal("spec/egress[0]/hosts[0]", vals[0].Path)
	assert.NoError(validations.ConfirmIstioCheckMessage("sidecar.egress.servicenotfound", vals[0]))
}

func TestEgressExportedNonFQDNInternalServiceEntryNotPresent(t *testing.T) {
	assert := assert.New(t)

	vals, valid := EgressHostChecker{
		Services:       []core_v1.Service{},
		ServiceEntries: kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*data.CreateEmptyMeshInternalServiceEntry("details-se", "bookinfo3", []string{"details"})}),
		Sidecar: *sidecarWithHosts([]string{
			"bookinfo/details.bookinfo2.svc.cluster.local",
		}),
	}.Check()

	assert.NotEmpty(vals)
	assert.True(valid)
	assert.Equal(models.WarningSeverity, vals[0].Severity)
	assert.Equal("spec/egress[0]/hosts[0]", vals[0].Path)
	assert.NoError(validations.ConfirmIstioCheckMessage("sidecar.egress.servicenotfound", vals[0]))
}

func TestEgressHostCrossNamespaceServiceNotFound(t *testing.T) {
	assert := assert.New(t)

	hosts := []string{
		"*/*.example.com",
		"*/www.example.com",
		"*/example.prod.svc.cluster.local",
		"~/*.example.com",
		"~/www.example.com",
		"~/example.prod.svc.cluster.local",
		"bookinfo/reviews.bogus.svc.cluster.local",
		"bookinfo/*.bogus.svc.cluster.local",
	}

	vals, valid := EgressHostChecker{
		Sidecar: *sidecarWithHosts(hosts),
	}.Check()

	assert.NotEmpty(vals)
	assert.Len(vals, len(hosts))
	assert.True(valid)

	for i, c := range vals {
		assert.Equal(models.WarningSeverity, c.Severity)
		assert.Equal(fmt.Sprintf("spec/egress[0]/hosts[%d]", i), c.Path)
		assert.NoError(validations.ConfirmIstioCheckMessage("sidecar.egress.servicenotfound", c))
	}
}

func TestEgressInvalidHostFormat(t *testing.T) {
	assert := assert.New(t)

	vals, valid := EgressHostChecker{
		Sidecar: *sidecarWithHosts([]string{
			"no-dash-used",
		}),
	}.Check()

	assert.NotEmpty(vals)
	assert.Len(vals, 1)
	assert.False(valid)

	assert.Equal(models.ErrorSeverity, vals[0].Severity)
	assert.Equal("spec/egress[0]/hosts[0]", vals[0].Path)
	assert.NoError(validations.ConfirmIstioCheckMessage("sidecar.egress.invalidhostformat", vals[0]))
}

func TestEgressServiceNotFound(t *testing.T) {
	assert := assert.New(t)

	vals, valid := EgressHostChecker{
		Sidecar: *sidecarWithHosts([]string{
			"bookinfo/boggus.bookinfo.svc.cluster.local",
			"bookinfo/boggus.org",
		}),
	}.Check()

	assert.NotEmpty(vals)
	assert.Len(vals, 2)
	assert.True(valid)

	for i, c := range vals {
		assert.Equal(models.WarningSeverity, c.Severity)
		assert.Equal(fmt.Sprintf("spec/egress[0]/hosts[%d]", i), c.Path)
		assert.NoError(validations.ConfirmIstioCheckMessage("sidecar.egress.servicenotfound", c))
	}
}

func sidecarWithHosts(hl []string) *networking_v1alpha3.Sidecar {
	return data.AddHostsToSidecar(hl, data.CreateSidecar("sidecar", "bookinfo"))
}

func fakeServices(serviceNames []string) []core_v1.Service {
	services := make([]core_v1.Service, 0, len(serviceNames))

	for _, sName := range serviceNames {
		service := core_v1.Service{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      sName,
				Namespace: "bookinfo",
				Labels: map[string]string{
					"app":     sName,
					"version": "v1"}},
			Spec: core_v1.ServiceSpec{
				ClusterIP: "fromservice",
				Type:      "ClusterIP",
				Selector:  map[string]string{"app": sName},
			},
		}

		services = append(services, service)
	}

	return services
}
