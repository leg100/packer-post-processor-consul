package consul

import (
	"os"
	"testing"
	"reflect"

	"github.com/mitchellh/packer/packer"
)


func testUi() *BasicUi {
	return &BasicUi{
		Reader:      new(bytes.Buffer),
		Writer:      new(bytes.Buffer),
		ErrorWriter: new(bytes.Buffer),
	}
}

// func TestPostProcessor(t *testing.T) {
// 	var p PostProcessor
// 	if err := p.Configure(validDefaults()); err != nil {
// 		t.Fatalf("err: %s", err)
// 	}
// 
//   func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
// 
//     if err := p.PostProcess(testUi(), Artifact{} 
// 		t.Fatal("should have client")
// 	}
// }

func TestPostProcessorConfigure(t *testing.T) {
	var p PostProcessor
	if err := p.Configure(validDefaults()); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.client == nil {
		t.Fatal("should have client")
	}
}

func TestPostProcessorConfigure_buildId(t *testing.T) {
	defer os.Setenv(BuildEnvKey, os.Getenv(BuildEnvKey))
	os.Setenv(BuildEnvKey, "5")

	var p PostProcessor
	if err := p.Configure(validDefaults()); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.buildId != 5 {
		t.Fatalf("bad: %#v", p.config.buildId)
	}
}

func TestPostProcessorMetadata(t *testing.T) {
	var p PostProcessor
	if err := p.Configure(validDefaults()); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact := new(packer.MockArtifact)
	metadata := p.metadata(artifact)
	if len(metadata) > 0 {
		t.Fatalf("bad: %#v", metadata)
	}
}

func TestPostProcessorMetadata_artifact(t *testing.T) {
	config := validDefaults()
	config["metadata"] = map[string]string{
		"foo": "bar",
	}

	var p PostProcessor
	if err := p.Configure(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact := new(packer.MockArtifact)
	artifact.StateValues = map[string]interface{}{
		ArtifactStateMetadata: map[interface{}]interface{}{
			"bar": "baz",
		},
	}

	metadata := p.metadata(artifact)
	expected := map[string]string{
		"foo": "bar",
		"bar": "baz",
	}
	if !reflect.DeepEqual(metadata, expected) {
		t.Fatalf("bad: %#v", metadata)
	}
}

func TestPostProcessorMetadata_config(t *testing.T) {
	config := validDefaults()
	config["metadata"] = map[string]string{
		"foo": "bar",
	}

	var p PostProcessor
	if err := p.Configure(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact := new(packer.MockArtifact)
	metadata := p.metadata(artifact)
	expected := map[string]string{
		"foo": "bar",
	}
	if !reflect.DeepEqual(metadata, expected) {
		t.Fatalf("bad: %#v", metadata)
	}
}

func TestPostProcessorType(t *testing.T) {
	var p PostProcessor
	if err := p.Configure(validDefaults()); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact := new(packer.MockArtifact)
	actual := p.artifactType(artifact)
	if actual != "foo" {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestPostProcessorType_artifact(t *testing.T) {
	var p PostProcessor
	if err := p.Configure(validDefaults()); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact := new(packer.MockArtifact)
	artifact.StateValues = map[string]interface{}{
		ArtifactStateType: "bar",
	}
	actual := p.artifactType(artifact)
	if actual != "bar" {
		t.Fatalf("bad: %#v", actual)
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
