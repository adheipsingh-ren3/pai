package controller

import (
	"testing"

	observabilityv1alpha1 "github.com/parseable/pai/api/v1alpha1"
)

func TestEffectiveScrapeConfigsTraefikDisabledByDefault(t *testing.T) {
	configs := effectiveScrapeConfigs(&observabilityv1alpha1.MetricsConfig{})
	if len(configs) != 0 {
		t.Fatalf("expected no scrape configs, got %d", len(configs))
	}
}

func TestEffectiveScrapeConfigsTraefikDefaults(t *testing.T) {
	configs := effectiveScrapeConfigs(&observabilityv1alpha1.MetricsConfig{
		Traefik: &observabilityv1alpha1.TraefikMetricsConfig{
			Enabled:       true,
			TargetDataset: "traefik-metrics",
		},
	})

	if len(configs) != 1 {
		t.Fatalf("expected one scrape config, got %d", len(configs))
	}
	got := configs[0]
	if got.Name != "traefik" || got.Port != 9100 || got.URI != "/metrics" {
		t.Fatalf("unexpected Traefik defaults: %#v", got)
	}
	if got.TargetDataset != "traefik-metrics" {
		t.Fatalf("unexpected dataset: %q", got.TargetDataset)
	}
	if got.PodSelector["app.kubernetes.io/name"] != "traefik" {
		t.Fatalf("unexpected pod selector: %#v", got.PodSelector)
	}
}

func TestEffectiveScrapeConfigsTraefikOverrides(t *testing.T) {
	metrics := &observabilityv1alpha1.MetricsConfig{
		Traefik: &observabilityv1alpha1.TraefikMetricsConfig{
			Enabled:       true,
			TargetDataset: "edge-metrics",
			Port:          9200,
			URI:           "custom-metrics",
			Headers:       map[string]string{"X-P-Team": "platform"},
			NamespaceSelector: observabilityv1alpha1.NamespaceSelector{
				Mode:       "include",
				Namespaces: []string{"edge"},
			},
		},
		ScrapeConfigs: []observabilityv1alpha1.ScrapeConfig{{
			Name: "custom", Port: 8080, URI: "/metrics", TargetDataset: "custom-metrics",
		}},
	}

	configs := effectiveScrapeConfigs(metrics)
	if len(configs) != 2 {
		t.Fatalf("expected built-in and custom scrape configs, got %d", len(configs))
	}
	got := configs[0]
	if got.Port != 9200 || got.URI != "custom-metrics" || got.TargetDataset != "edge-metrics" {
		t.Fatalf("overrides not preserved: %#v", got)
	}
	if got.Headers["X-P-Team"] != "platform" || len(got.NamespaceSelector.Namespaces) != 1 {
		t.Fatalf("headers or namespace selector not preserved: %#v", got)
	}
}
