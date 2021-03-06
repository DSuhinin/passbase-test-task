// +build integration

package keys

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/DSuhinin/passbase-test-task/core"
	"github.com/DSuhinin/passbase-test-task/core/errors"

	"github.com/DSuhinin/passbase-test-task/app"
	"github.com/DSuhinin/passbase-test-task/app/api/response"
	"github.com/DSuhinin/passbase-test-task/app/config"
	"github.com/DSuhinin/passbase-test-task/test/fixtures"
)

type CreateKeyTestSuite struct {
	suite.Suite
	fixtures *fixtures.Fixtures
}

// TestCreateKey is an entry point to run all tests in current Test Suite.
func TestCreateKey(t *testing.T) {
	suite.Run(t, new(CreateKeyTestSuite))
}

// SetupSuite prepare everything for tests.
func (s *CreateKeyTestSuite) SetupSuite() {
	// 1. init config.
	appConfig, err := config.New()
	assert.Nil(s.T(), err)

	// 2.initialize db connections.
	dbConnection, err := core.NewDB().GetConnection(
		appConfig.DatabaseUser,
		appConfig.DatabasePass,
		core.PostgresType,
		appConfig.DatabaseName,
		appConfig.DatabaseHost,
	)
	assert.Nil(s.T(), err)

	// 3. init fixtures.
	s.fixtures = fixtures.NewFixtures(dbConnection)
}

// TestCreateKey_OK makes test of `POST /keys` for success case.
func (s *CreateKeyTestSuite) TestCreateKey_OK() {
	defer func() {
		assert.Nil(s.T(), s.fixtures.UnloadFixtures())
	}()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		fmt.Sprintf("%s%s", os.Getenv("SERVICE_BASE_URL"), app.CreateKeyRoute),
		nil,
	)
	req.Header.Set("Authorization", fmt.Sprintf("AdminKey %s", "supersecurekey"))
	assert.Nil(s.T(), err)

	client := http.Client{}
	resp, err := client.Do(req)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), resp.Body)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(s.T(), err)
	// nolint
	defer resp.Body.Close()

	response := response.Key{}
	assert.Nil(s.T(), json.Unmarshal(body, &response))
	assert.Equal(s.T(), 1, response.ID)
	assert.NotEmpty(s.T(), response.Value)

	_, err = uuid.FromString(response.Value)
	assert.Nil(s.T(), err)
}

// TestCreateKey_AuthError makes test of `POST /keys` for Auth error case.
func (s *CreateKeyTestSuite) TestCreateKey_AuthError() {
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		fmt.Sprintf("%s%s", os.Getenv("SERVICE_BASE_URL"), app.CreateKeyRoute),
		nil,
	)
	req.Header.Set("Authorization", "AdminKey incorrectkey")
	assert.Nil(s.T(), err)

	client := http.Client{}
	resp, err := client.Do(req)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), resp.Body)
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(s.T(), err)
	// nolint
	defer resp.Body.Close()

	response := errors.HTTP{}
	assert.Nil(s.T(), json.Unmarshal(body, &response))
	assert.Equal(s.T(), 100001, response.Code)
	assert.Equal(s.T(), "unauthorized request", response.Message)
}
