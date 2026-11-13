package advertisementCreate

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"

	"api-tests-template/internal/assertions"
	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/utils"
	base "api-tests-template/tests"
	"api-tests-template/tests/helpers"
)

const (
	advertisementDescription = "Объявление для API-тестов"
	advertisementPrice       = "1234.5"
	advertisementQuantity    = "7"
	maxAdvertisementTitle    = 255
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

// TestCreateAdvertisementPositive проверяет создание, получение, поиск и загрузку фотографий
func (s *TestSuite) TestCreateAdvertisementPositive() {
	t := s.T()
	title := "api-test-" + utils.RandomString(12)
	photoPath := helpers.CreatePhotoFile(t)
	base.Precondition("Создаем объявление со всеми параметрами и фотографией")

	s.Run("Создание и проверка объявления", func() {
		created, _ := advertisements.CreateAdvertisement(t, s.loginData.Token, title, advertisementDescription,
			advertisementPrice, advertisementQuantity, photoPath, http.StatusCreated)
		require.NotEmpty(t, created.ID, "API не вернул id созданного объявления")
		t.Cleanup(func() { advertisements.DeleteAdvertisement(t, s.loginData.Token, created.ID, http.StatusNoContent) })

		s.Run("GET возвращает переданные параметры", func() {
			subtestT := s.T()
			actual := advertisements.GetAdvertisement(subtestT, created.ID)
			require.Equal(subtestT, title, actual.Title)
			require.Equal(subtestT, advertisementDescription, actual.Description)
			require.Equal(subtestT, 1234.5, actual.Price)
			require.Equal(subtestT, 7, actual.Quantity)
			require.Equal(subtestT, s.loginData.User.Id, actual.UserID)
		})

		s.Run("Объявление находится полным поисковым запросом", func() {
			subtestT := s.T()
			found := advertisements.GetAdvertisementsList(subtestT, s.loginData.Token, title)
			assertions.ContainsAdvertisementByTitle(subtestT, found.Items, title)
		})

		s.Run("Фотография доступна через ручку объявления", func() {
			subtestT := s.T()
			photos := advertisements.GetAdvertisementPhotos(subtestT, s.loginData.Token, created.ID)
			require.Len(subtestT, photos, 1)
			require.NotEmpty(subtestT, photos[0].ID)
			require.NotEmpty(subtestT, photos[0].URL)
		})
	})
}

// TestCreateAdvertisementNegative проверяет некорректные поля и отсутствие авторизации
func (s *TestSuite) TestCreateAdvertisementNegative() {
	photoPath := helpers.CreatePhotoFile(s.T())
	testCases := []struct {
		name   string
		token  string
		title  string
		price  string
		status int
	}{
		{name: "пустой title", token: s.loginData.Token, title: "", price: advertisementPrice, status: http.StatusBadRequest},
		{name: "отрицательная цена", token: s.loginData.Token, title: "api-test-" + utils.RandomString(8), price: "-1", status: http.StatusBadRequest},
		{name: "слишком длинный title", token: s.loginData.Token, title: strings.Repeat("a", maxAdvertisementTitle+1), price: advertisementPrice, status: http.StatusBadRequest},
		{name: "неверный токен", token: "invalid_token", title: "api-test-" + utils.RandomString(8), price: advertisementPrice, status: http.StatusUnauthorized},
		{name: "отсутствующий токен", token: "", title: "api-test-" + utils.RandomString(8), price: advertisementPrice, status: http.StatusUnauthorized},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			base.Precondition(fmt.Sprintf("Проверяем отказ создания объявления: %s", testCase.name))
			_, body := advertisements.CreateAdvertisement(s.T(), testCase.token, testCase.title, advertisementDescription,
				testCase.price, advertisementQuantity, photoPath, testCase.status)
			require.NotEmpty(s.T(), gjson.Get(body, "error").String(), "в ответе отсутствует код ошибки")
			require.NotEmpty(s.T(), gjson.Get(body, "message").String(), "в ответе отсутствует сообщение об ошибке")
		})
	}
}
