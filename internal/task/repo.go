package task

import (
	"errors"
	"sort"
	"time"
)

var ErrTaskNotFound = errors.New("task not found")

type Repo struct {
	data map[int64]Task
}

func NewRepo() *Repo {
	return &Repo{
		data: map[int64]Task{
			1: {
				ID:          1,
				Title:       "Изучить Redis",
				Description: "Разобрать cache-aside",
				DueDate:     time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
			},
			2: {
				ID:          2,
				Title:       "Сделать ПЗ",
				Description: "Реализовать кэширование по id",
				DueDate:     time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC),
			},
		},
	}
}

func (r *Repo) ListAll() []Task {
	tasks := make([]Task, 0, len(r.data))
	for _, t := range r.data {
		tasks = append(tasks, t)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return tasks
}

func (r *Repo) ListPaginated(page, limit int) []Task {
	all := r.ListAll()
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	start := (page - 1) * limit
	if start >= len(all) {
		return []Task{}
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	return all[start:end]
}

func (r *Repo) GetByID(id int64) (Task, error) {
	t, ok := r.data[id]
	if !ok {
		return Task{}, ErrTaskNotFound
	}
	return t, nil
}

func (r *Repo) Update(task Task) error {
	if _, ok := r.data[task.ID]; !ok {
		return ErrTaskNotFound
	}
	r.data[task.ID] = task
	return nil
}

func (r *Repo) Delete(id int64) error {
	if _, ok := r.data[id]; !ok {
		return ErrTaskNotFound
	}
	delete(r.data, id)
	return nil
}
