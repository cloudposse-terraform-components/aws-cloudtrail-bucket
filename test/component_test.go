package test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/aws-component-helper"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/assert"
)

func TestComponent(t *testing.T) {
	awsRegion := "us-east-2"

	fixture := helper.NewFixture(t, "../", awsRegion, "test/fixtures")

	defer fixture.TearDown()
	fixture.SetUp(&atmos.Options{})

	fixture.Suite("default", func(t *testing.T, suite *helper.Suite) {
		suite.Test(t, "basic", func(t *testing.T, atm *helper.Atmos) {

			defer atm.GetAndDestroy("cloudtrail-bucket/basic", "default-test", map[string]interface{}{})
			component := atm.GetAndDeploy("cloudtrail-bucket/basic", "default-test", map[string]interface{}{})

			bucketId := atm.Output(component, "cloudtrail_bucket_id")
			assert.True(t, strings.HasPrefix(bucketId, "eg-default-ue2-test-cloudtrail-"))

			bucketDomainName := atm.Output(component, "cloudtrail_bucket_domain_name")
			assert.Equal(t, fmt.Sprintf("%s.s3.amazonaws.com", bucketId), bucketDomainName)

			bucketArn := atm.Output(component, "cloudtrail_bucket_arn")
			assert.True(t, strings.HasSuffix(bucketArn, bucketId))

			policy := aws.GetS3BucketPolicy(t, awsRegion, bucketId)

			// Parse bucket policy into map to validate structure
			var policyMap map[string]interface{}
			err := json.Unmarshal([]byte(policy), &policyMap)
			assert.NoError(t, err)

			statements := policyMap["Statement"].([]interface{})
			statement := statements[1].(map[string]interface{})
			principal := statement["Principal"].(map[string]interface{})
			assert.Equal(t, "cloudtrail.amazonaws.com", principal["Service"])

			statement = statements[2].(map[string]interface{})
			principal = statement["Principal"].(map[string]interface{})
			assert.ElementsMatch(t, []string{"cloudtrail.amazonaws.com", "config.amazonaws.com"}, principal["Service"])
		})
	})
}
