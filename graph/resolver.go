package graph

import (
	"ozon-graphql-api/graph/model"
	"ozon-graphql-api/internal/repository"
)

type Resolver struct {
	Repos       *repository.Repository
	Subscribers map[string]chan *model.Comment
}

func NewResolver(repos *repository.Repository) *Resolver {
	return &Resolver{
		Repos:       repos,
		Subscribers: make(map[string]chan *model.Comment),
	}
}
