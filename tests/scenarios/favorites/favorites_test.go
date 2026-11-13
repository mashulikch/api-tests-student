package favorites

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"
	base "api-tests-template/tests"
)

const (
	favoriteAdDescription = "Объявление для проверки избранного"
	favoriteAdPrice       = "99"
	favoriteAdQuantity    = "1"

	favoriteErrorForbidden    = "forbidden"
	favoriteErrorNotFound     = "not_found"
	favoriteErrorUnauthorized = "unauthorized"

	missingAdvertisementMessage       = "Advertisement not found"
	missingFavoriteMessage            = "Favorite entry not found"
	ownAdvertisementMessage           = "Cannot add own advertisement to favorites"
	alreadyExistsMessage              = "Advertisement is already in favorites"
	invalidAuthMessage                = "Invalid or expired token"
	missingAuthorizationHeaderMessage = "Authorization header required"
)

type TestSuite struct {
	suite.Suite
	loginData models.LoginOkResponse
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (s *TestSuite) SetupSuite() {
	base.SetupSuite()
	s.loginData = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
}
