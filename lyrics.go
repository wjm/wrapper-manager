package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
)

func GetInstanceAuthToken(instance *WrapperInstance) (string, string, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("data/wrapper/rootfs/data/instances/%s/mpl_db/cookies.sqlitedb", instance.Id))
	if err != nil {
		return "", "", err
	}
	defer db.Close()
	dsidQuery := "select value from cookies where name == \"X-Dsid\";"
	result, err := db.Query(dsidQuery)
	if err != nil {
		return "", "", err
	}
	var dsid string
	for result.Next() {
		err = result.Scan(&dsid)
		if err != nil {
			return "", "", err
		}
	}
	tokenQuery := fmt.Sprintf("select value from cookies where name == \"mz_at_ssl-%s\";", dsid)
	result, err = db.Query(tokenQuery)
	if err != nil {
		return "", "", err
	}
	var token string
	for result.Next() {
		err = result.Scan(&token)
		if err != nil {
			return "", "", err
		}
	}
	return dsid, token, nil
}

func GetLyrics(adamID string, region string, language string, dsid string, token string, accessToken string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://amp-api.music.apple.com/v1/catalog/%s/songs/%s/lyrics?l=%s", region, adamID, language), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Music/5.7 Android/10 model/Pixel6GR1YH build/1234 (dt:66)")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Dsid", dsid)
	req.Header.Set("Origin", "https://music.apple.com")
	req.AddCookie(&http.Cookie{
		Name:  fmt.Sprintf("mz_at_ssl-%s", dsid),
		Value: accessToken,
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New("no available lyrics")
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var respJson map[string][]interface{}
	if err := json.Unmarshal(respBody, &respJson); err != nil {
		return "", err
	}
	ttml := respJson["data"][0].(map[string]interface{})["attributes"].(map[string]interface{})["ttml"].(string)
	return ttml, nil
}
