package favorites

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	favoritesClient "api-tests-template/internal/client/http/favorites"
	"api-tests-template/internal/managers/favorites/models"
)

// GetFavorites возвращает избранные объявления текущего пользователя
func GetFavorites(t *testing.T, token string) models.FavoriteList {
	response := favoritesClient.HttpGetFavorites(t, token)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equalf(t, http.StatusOK, response.StatusCode, "получение избранного вернуло неожиданный ответ: %s", body)

	var favorites models.FavoriteList
	require.NoError(t, json.Unmarshal(body, &favorites))
	return favorites
}

// AddFavorite отправляет запрос на добавление объявления в избранное
func AddFavorite(t *testing.T, token, advertisementID string) *http.Response {
	return favoritesClient.HttpPostFavorite(t, token, advertisementID)
}

// GetFavoritesResponse отправляет запрос на получение списка,
// а проверку статуса оставляет на усмотрение вызывающей стороны
func GetFavoritesResponse(t *testing.T, token string) *http.Response {
	return favoritesClient.HttpGetFavorites(t, token)
}

// RemoveFavorite удаляет объявление из избранного и проверяет статус ответа
func RemoveFavorite(t *testing.T, token, advertisementID string, expectedStatus int) {
	response := favoritesClient.HttpDeleteFavorite(t, token, advertisementID)
	defer response.Body.Close()
	require.Equal(t, expectedStatus, response.StatusCode)
}

// RemoveFavoriteResponse отправляет запрос на удаление,
// а проверку статуса оставляет на усмотрение вызывающей стороны
func RemoveFavoriteResponse(t *testing.T, token, advertisementID string) *http.Response {
	return favoritesClient.HttpDeleteFavorite(t, token, advertisementID)
}
