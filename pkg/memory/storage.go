package memory

import (
	"encoding/json"
	"os"
	"ozon-graphql-api/graph/model"
	"sync"
)

type Storage struct {
	Mu       sync.RWMutex
	Posts    map[string]*model.Post
	Comments map[string]*model.Comment
	Users    map[string]*model.User

	/*
		В базе данных мы используем автоинкременту для каждой из сущностей
		Чтобы не уходить от этой логики, используем этот же подход через счетчики в структуре
	*/
	PostIdCounter    int
	CommentIdCounter int
	UserIdCounter    int
}

func NewStorage() *Storage {
	storage := &Storage{
		Posts:            make(map[string]*model.Post),
		Comments:         make(map[string]*model.Comment),
		Users:            make(map[string]*model.User),
		PostIdCounter:    0,
		CommentIdCounter: 0,
		UserIdCounter:    3,
	}

	//В нашей системе нет ручек, которые создают юзеров, поэтому добавим несколько в ручную.
	user1 := &model.User{
		ID:       "1",
		Username: "Maxim",
	}
	user2 := &model.User{
		ID:       "2",
		Username: "Vika",
	}
	user3 := &model.User{
		ID:       "3",
		Username: "Ruslan",
	}
	storage.Users["1"] = user1
	storage.Users["2"] = user2
	storage.Users["3"] = user3

	return storage
}

func (s *Storage) SaveToFile(filename string) error {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
