package provider

import (
	"encoding/json"
	"net/http"

	"github.com/sevigo/shugosha/pkg/model"
)

// NewProviderInfoHandler returns an HTTP handler function that uses ProviderMetaInfoGetter.
func NewProviderInfoHandler(getter model.ProviderMetaInfoGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providers, err := getter.GetProviders()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		providerInfos := make(map[string]*model.ProviderMetaInfo)
		for _, providerName := range providers {
			metaInfo, err := getter.GetMetaInfo(providerName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			providerInfos[providerName] = metaInfo
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(providerInfos)
	}
}
