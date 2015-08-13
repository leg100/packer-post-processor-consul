package consul

import (
	"testing"
)

func TestPostProcessorConfigure(t *testing.T) {
	var p PostProcessor
	if err := p.Configure(validDefaults()); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.client == nil {
		t.Fatal("should have client")
	}
}

func validDefaults() map[string]interface{} {
	return map[string]interface{}{
		"artifact":      "mitchellh/test",
		"artifact_type": "foo",
		"test":          true,
    "consul_address":      "consul:8500",
		"aws_access_key":      "ABC123",
		"aws_secret_key":      "123123",
		"project_name":        "kafka",
    "project_version":     "2",
	}
}
