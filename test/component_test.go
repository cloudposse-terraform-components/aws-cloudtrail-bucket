package test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/component-helper"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/assert"
)

type ComponentSuite struct {
	helper.TestSuite
}

func (s *ComponentSuite) TestBasic() {
	const component = "cloudtrail-bucket/basic"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	defer s.DestroyAtmosComponent(s.T(), component, stack, nil)
	options, _ := s.DeployAtmosComponent(s.T(), component, stack, nil)

	bucketId := atmos.Output(s.T(), options, "cloudtrail_bucket_id")
	assert.True(s.T(), strings.HasPrefix(bucketId, "eg-default-ue2-test-cloudtrail-"))

	bucketDomainName := atmos.Output(s.T(), options, "cloudtrail_bucket_domain_name")
	assert.Equal(s.T(), fmt.Sprintf("%s.s3.amazonaws.com", bucketId), bucketDomainName)

	bucketArn := atmos.Output(s.T(), options, "cloudtrail_bucket_arn")
	assert.True(s.T(), strings.HasSuffix(bucketArn, bucketId))

	policy := aws.GetS3BucketPolicy(s.T(), awsRegion, bucketId)

	// Parse bucket policy into map to validate structure
	var policyMap map[string]interface{}
	err := json.Unmarshal([]byte(policy), &policyMap)
	assert.NoError(s.T(), err)

	statements := policyMap["Statement"].([]interface{})
	statement := statements[1].(map[string]interface{})
	principal := statement["Principal"].(map[string]interface{})
	assert.Equal(s.T(), "cloudtrail.amazonaws.com", principal["Service"])

	statement = statements[2].(map[string]interface{})
	principal = statement["Principal"].(map[string]interface{})
	assert.ElementsMatch(s.T(), []string{"cloudtrail.amazonaws.com", "config.amazonaws.com"}, principal["Service"])

	s.DriftTest(component, stack, nil)
}

func (s *ComponentSuite) TestEnabledFlag() {
	const component = "cloudtrail-bucket/disabled"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	s.VerifyEnabledFlag(component, stack, nil)
}


func TestRunSuite(t *testing.T) {
	suite := new(ComponentSuite)
	helper.Run(t, suite)
}
