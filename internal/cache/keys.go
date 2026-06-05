package cache

import "fmt"

func TaskByIDKey(id int64) string {
	return fmt.Sprintf("tasks:task:%d", id)
}

func TasksListKey() string {
	return "tasks:list"
}

func TasksListKeyPaginated(page, limit int) string {
	return fmt.Sprintf("tasks:list:page=%d:limit=%d", page, limit)
}
