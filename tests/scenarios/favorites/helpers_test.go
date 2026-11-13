package favorites

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/advertisements/models"
	favoritesModels "api-tests-template/internal/managers/favorites/models"
)

// findOtherUsersAdvertisement выбирает существующее объявление, которое текущий пользователь может добавить в избранное
func findOtherUsersAdvertisement(t *testing.T, userID, token string) models.Advertisement {
	t.Helper()
	list := advertisements.GetAdvertisementsList(t, token, "")
	for _, advertisement := range list.Items {
		if advertisement.UserID != "" && advertisement.UserID != userID {
			return advertisement
		}
	}
	t.Fatalf("в каталоге нет чужого объявления для проверки избранного")
	return models.Advertisement{}
}

// favoriteAdvertisementIDs возвращает идентификаторы объявлений из списка избранного
func favoriteAdvertisementIDs(list favoritesModels.FavoriteList) []string {
	ids := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		ids = append(ids, item.Advertisement.ID)
	}
	return ids
}

// readBody считывает тело HTTP-ответа и логирует его при ошибочном статусе
func readBody(t *testing.T, response *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	if response.StatusCode >= http.StatusBadRequest {
		t.Logf("Ответ API: status=%d body=%s", response.StatusCode, body)
	}
	return string(body)
}
