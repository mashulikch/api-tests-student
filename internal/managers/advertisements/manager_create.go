package advertisements

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	"api-tests-template/internal/managers/advertisements/models"
)

// CreateAdvertisement создает рекламное объявление и проверяет статус ответа
func CreateAdvertisement(t *testing.T, token, title, description, price, quantity, photoPath string, expectedStatus int) (models.CreateAdvertisementResponse, string) {
	response := advertisementsClient.HttpPostAdvertisement(t, token, title, description, price, quantity, photoPath)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equalf(t, expectedStatus, response.StatusCode, "создание объявления вернуло неожиданный ответ: %s", body)

	var created models.CreateAdvertisementResponse
	if response.StatusCode == http.StatusCreated {
		require.NoError(t, json.Unmarshal(body, &created))
	}
	return created, string(body)
}

// GetAdvertisement возвращает рекламное объявление по его идентификатору
func GetAdvertisement(t *testing.T, id string) models.Advertisement {
	response := advertisementsClient.HttpGetAdvertisement(t, id)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equalf(t, http.StatusOK, response.StatusCode, "получение объявления вернуло неожиданный ответ: %s", body)

	var advertisement models.Advertisement
	require.NoError(t, json.Unmarshal(body, &advertisement))
	return advertisement
}

// GetAdvertisementPhotos возвращает загруженные к объявлению фотографии
func GetAdvertisementPhotos(t *testing.T, token, id string) []models.PhotoResponse {
	response := advertisementsClient.HttpGetAdvertisementPhotos(t, token, id)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equalf(t, http.StatusOK, response.StatusCode, "получение фотографий вернуло неожиданный ответ: %s", body)

	var photos []models.PhotoResponse
	require.NoError(t, json.Unmarshal(body, &photos))
	return photos
}

// DeleteAdvertisement удаляет рекламное объявление и проверяет ответ сервера
func DeleteAdvertisement(t *testing.T, token, id string, expectedStatus int) {
	response := advertisementsClient.HttpDeleteAdvertisement(t, token, id)
	defer response.Body.Close()
	require.Equal(t, expectedStatus, response.StatusCode)
}
