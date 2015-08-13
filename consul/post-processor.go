package consul

import (
	"fmt"
	"strings"
	"encoding/json"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
        "github.com/hashicorp/consul/api"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/template/interpolate"
)

const BuildEnvKey = "CONSUL_BUILD_ID"

// Artifacts can return a string for this state key and the post-processor
// will use automatically use this as the type. The user's value overrides
// this if `artifact_type_override` is set to true.
const ArtifactStateType = "consul.artifact.type"

// Artifacts can return a map[string]string for this state key and this
// post-processor will automatically merge it into the metadata for any
// uploaded artifact versions.
const ArtifactStateMetadata = "consul.artifact.metadata"

var builtins = map[string]string{
	"mitchellh.amazonebs": "amazonebs",
	"mitchellh.amazon.instance": "amazoninstance",
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Artifact     string
	Type         string `mapstructure:"artifact_type"`
	TypeOverride bool   `mapstructure:"artifact_type_override"`
	Metadata     map[string]string

	AwsAccessKey     string `mapstructure:"aws_access_key"`
	AwsSecretKey     string `mapstructure:"aws_secret_key"`
	AwsToken         string `mapstructure:"aws_token"`
	ConsulAddress    string `mapstructure:"consul_address"`
	ConsulScheme     string `mapstructure:"consul_scheme"`
	ConsulToken      string `mapstructure:"consul_token"`

	ProjectName      string `mapstructure:"project_name"`
	ProjectVersion   string `mapstructure:"project_version"`

	// This shouldn't ever be set outside of unit tests.
	Test bool `mapstructure:"test"`

	ctx        interpolate.Context
	user, name string
	buildId    int
}

type PostProcessor struct {
	config Config
	client *api.Client
	auth aws.Auth
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	required := map[string]*string{
		"artifact":      &p.config.Artifact,
		"artifact_type": &p.config.Type,
		"consul_address":      &p.config.ConsulAddress,
		"aws_access_key":      &p.config.AwsAccessKey,
		"aws_secret_key":      &p.config.AwsSecretKey,
		"project_name":        &p.config.ProjectName,
                "project_version":     &p.config.ProjectVersion,
	}

	var errs *packer.MultiError
	for key, ptr := range required {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	p.auth, err = aws.GetAuth(p.config.AwsAccessKey, p.config.AwsSecretKey)
	if err != nil {
		return err
	}

	if p.config.AwsToken != "" {
		p.config.AwsToken = p.auth.Token
	}

  config := api.DefaultConfig()
  config.Address = p.config.ConsulAddress
  //config.Datacenter = parts[0]

  if p.config.ConsulScheme != "" {
    config.Scheme = p.config.ConsulScheme
  }

  if p.config.ConsulToken != "" {
    config.Token = p.config.ConsulToken
  }

  p.client, err = api.NewClient(config)
  if err != nil {
    errs = packer.MultiErrorAppend(
      errs, fmt.Errorf("Error initializing consul client: %s", err))
    return errs
  }

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	_, ok := builtins[artifact.BuilderId()]
	if !ok {
		return nil, false, fmt.Errorf(
			"Unsupported artifact type: %s", artifact.BuilderId())
	}

	ui.Say("Putting build artifacts into consul")

	for _, regions := range strings.Split(artifact.Id(), ",") {
		parts := strings.Split(regions, ":")
		if len(parts) != 2 {
			err := fmt.Errorf("Poorly formatted artifact ID: %s", artifact.Id())
			return nil, false, err
		}

		regionconn := ec2.New(p.auth, aws.Regions[parts[0]])
		ids := []string{parts[1]}
		if images, err := regionconn.Images(ids, nil); err == nil {
			config := api.DefaultConfig()
			config.Address = p.config.ConsulAddress
			config.Datacenter = parts[0]

			if p.config.ConsulScheme != "" {
				config.Scheme = p.config.ConsulScheme
			}

			if p.config.ConsulToken != "" {
				config.Token = p.config.ConsulToken
			}

		        client, err := api.NewClient(config)
		        if err == nil {
				kv := client.KV()
				consul_key_prefix := fmt.Sprintf("amis/%s/%s/%s", p.config.ProjectName, images.Images[0].RootDeviceType, p.config.ProjectVersion)

				ui.Message(fmt.Sprintf("Putting %s image data into consul key prefix %s in datacenter %s",
					parts[1], consul_key_prefix, config.Datacenter))

				consul_data_key := fmt.Sprintf("%s/data", consul_key_prefix)
				ami_data, _ := json.Marshal(images.Images)
				kv_ami_data := &api.KVPair{Key: consul_data_key, Value: ami_data}
				_, err := kv.Put(kv_ami_data, nil)
				if err != nil {
					return artifact, false, err
				}

				consul_ami_key := fmt.Sprintf("%s/ami", consul_key_prefix)
				kv_ami_id := &api.KVPair{Key: consul_ami_key, Value: []byte(parts[1])}
				_, err = kv.Put(kv_ami_id, nil)

				if err != nil {
					return artifact, false, err
				}
			} else {
				return artifact, false, err
		        }
		} else {
			return artifact, false, err
		}
	}

	return artifact, true, nil
}
