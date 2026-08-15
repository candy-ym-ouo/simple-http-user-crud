// Package model 定义领域数据以及内存存储实现。
package model

import (
	"errors"
	"sort"
	"sync"
)

var ErrUserNotFound = errors.New("用户不存在")

// User 是接口返回的用户实体。ID 由服务端生成，客户端不应传入。
type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// UserStore 用 map 保存用户；mu 使并发请求读写同一份内存数据时保持安全。
type UserStore struct {
	mu     sync.RWMutex
	nextID int64
	users  map[int64]User
}

// NewUserStore 创建一个空的用户仓储。
func NewUserStore() *UserStore {
	return &UserStore{nextID: 1, users: make(map[int64]User)}
}

// Create 保存用户并分配递增 ID，返回保存后的完整对象。
func (s *UserStore) Create(user User) User {
	s.mu.Lock()
	defer s.mu.Unlock()

	user.ID = s.nextID
	s.nextID++
	s.users[user.ID] = user
	return user
}

// Get 按 ID 查询用户。返回副本，调用方无法修改仓储中的原始数据。
func (s *UserStore) Get(id int64) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

// List 返回所有用户，并按 ID 升序排列，保证接口结果稳定。
func (s *UserStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool { return users[i].ID < users[j].ID })
	return users
}

// Update 用新字段替换已有用户，但始终保留路径指定的 ID。
func (s *UserStore) Update(id int64, user User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[id]; !ok {
		return User{}, ErrUserNotFound
	}
	// 入参 user 来自请求体，不含 ID；这里用路径指定的 ID 覆盖，
	// 保证存储与返回的记录都保留原 ID，避免数据一致性问题。
	user.ID = id
	s.users[id] = user
	return user, nil
}

// Delete 删除指定用户；不存在时返回 ErrUserNotFound。
func (s *UserStore) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[id]; !ok {
		return ErrUserNotFound
	}
	delete(s.users, id)
	return nil
}
