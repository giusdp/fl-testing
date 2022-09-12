package tests

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/funlessdev/fl-cli/pkg/deploy"
	swagger "github.com/funlessdev/fl-client-sdk-go"
	"github.com/funlessdev/fl-testing/e2e-tests/internal/cli"
	"github.com/funlessdev/fl-testing/e2e-tests/internal/sdk"
	"github.com/stretchr/testify/suite"
)

type SDKTestSuite struct {
	suite.Suite
	ctx         context.Context
	deployer    deploy.DockerDeployer
	fnName      string
	fnNamespace string
	fnCode      string
	fnImage     string
	fnHost      string
	fnArgs      interface{}
	fnClient    *swagger.APIClient
}

func (suite *SDKTestSuite) SetupSuite() {
	host := os.Getenv("FL_TEST_HOST")
	if host == "" {
		suite.T().Skip("set FL_TEST_HOST to run this test")
	}
	suite.ctx = context.Background()
	suite.fnName = "hellojs"
	suite.fnNamespace = "helloNS"
	suite.fnImage = "nodejs"
	suite.fnArgs = map[string]string{"name": "Test"}
	source, err := os.ReadFile("../functions/hello.js")

	if err != nil {
		suite.T().Errorf("Error while reading source code file: %+v\n", err)
	}

	suite.fnCode = string(source)
	suite.fnHost = host

	suite.fnClient = sdk.BuildClient(suite.fnHost)

	deployer, err := cli.NewDeployer(suite.ctx)
	if err != nil {
		suite.T().Errorf("Error during docker deployer creation: %+v\n", err)
	}

	suite.deployer = deployer
	_ = cli.DeployDev(suite.ctx, suite.deployer)

	//wait for everything to be up
	time.Sleep(1 * time.Second)
}

func (suite *SDKTestSuite) TearDownSuite() {
	_ = cli.DestroyDev(suite.ctx, suite.deployer)
}

func (suite *SDKTestSuite) TestInvocationSuccess() {
	// create function
	suite.Run("should successfully create function", func() {
		result, _, err := suite.fnClient.DefaultApi.CreatePost(suite.ctx, swagger.FunctionCreation{
			Name:      suite.fnName,
			Namespace: suite.fnNamespace,
			Code:      suite.fnCode,
			Image:     suite.fnImage,
		})
		suite.NoError(err)
		suite.Equal(suite.fnName, result.Result)
	})
	// invoke function
	suite.Run("should return no error when invoking an existing function", func() {
		_, _, err := suite.fnClient.DefaultApi.InvokePost(suite.ctx, swagger.FunctionInvocation{
			Function:  suite.fnName,
			Namespace: suite.fnNamespace,
			Args:      &suite.fnArgs,
		})

		suite.NoError(err)
	})
	suite.Run("should return the correct result when invoking hellojs with no args", func() {
		result, _, err := suite.fnClient.DefaultApi.InvokePost(suite.ctx, swagger.FunctionInvocation{
			Function:  suite.fnName,
			Namespace: suite.fnNamespace,
			Args:      &suite.fnArgs,
		})
		name := suite.fnArgs.(map[string]string)["name"]
		decodedResult, jErr := json.Marshal(*result.Result)

		suite.NoError(err)
		suite.NoError(jErr)
		suite.Equal("{\"payload\":\"Hello "+name+"!\"}", string(decodedResult))

	})
	suite.Run("should return the correct result when invoking hellojs with args", func() {
		result, _, err := suite.fnClient.DefaultApi.InvokePost(suite.ctx, swagger.FunctionInvocation{
			Function:  suite.fnName,
			Namespace: suite.fnNamespace,
		})
		decodedResult, jErr := json.Marshal(*result.Result)

		suite.NoError(err)
		suite.NoError(jErr)
		suite.Equal("{\"payload\":\"Hello World!\"}", string(decodedResult))
	})

	//delete function
	suite.Run("should successfully delete function", func() {
		result, _, err := suite.fnClient.DefaultApi.DeletePost(suite.ctx, swagger.FunctionDeletion{
			Name:      suite.fnName,
			Namespace: suite.fnNamespace,
		})
		suite.NoError(err)
		suite.Equal(suite.fnName, result.Result)
	})
}

func (suite *SDKTestSuite) TestInvocationFailure() {

}

func TestSDKSuite(t *testing.T) {
	suite.Run(t, new(SDKTestSuite))
}
