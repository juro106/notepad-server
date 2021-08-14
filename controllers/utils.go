package controllers

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"

	"notepad/models"

	_ "github.com/go-sql-driver/mysql"
)

func makeImageDir(dir string) {
	dirname := "./images/" + dir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dirname, 0777)
	}
}

func time2int(arg string) int {
	var i int
	time := strings.Replace(arg, "-", "", -1)
	time = strings.Replace(time, ":", "", -1)
	time = strings.Replace(time, ".", "", -1)
	time = strings.Replace(time, " ", "", -1)
	i, _ = strconv.Atoi(time)
	return i
}

// 汎用データ取得
type JsonObject map[string]interface{}

func (j *JsonObject) Scan(src interface{}) error {
	var _src []byte
	switch src.(type) {
	case []byte:
		_src = src.([]byte)
	default:
		return errors.New("failed to scan JsonObject")
	}
	if err := json.NewDecoder(bytes.NewReader(_src)).Decode(j); err != nil {
		return err
	}
	return nil
}

func (j JsonObject) Value() (driver.Value, error) {
	b := make([]byte, 0)
	buf := bytes.NewBuffer(b)
	if err := json.NewEncoder(buf).Encode(j); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// コンテンツデータ取得用
type ContentObject models.Content

func (content *ContentObject) Scan(src interface{}) error {
	var _src []byte
	switch src.(type) {
	case []byte:
		_src = src.([]byte)
	default:
		return errors.New("failed to scan JsonObject")
	}
	if err := json.NewDecoder(bytes.NewReader(_src)).Decode(content); err != nil {
		return err
	}
	return nil
}

// タグリスト取得用。構造体を定義したほうが後の加工がわかりやすい。（これしか知らない）
type TagsObject models.Tags

func (t *TagsObject) Scan(src interface{}) error {
	var _src []byte
	switch src.(type) {
	case []byte:
		_src = src.([]byte)
	default:
		return errors.New("failed to scan JsonObject")
	}
	if err := json.NewDecoder(bytes.NewReader(_src)).Decode(t); err != nil {
		return err
	}
	return nil
}

type TagNumObject struct {
	Name   string `json:"name"`
	Number int    `json:"number"`
}

func (t *TagNumObject) Scan(src interface{}) error {
	var _src []byte
	switch src.(type) {
	case []byte:
		_src = src.([]byte)
	default:
		return errors.New("failed to scan JsonObject")
	}
	if err := json.NewDecoder(bytes.NewReader(_src)).Decode(t); err != nil {
		return err
	}
	return nil
}

// 現在のユーザー自身のプロジェクト以外は見られないようにするためのチェック
func checkUserProjects(project string, projects []string) bool {
	var isProject bool
	for _, v := range projects {
		if project == v {
			isProject = true
			return isProject
		}
	}
	isProject = false
	return isProject
}
