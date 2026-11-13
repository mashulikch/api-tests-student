package advertisementCreate

import (
	"net/http"

	"github.com/stretchr/testify/require"

	"api-tests-template/internal/assertions"
	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/utils"
	base "api-tests-template/tests"
	"api-tests-template/tests/helpers"
)

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
