package advertisementCreate

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/utils"
	base "api-tests-template/tests"
	"api-tests-template/tests/helpers"
)

// TestCreateAdvertisementNegative проверяет некорректные поля и отсутствие авторизации
func (s *TestSuite) TestCreateAdvertisementNegative() {
	photoPath := helpers.CreatePhotoFile(s.T())
	testCases := []struct {
		name            string
		token           string
		title           string
		price           string
		status          int
		expectedError   string
		expectedMessage string
	}{
		{
			name: "пустой title", token: s.loginData.Token, title: "", price: advertisementPrice,
			status: http.StatusBadRequest, expectedError: "invalid_request", expectedMessage: "title is required",
		},
		{
			name: "отрицательная цена", token: s.loginData.Token, title: "api-test-" + utils.RandomString(8), price: "-1",
			status: http.StatusBadRequest, expectedError: "invalid_request",
			expectedMessage: "price must be between 0 and 1000000 inclusive",
		},
		{
			name: "слишком длинный title", token: s.loginData.Token,
			title: strings.Repeat("a", maxAdvertisementTitle+1), price: advertisementPrice,
			status: http.StatusBadRequest, expectedError: "invalid_request",
			expectedMessage: fmt.Sprintf("title must not exceed %d characters provided %d", maxAdvertisementTitle, maxAdvertisementTitle+1),
		},
		{
			name: "неверный токен", token: "invalid_token", title: "api-test-" + utils.RandomString(8),
			price: advertisementPrice, status: http.StatusUnauthorized, expectedError: "unauthorized",
			expectedMessage: "Invalid or expired token",
		},
		{
			name: "отсутствующий токен", token: "", title: "api-test-" + utils.RandomString(8),
			price: advertisementPrice, status: http.StatusUnauthorized, expectedError: "unauthorized",
			expectedMessage: "Authorization header required",
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			base.Precondition(fmt.Sprintf("Проверяем отказ создания объявления: %s", testCase.name))
			_, body := advertisements.CreateAdvertisement(s.T(), testCase.token, testCase.title, advertisementDescription,
				testCase.price, advertisementQuantity, photoPath, testCase.status)
			require.Equal(s.T(), testCase.expectedError, gjson.Get(body, "error").String(), "неожиданный код ошибки")
			require.Equal(s.T(), testCase.expectedMessage, gjson.Get(body, "message").String(), "неожиданное сообщение об ошибке")
		})
	}
}
