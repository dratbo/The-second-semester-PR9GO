package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"example.com/pz9-redis-cache/internal/cache"
	"example.com/pz9-redis-cache/internal/config"
	"example.com/pz9-redis-cache/internal/task"
	"github.com/redis/go-redis/v9"
)

type TaskService struct {
	repo  *task.Repo
	redis *redis.Client
	cfg   config.Config
}

func NewTaskService(repo *task.Repo, redisClient *redis.Client, cfg config.Config) *TaskService {
	return &TaskService{
		repo:  repo,
		redis: redisClient,
		cfg:   cfg,
	}
}

func (s *TaskService) GetTaskByID(ctx context.Context, id int64) (task.Task, error) {
	key := cache.TaskByIDKey(id)

	if s.redis != nil {
		cached, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			var t task.Task
			if err := json.Unmarshal([]byte(cached), &t); err == nil {
				log.Println("cache hit:", key)
				return t, nil
			}
			log.Println("cache decode error:", err)
		} else if !errors.Is(err, redis.Nil) {
			log.Println("redis read error:", err)
		} else {
			log.Println("cache miss:", key)
		}
	}

	t, err := s.repo.GetByID(id)
	if err != nil {
		return task.Task{}, err
	}

	if s.redis != nil {
		bytes, err := json.Marshal(t)
		if err != nil {
			log.Println("cache encode error:", err)
			return t, nil
		}

		ttl := cache.TTLWithJitter(s.cfg.CacheTTL, s.cfg.CacheTTLJitter)
		if err := s.redis.Set(ctx, key, bytes, ttl).Err(); err != nil {
			log.Println("redis write error:", err)
		}
	}

	return t, nil
}

func (s *TaskService) GetTasksList(ctx context.Context, page, limit int) ([]task.Task, error) {
	var key string
	paginated := page > 0 && limit > 0
	if paginated {
		key = cache.TasksListKeyPaginated(page, limit)
	} else {
		key = cache.TasksListKey()
	}

	if s.redis != nil {
		cached, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			var tasks []task.Task
			if err := json.Unmarshal([]byte(cached), &tasks); err == nil {
				log.Println("cache hit:", key)
				return tasks, nil
			}
			log.Println("cache decode error:", err)
		} else if !errors.Is(err, redis.Nil) {
			log.Println("redis read error:", err)
		} else {
			log.Println("cache miss:", key)
		}
	}

	var tasks []task.Task
	if paginated {
		tasks = s.repo.ListPaginated(page, limit)
	} else {
		tasks = s.repo.ListAll()
	}

	if s.redis != nil {
		bytes, err := json.Marshal(tasks)
		if err != nil {
			log.Println("cache encode error:", err)
			return tasks, nil
		}

		listTTL := cache.TTLWithJitter(s.cfg.CacheTTL/2, s.cfg.CacheTTLJitter)
		if err := s.redis.Set(ctx, key, bytes, listTTL).Err(); err != nil {
			log.Println("redis write error:", err)
		} else {
			log.Println("cache set:", key)
		}
	}

	return tasks, nil
}

func (s *TaskService) invalidateTasksListCache(ctx context.Context) {
	if s.redis == nil {
		return
	}

	keys, err := s.redis.Keys(ctx, "tasks:list*").Result()
	if err != nil {
		log.Println("redis keys error:", err)
		return
	}
	if len(keys) == 0 {
		return
	}
	if err := s.redis.Del(ctx, keys...).Err(); err != nil {
		log.Println("redis delete error:", err)
		return
	}
	log.Println("cache invalidated: tasks:list*")
}

func (s *TaskService) UpdateTask(ctx context.Context, t task.Task) error {
	if err := s.repo.Update(t); err != nil {
		return err
	}

	if s.redis != nil {
		key := cache.TaskByIDKey(t.ID)
		if err := s.redis.Del(ctx, key).Err(); err != nil {
			log.Println("redis delete error:", err)
		}
		s.invalidateTasksListCache(ctx)
	}

	return nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	if s.redis != nil {
		key := cache.TaskByIDKey(id)
		if err := s.redis.Del(ctx, key).Err(); err != nil {
			log.Println("redis delete error:", err)
		}
		s.invalidateTasksListCache(ctx)
	}

	return nil
}
