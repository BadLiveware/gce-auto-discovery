package gce_auto_discovery

import (
	"bytes"
	"context"
	"html/template"
	golog "log"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

type MockGcpInstances struct {
	instances []Instance
}

func (m *MockGcpInstances) listAllInstances(projectID string) ([]Instance, error) {
	return m.instances, nil
}

func TestStraightServe(t *testing.T) {
	gcpInstances := &MockGcpInstances{
		instances: []Instance{
			{
				Name: "test-instance",
				IP:   "10.0.0.0",
			},
		},
	}

	template := template.Must(template.New("test").Parse("example.org."))
	autodiscovery := GceAutoDiscovery{
		Next:      test.ErrorHandler(),
		client:    gcpInstances,
		template:  template,
		projectID: "test-project",
		interval:  0,
		zoneNames: []string{
			"example.org.",
		}}

	// Setup a new output buffer that is *not* standard output, so we can check if
	// example is really being printed.
	b := &bytes.Buffer{}
	golog.SetOutput(b)

	ctx := context.TODO()
	r := new(dns.Msg)
	r.SetQuestion("example.org.", dns.TypeA)
	// Create a new Recorder that captures the result, this isn't actually used in this test
	// as it just serves as something that implements the dns.ResponseWriter interface.
	rec := dnstest.NewRecorder(&test.ResponseWriter{})

	autodiscovery.Run(ctx)
	// Call our plugin directly, and check the result.
	ret, _ := autodiscovery.ServeDNS(ctx, rec, r)
	t.Log(ret)
	if rec.Msg.Rcode != dns.RcodeSuccess {
		t.Errorf("expected no error, but got %v", rec.Msg.Rcode)
	}
	if len(rec.Msg.Answer) != 1 {
		t.Errorf("expected 1 answer, but got %v", len(rec.Msg.Answer))
	}
	if rec.Msg.Answer[0].Header().Name != "example.org." {
		t.Errorf("expected example.org., but got %v", rec.Msg.Answer[0].Header().Name)
	}
}

func TestMultipleAnswers(t *testing.T) {
	gcpInstances := &MockGcpInstances{
		instances: []Instance{
			{
				Name: "test-instance",
				IP:   "10.0.0.0",
				Labels: map[string]string{
					"dns": "testing",
				},
			},
			{
				Name: "test-instance-2",
				IP:   "10.0.0.1",
				Labels: map[string]string{
					"dns": "testing",
				},
			},
		},
	}

	template := template.Must(template.New("test").Parse("{{ .Labels.dns }}.example.org."))
	autodiscovery := GceAutoDiscovery{
		Next:      test.ErrorHandler(),
		client:    gcpInstances,
		template:  template,
		projectID: "test-project",
		interval:  0,
		zoneNames: []string{
			"example.org.",
		}}

	// Setup a new output buffer that is *not* standard output, so we can check if
	// example is really being printed.
	b := &bytes.Buffer{}
	golog.SetOutput(b)

	ctx := context.TODO()
	r := new(dns.Msg)
	r.SetQuestion("testing.example.org.", dns.TypeA)
	// Create a new Recorder that captures the result, this isn't actually used in this test
	// as it just serves as something that implements the dns.ResponseWriter interface.
	rec := dnstest.NewRecorder(&test.ResponseWriter{})

	autodiscovery.Run(ctx)
	// Call our plugin directly, and check the result.
	ret, _ := autodiscovery.ServeDNS(ctx, rec, r)
	t.Log(ret)
	if rec.Msg.Rcode != dns.RcodeSuccess {
		t.Errorf("expected no error, but got %v", rec.Msg.Rcode)
	}
	if len(rec.Msg.Answer) != 2 {
		t.Errorf("expected 2 answer, but got %v", len(rec.Msg.Answer))
	}
	if rec.Msg.Answer[0].Header().Name != "testing.example.org." {
		t.Errorf("expected testing.example.org., but got %v", rec.Msg.Answer[0].Header().Name)
	}
	if rec.Msg.Answer[1].Header().Name != "testing.example.org." {
		t.Errorf("expected testing.example.org., but got %v", rec.Msg.Answer[1].Header().Name)
	}
}

func TestTemplating(t *testing.T) {
	gcpInstances := &MockGcpInstances{
		instances: []Instance{
			{
				Name: "test-instance",
				IP:   "10.0.0.0",
			},
		},
	}

	template := template.Must(template.New("test").Parse("{{.Name}}.example.org."))
	autodiscovery := GceAutoDiscovery{
		Next:      test.ErrorHandler(),
		client:    gcpInstances,
		template:  template,
		projectID: "test-project",
		interval:  0,
		zoneNames: []string{
			"example.org.",
		}}

	// Setup a new output buffer that is *not* standard output, so we can check if
	// example is really being printed.
	b := &bytes.Buffer{}
	golog.SetOutput(b)

	ctx := context.Background()
	r := new(dns.Msg)
	r.SetQuestion("test-instance.example.org.", dns.TypeA)
	// Create a new Recorder that captures the result, this isn't actually used in this test
	// as it just serves as something that implements the dns.ResponseWriter interface.
	rec := dnstest.NewRecorder(&test.ResponseWriter{})

	autodiscovery.Run(ctx)
	// Call our plugin directly, and check the result.
	ret, _ := autodiscovery.ServeDNS(ctx, rec, r)
	t.Log(ret)
	if rec.Msg.Rcode != dns.RcodeSuccess {
		t.Errorf("expected no error, but got %v", rec.Msg.Rcode)
	}
	if len(rec.Msg.Answer) != 1 {
		t.Errorf("expected 1 answer, but got %v", len(rec.Msg.Answer))
	}
	if rec.Msg.Answer[0].Header().Name != "test-instance.example.org." {
		t.Errorf("expected test-instance.example.org., but got %v", rec.Msg.Answer[0].Header().Name)
	}
}
