// nolint:all
package mqtt

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/xBlaz3kx/DevX/observability"
)

type mqttTestSuite struct {
	suite.Suite
	obs observability.Observability
}

func (suite *mqttTestSuite) SetupTest() {
	suite.obs = observability.NewNoopObservability()
}

func (suite *mqttTestSuite) TestGetIdsFromTopic() {
	expectedIds := []string{"examplePlugin"}
	ids, err := GetIdsFromTopic(suite.obs.Log(), "cmd/examplePlugin/execute", "cmd/+/execute")
	suite.Require().NoError(err)
	suite.Require().Equal(expectedIds, ids)

	ids, err = GetIdsFromTopic(suite.obs.Log(), "cmd/execute", "cmd/+/execute")
	suite.Require().Error(err)

	ids, err = GetIdsFromTopic(suite.obs.Log(), "ploogin/examplePlugin/execute", "cmd/+/execute")
	suite.Require().Error(err)

	ids, err = GetIdsFromTopic(suite.obs.Log(), "ploogin/examplePlugin/execute", "cmd/execute")
	suite.Require().Error(err)

	ids, err = GetIdsFromTopic(suite.obs.Log(), "cmd/examplePlugin/execute", "cmd/examplePlugin/execute")
	suite.Require().Error(err)

	ids, err = GetIdsFromTopic(suite.obs.Log(), "cmd/examplePlugin/execute/example2/abc", "cmd/+/execute/+/abc")
	suite.Require().NoError(err)
	suite.Require().Equal([]string{"examplePlugin", "example2"}, ids)
}

func (suite *mqttTestSuite) TestCreateTopicWithIds() {
	ids, err := CreateTopicWithIds(suite.obs.Log(), "cmd/+/execute", "exampleId")
	suite.Require().NoError(err)
	suite.Require().Equal("cmd/exampleId/execute", ids)

	ids, err = CreateTopicWithIds(suite.obs.Log(), "cmd/+/execute/+/", "exampleId1", "exampleId2")
	suite.Require().NoError(err)
	suite.Require().Equal("cmd/exampleId1/execute/exampleId2/", ids)

	ids, err = CreateTopicWithIds(suite.obs.Log(), "cmd/+/execute/+/", "exampleId")
	suite.Require().Error(err)

	ids, err = CreateTopicWithIds(suite.obs.Log(), "cmd/+/execute/+/", "exampleId", "")
	suite.Require().Error(err)
}

func TestGetIdsFromTopic(t *testing.T) {
	suite.Run(t, new(mqttTestSuite))
}
