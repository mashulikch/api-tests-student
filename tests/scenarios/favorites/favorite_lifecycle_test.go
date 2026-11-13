package favorites

import (
	"net/http"
	"slices"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"api-tests-template/internal/managers/favorites"
	base "api-tests-template/tests"
)

// TestFavoriteLifecycle проверяет добавление, вывод списка, отклонение дубликатов и удаление
func (s *TestSuite) TestFavoriteLifecycle() {
	t := s.T()
	base.Precondition("Находим чужое объявление для проверки полного цикла избранного")
	advertisement := findOtherUsersAdvertisement(t, s.loginData.User.Id, s.loginData.Token)
	initiallyFavorite := slices.Contains(favoriteAdvertisementIDs(favorites.GetFavorites(t, s.loginData.Token)), advertisement.ID)
	favoritePresent := initiallyFavorite
	t.Cleanup(func() {
		if favoritePresent == initiallyFavorite {
			return
		}
		if initiallyFavorite {
			restored := favorites.AddFavorite(t, s.loginData.Token, advertisement.ID)
			defer restored.Body.Close()
			require.Equalf(t, http.StatusCreated, restored.StatusCode, "не удалось восстановить исходное избранное: %s", readBody(t, restored))
			return
		}
		favorites.RemoveFavorite(t, s.loginData.Token, advertisement.ID, http.StatusNoContent)
	})
	if initiallyFavorite {
		base.Precondition("Временно удаляем объявление из избранного, чтобы проверить повторное добавление")
		favorites.RemoveFavorite(t, s.loginData.Token, advertisement.ID, http.StatusNoContent)
		favoritePresent = false
	}

	base.Precondition("Добавляем объявление в избранное")
	response := favorites.AddFavorite(t, s.loginData.Token, advertisement.ID)
	defer response.Body.Close()
	require.Equalf(t, http.StatusCreated, response.StatusCode, "добавление в избранное: %s", readBody(t, response))
	favoritePresent = true

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
		body := readBody(subtestT, duplicate)
		require.Equal(subtestT, "already_exists", gjson.Get(body, "error").String())
		require.Equal(subtestT, alreadyExistsMessage, gjson.Get(body, "message").String())
	})

	base.Precondition("Удаляем объявление из избранного и проверяем список")
	favorites.RemoveFavorite(t, s.loginData.Token, advertisement.ID, http.StatusNoContent)
	favoritePresent = false
	list := favorites.GetFavorites(t, s.loginData.Token)
	require.NotContains(t, favoriteAdvertisementIDs(list), advertisement.ID)
}
