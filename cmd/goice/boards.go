package main

import (
	"encoding/json"
	"os"
	"path"
)

func readJson(path string, obj interface{}) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	json.NewDecoder(jsonFile).Decode(obj)
	return nil
}

const (
	JSON_MENU   = "menu.json"
	JSON_INFO   = "info.json"
	JSON_PINOUT = "pinout.json"
	JSON_RULES  = "rules.json"
)

func LoadBoards(boardsPath string) (map[string]*Board, error) {
	var sections []Section
	boards := make(map[string]*Board)
	menuPath := path.Join(boardsPath, JSON_MENU)

	if err := readJson(menuPath, &sections); err != nil {
		return nil, err
	}

	for _, FPGAfamily := range sections {
		for _, boardName := range FPGAfamily.Boards {
			board := Board{
				Name: boardName,
				Type: FPGAfamily.Type,
			}

			if err := readJson(path.Join(boardsPath, boardName, JSON_INFO), &board.Info); err != nil {
				continue
			}
			if err := readJson(path.Join(boardsPath, boardName, JSON_PINOUT), &board.Pinouts); err != nil {
				continue
			}
			if err := readJson(path.Join(boardsPath, boardName, JSON_RULES), &board.Rules); err != nil {
				continue
			}

			boards[boardName] = &board
		}
	}

	return boards, nil
}
