/*
 * @Author: Vincent Young
 * @Date: 2023-07-23 19:57:34
 * @LastEditors: Vincent Young
 * @LastEditTime: 2023-07-23 20:16:27
 * @FilePath: /gdeeplx/gdeeplx.go
 * @Telegram: https://t.me/missuo
 *
 * Copyright © 2023 by Vincent, All Rights Reserved.
 */
package gdeeplx

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/abadojack/whatlanggo"
	"github.com/andybalholm/brotli"
	"github.com/tidwall/gjson"
)

type Lang struct {
	SourceLangUserSelected string `json:"source_lang_user_selected"`
	TargetLang             string `json:"target_lang"`
}

type CommonJobParams struct {
	WasSpoken    bool   `json:"wasSpoken"`
	TranscribeAS string `json:"transcribe_as"`
	// RegionalVariant string `json:"regionalVariant"`
}

type Params struct {
	Texts           []Text          `json:"texts"`
	Splitting       string          `json:"splitting"`
	Lang            Lang            `json:"lang"`
	Timestamp       int64           `json:"timestamp"`
	CommonJobParams CommonJobParams `json:"commonJobParams"`
}

type Text struct {
	Text                string `json:"text"`
	RequestAlternatives int    `json:"requestAlternatives"`
}

type PostData struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      int64  `json:"id"`
	Params  Params `json:"params"`
}

func initData(sourceLang string, targetLang string) *PostData {
	return &PostData{
		Jsonrpc: "2.0",
		Method:  "LMT_handle_texts",
		Params: Params{
			Splitting: "newlines",
			Lang: Lang{
				SourceLangUserSelected: sourceLang,
				TargetLang:             targetLang,
			},
			CommonJobParams: CommonJobParams{
				WasSpoken:    false,
				TranscribeAS: "",
				// RegionalVariant: "en-US",
			},
		},
	}
}

func getICount(translateText string) int64 {
	return int64(strings.Count(translateText, "i"))
}

func getRandomNumber() int64 {
	rand.Seed(time.Now().Unix())
	num := rand.Int63n(99999) + 8300000
	return num * 1000
}

func getTimeStamp(iCount int64) int64 {
	ts := time.Now().UnixMilli()
	if iCount != 0 {
		iCount = iCount + 1
		return ts - ts%iCount + iCount
	} else {
		return ts
	}
}

type ResData struct {
	TransText  string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

func Translate(translateText string, sourceLang string, targetLang string, numberAlternative int) (interface{}, error) {
	id := getRandomNumber()

	if sourceLang == "" {
		lang := whatlanggo.DetectLang(translateText)
		deepLLang := strings.ToUpper(lang.Iso6391())
		sourceLang = deepLLang
	}
	if targetLang == "" {
		targetLang = "EN"
	}

	if translateText == "" {
		return map[string]interface{}{
			"message": "No Translate Text Found",
		}, errors.New("No Translate Text Found")
	} else {
		url := "https://www2.deepl.com/jsonrpc"
		id = id + 1
		postData := initData(sourceLang, targetLang)
		text := Text{
			Text:                translateText,
			RequestAlternatives: numberAlternative,
		}
		postData.ID = id
		postData.Params.Texts = append(postData.Params.Texts, text)
		postData.Params.Timestamp = getTimeStamp(getICount(translateText))
		post_byte, _ := json.Marshal(postData)
		postStr := string(post_byte)

		if (id+5)%29 == 0 || (id+3)%13 == 0 {
			postStr = strings.Replace(postStr, "\"method\":\"", "\"method\" : \"", -1)
		} else {
			postStr = strings.Replace(postStr, "\"method\":\"", "\"method\": \"", -1)
		}

		post_byte = []byte(postStr)
		reader := bytes.NewReader(post_byte)
		request, err := http.NewRequest("POST", url, reader)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		// Set Headers
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "*/*")
		request.Header.Set("x-app-os-name", "iOS")
		request.Header.Set("x-app-os-version", "16.3.0")
		request.Header.Set("Accept-Language", "en-US,en;q=0.9")
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")
		request.Header.Set("x-app-device", "iPhone13,2")
		request.Header.Set("User-Agent", "DeepL-iOS/2.9.1 iOS 16.3.0 (iPhone13,2)")
		request.Header.Set("x-app-build", "510265")
		request.Header.Set("x-app-version", "2.9.1")
		request.Header.Set("Connection", "keep-alive")

		client := &http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer resp.Body.Close()

		var bodyReader io.Reader
		switch resp.Header.Get("Content-Encoding") {
		case "br":
			bodyReader = brotli.NewReader(resp.Body)
		default:
			bodyReader = resp.Body
		}

		body, err := io.ReadAll(bodyReader)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		res := gjson.ParseBytes(body)

		if res.Get("error.code").String() == "-32600" {
			log.Println(res.Get("error").String())
			return nil, errors.New("Invalid targetLang")
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, errors.New("Too Many Requests")
		} else {
			var alternatives []string
			res.Get("result.texts.0.alternatives").ForEach(func(key, value gjson.Result) bool {
				alternatives = append(alternatives, value.Get("text").String())
				return true
			})
			return map[string]interface{}{
				"id":           id,
				"data":         res.Get("result.texts.0.text").String(),
				"alternatives": alternatives,
			}, nil
		}
	}
}
