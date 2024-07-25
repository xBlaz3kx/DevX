package mqtt

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type mqttTestSuite struct {
	suite.Suite
}

func (suite *mqttTestSuite) SetupTest() {
}

func (suite *mqttTestSuite) TestGetIdsFromTopic() {
	expectedIds := []string{"examplePlugin"}
	ids, err := GetIdsFromTopic(nil, "cmd/examplePlugin/execute", "cmd/+/execute")
	suite.Require().NoError(err)
	suite.Require().Equal(expectedIds, ids)

	ids, err = GetIdsFromTopic(nil, "cmd/execute", "cmd/+/execute")
	suite.Require().Error(err)

	ids, err = GetIdsFromTopic(nil, "ploogin/examplePlugin/execute", "cmd/+/execute")
	suite.Require().Error(err)

	ids, err = GetIdsFromTopic(nil, "ploogin/examplePlugin/execute", "cmd/execute")
	suite.Require().Error(err)

	ids, err = GetIdsFromTopic(nil, "cmd/examplePlugin/execute", "cmd/examplePlugin/execute")
	suite.Require().Error(err)

	ids, err = GetIdsFromTopic(nil, "cmd/examplePlugin/execute/example2/abc", "cmd/+/execute/+/abc")
	suite.Require().NoError(err)
	suite.Require().Equal([]string{"examplePlugin", "example2"}, ids)
}

func (suite *mqttTestSuite) TestCreateTopicWithIds() {
	ids, err := CreateTopicWithIds(nil, "cmd/+/execute", "exampleId")
	suite.Require().NoError(err)
	suite.Require().Equal("cmd/exampleId/execute", ids)

	ids, err = CreateTopicWithIds(nil, "cmd/+/execute/+/", "exampleId1", "exampleId2")
	suite.Require().NoError(err)
	suite.Require().Equal("cmd/exampleId1/execute/exampleId2/", ids)

	ids, err = CreateTopicWithIds(nil, "cmd/+/execute/+/", "exampleId")
	suite.Require().Error(err)

	ids, err = CreateTopicWithIds(nil, "cmd/+/execute/+/", "exampleId", "")
	suite.Require().Error(err)
}

func TestGetIdsFromTopic(t *testing.T) {
	suite.Run(t, new(mqttTestSuite))
}
