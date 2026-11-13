package favorites

import (
	"net/http"
	"os"

	"github.com/stretchr/testify/require"

	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/favorites"
	"api-tests-template/internal/utils"
	base "api-tests-template/tests"
	"api-tests-template/tests/helpers"
)

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

	base.Precondition("Удаляем объявление владельца и проверяем очистку избранного")
	advertisements.DeleteAdvertisement(t, owner.Token, created.ID, http.StatusNoContent)
	deleted = true
	require.NotContains(t, favoriteAdvertisementIDs(favorites.GetFavorites(t, s.loginData.Token)), created.ID,
		"после удаления объявления оно должно исчезнуть из избранного")
}
