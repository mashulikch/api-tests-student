package favorites

import (
	"fmt"
	"net/http"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
)

// HttpGetFavorites получает список избранных объявлений
func HttpGetFavorites(t *testing.T, token string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Get(path.FavoritesPath).
		Query("page_size", "50").
		Expect(t).
		End().Response
}

// HttpPostFavorite добавляет объявление в избранное
func HttpPostFavorite(t *testing.T, token, advertisementID string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Post(path.FavoritesPath).
		JSON(map[string]string{"advertisement_id": advertisementID}).
		Expect(t).
		End().Response
}

// HttpDeleteFavorite удаляет объявление из избранного
func HttpDeleteFavorite(t *testing.T, token, advertisementID string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Delete(fmt.Sprintf("%s/%s", path.FavoritesPath, advertisementID)).
		Expect(t).
		End().Response
}
