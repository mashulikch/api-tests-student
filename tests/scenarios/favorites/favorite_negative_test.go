package favorites

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/favorites"
	"api-tests-template/internal/utils"
	"api-tests-template/tests/helpers"
)

// TestFavoriteNegative проверяет отсутствие ресурсов, ограничение на показ собственной рекламы и авторизацию
func (s *TestSuite) TestFavoriteNegative() {
	t := s.T()
	missingID := "00000000-0000-4000-8000-000000000000"

	s.Run("Добавление несуществующего объявления возвращает 404", func() {
		response := favorites.AddFavorite(t, s.loginData.Token, missingID)
		defer response.Body.Close()
		body := readBody(t, response)
		require.Equal(t, http.StatusNotFound, response.StatusCode)
		require.Equal(t, favoriteErrorNotFound, gjson.Get(body, "error").String())
		require.Equal(t, missingAdvertisementMessage, gjson.Get(body, "message").String())
	})

	s.Run("Удаление несуществующего объявления возвращает 404", func() {
		response := favorites.RemoveFavoriteResponse(t, s.loginData.Token, missingID)
		defer response.Body.Close()
		body := readBody(t, response)
		require.Equal(t, http.StatusNotFound, response.StatusCode)
		require.Equal(t, favoriteErrorNotFound, gjson.Get(body, "error").String())
		require.Equal(t, missingFavoriteMessage, gjson.Get(body, "message").String())
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
		require.Equal(t, favoriteErrorForbidden, gjson.Get(body, "error").String())
		require.Equal(t, ownAdvertisementMessage, gjson.Get(body, "message").String())
	})

	for _, operation := range []string{"GET", "POST", "DELETE"} {
		for _, token := range []string{"invalid_token", ""} {
			authorizationLabel := "неверным токеном"
			if token == "" {
				authorizationLabel = "без токена"
			}
			s.Run(fmt.Sprintf("%s %s возвращает 401", operation, authorizationLabel), func() {
				var response *http.Response
				expectedMessage := invalidAuthMessage
				if token == "" {
					expectedMessage = missingAuthorizationHeaderMessage
				}
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
				require.Equal(t, favoriteErrorUnauthorized, gjson.Get(body, "error").String())
				require.Equal(t, expectedMessage, gjson.Get(body, "message").String())
			})
		}
	}
}
