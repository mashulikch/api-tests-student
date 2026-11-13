package favorites

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"

	"api-tests-template/internal/managers/advertisements"
	advertisementModels "api-tests-template/internal/managers/advertisements/models"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/managers/favorites"
	favoritesModels "api-tests-template/internal/managers/favorites/models"
	"api-tests-template/internal/utils"
	base "api-tests-template/tests"
	"api-tests-template/tests/helpers"
)

const (
	favoriteAdDescription = "Объявление для проверки избранного"
	favoriteAdPrice       = "99"
	favoriteAdQuantity    = "1"
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

// TestFavoriteLifecycle проверяет добавление, вывод списка, отклонение дубликатов и удаление
func (s *TestSuite) TestFavoriteLifecycle() {
	t := s.T()
	base.Precondition("Находим чужое объявление для проверки полного цикла избранного")
	advertisement := findOtherUsersAdvertisement(t, s.loginData.User.Id, s.loginData.Token)

	base.Precondition("Добавляем объявление в избранное")
	response := favorites.AddFavorite(t, s.loginData.Token, advertisement.ID)
	require.Equalf(t, http.StatusCreated, response.StatusCode, "добавление в избранное: %s", readBody(t, response))
	removed := false
	t.Cleanup(func() {
		if !removed {
			favorites.RemoveFavorite(t, s.loginData.Token, advertisement.ID, http.StatusNoContent)
		}
	})

	s.Run("Объявление появляется в избранном", func() {
		subtestT := s.T()
		list := favorites.GetFavorites(subtestT, s.loginData.Token)
		require.Contains(subtestT, favoriteAdvertisementIDs(list), advertisement.ID)
	})

	s.Run("Повторное добавление возвращает 409", func() {
		subtestT := s.T()
		duplicate := favorites.AddFavorite(subtestT, s.loginData.Token, advertisement.ID)
		defer duplicate.Body.Close()
		require.Equal(subtestT, http.StatusConflict, duplicate.StatusCode)
		require.Equal(subtestT, "already_exists", gjson.Get(readBody(subtestT, duplicate), "error").String())
	})

	base.Precondition("Удаляем объявление из избранного и проверяем список")
	favorites.RemoveFavorite(t, s.loginData.Token, advertisement.ID, http.StatusNoContent)
	removed = true
	list := favorites.GetFavorites(t, s.loginData.Token)
	require.NotContains(t, favoriteAdvertisementIDs(list), advertisement.ID)
}

// TestFavoriteNegative проверяет отсутствие ресурсов, ограничение на показ собственной рекламы и авторизацию
func (s *TestSuite) TestFavoriteNegative() {
	t := s.T()
	missingID := "00000000-0000-4000-8000-000000000000"

	s.Run("Добавление несуществующего объявления возвращает 404", func() {
		response := favorites.AddFavorite(t, s.loginData.Token, missingID)
		defer response.Body.Close()
		body := readBody(t, response)
		require.Equal(t, http.StatusNotFound, response.StatusCode)
		require.NotEmpty(t, gjson.Get(body, "message").String())
	})

	s.Run("Удаление несуществующего объявления возвращает 404", func() {
		response := favorites.RemoveFavoriteResponse(t, s.loginData.Token, missingID)
		defer response.Body.Close()
		body := readBody(t, response)
		require.Equal(t, http.StatusNotFound, response.StatusCode)
		require.NotEmpty(t, gjson.Get(body, "message").String())
	})

	s.Run("Своё объявление нельзя добавить в избранное", func() {
		title := "api-test-own-" + utils.RandomString(10)
		created, _ := advertisements.CreateAdvertisement(t, s.loginData.Token, title, favoriteAdDescription,
			favoriteAdPrice, favoriteAdQuantity, helpers.CreatePhotoFile(t), http.StatusCreated)
		t.Cleanup(func() { advertisements.DeleteAdvertisement(t, s.loginData.Token, created.ID, http.StatusNoContent) })

		response := favorites.AddFavorite(t, s.loginData.Token, created.ID)
		defer response.Body.Close()
		body := readBody(t, response)
		require.Equal(t, http.StatusForbidden, response.StatusCode)
		require.NotEmpty(t, gjson.Get(body, "message").String())
	})

	for _, operation := range []string{"GET", "POST", "DELETE"} {
		for _, token := range []string{"invalid_token", ""} {
			authorizationLabel := "неверным токеном"
			if token == "" {
				authorizationLabel = "без токена"
			}
			s.Run(fmt.Sprintf("%s %s возвращает 401", operation, authorizationLabel), func() {
				var response *http.Response
				switch operation {
				case "GET":
					response = favorites.GetFavoritesResponse(t, token)
				case "POST":
					response = favorites.AddFavorite(t, token, missingID)
				case "DELETE":
					response = favorites.RemoveFavoriteResponse(t, token, missingID)
				}
				defer response.Body.Close()
				body := readBody(t, response)
				require.Equal(t, http.StatusUnauthorized, response.StatusCode)
				require.NotEmpty(t, gjson.Get(body, "message").String())
			})
		}
	}
}

// TestDeletingAdvertisementRemovesFavorite проверяет очистку избранного для двух пользователей
func (s *TestSuite) TestDeletingAdvertisementRemovesFavorite() {
	ownerLogin := os.Getenv("TEST_OTHER_LOGIN")
	ownerPassword := os.Getenv("TEST_OTHER_PASSWORD")
	if ownerLogin == "" || ownerPassword == "" {
		s.T().Skip("для проверки удаления чужого объявления задайте TEST_OTHER_LOGIN и TEST_OTHER_PASSWORD")
	}

	t := s.T()
	owner := auth.Login(t, ownerLogin, ownerPassword)
	title := "api-test-delete-favorite-" + utils.RandomString(10)
	created, _ := advertisements.CreateAdvertisement(t, owner.Token, title, favoriteAdDescription,
		favoriteAdPrice, favoriteAdQuantity, helpers.CreatePhotoFile(t), http.StatusCreated)
	deleted := false
	addedToFavorites := false
	t.Cleanup(func() {
		if addedToFavorites && !deleted {
			favorites.RemoveFavorite(t, s.loginData.Token, created.ID, http.StatusNoContent)
		}
		if !deleted {
			advertisements.DeleteAdvertisement(t, owner.Token, created.ID, http.StatusNoContent)
		}
	})

	added := favorites.AddFavorite(t, s.loginData.Token, created.ID)
	defer added.Body.Close()
	require.Equalf(t, http.StatusCreated, added.StatusCode, "добавление тестового объявления в избранное: %s", readBody(t, added))
	addedToFavorites = true

	advertisements.DeleteAdvertisement(t, owner.Token, created.ID, http.StatusNoContent)
	deleted = true
	require.NotContains(t, favoriteAdvertisementIDs(favorites.GetFavorites(t, s.loginData.Token)), created.ID,
		"после удаления объявления оно должно исчезнуть из избранного")
}

// findOtherUsersAdvertisement выбирает существующее объявление, которое текущий пользователь может добавить в избранное
func findOtherUsersAdvertisement(t *testing.T, userID, token string) advertisementModels.Advertisement {
	t.Helper()
	list := advertisements.GetAdvertisementsList(t, token, "")
	for _, advertisement := range list.Items {
		if advertisement.UserID != "" && advertisement.UserID != userID {
			return advertisement
		}
	}
	t.Fatalf("в каталоге нет чужого объявления для проверки избранного")
	return advertisementModels.Advertisement{}
}

// favoriteAdvertisementIDs возвращает идентификаторы объявлений из списка избранного
func favoriteAdvertisementIDs(list favoritesModels.FavoriteList) []string {
	ids := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		ids = append(ids, item.Advertisement.ID)
	}
	return ids
}

// readBody считывает и закрывает тело HTTP-ответа, чтобы обеспечить получение понятных сообщений об ошибках
func readBody(t *testing.T, response *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	return string(body)
}
