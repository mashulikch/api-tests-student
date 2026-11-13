package models

import advertisementModels "api-tests-template/internal/managers/advertisements/models"

type FavoriteList struct {
	Items []FavoriteItem `json:"items"`
	Total int            `json:"total"`
}

type FavoriteItem struct {
	FavoriteID    string                            `json:"favorite_id"`
	Advertisement advertisementModels.Advertisement `json:"advertisement"`
}
