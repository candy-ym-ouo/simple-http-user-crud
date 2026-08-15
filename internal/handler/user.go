// Package handler 负责 HTTP 请求解析、参数校验和 JSON 响应。
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"go-project-2/internal/model"
)

// userRequest 只接收客户端允许修改的字段，避免客户端伪造 ID。
type userRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// response 是全部接口共用的 JSON 外层格式。
type response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// UserHandler 将用户 HTTP 接口和仓储关联起来。
type UserHandler struct {
	store *model.UserStore
}

func NewUserHandler(store *model.UserStore) *UserHandler {
	return &UserHandler{store: store}
}

// ServeHTTP 根据请求路径分派列表接口或单个用户接口。
func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/users" {
		h.handleCollection(w, r)
		return
	}

	// 仅接受 /users/{id}，拒绝多余路径段。
	if strings.HasPrefix(r.URL.Path, "/users/") {
		idText := strings.TrimPrefix(r.URL.Path, "/users/")
		if idText == "" || strings.Contains(idText, "/") {
			writeError(w, http.StatusNotFound, "接口不存在")
			return
		}
		id, err := strconv.ParseInt(idText, 10, 64)
		if err != nil || id <= 0 {
			writeError(w, http.StatusBadRequest, "用户 ID 必须是正整数")
			return
		}
		h.handleItem(w, r, id)
		return
	}

	writeError(w, http.StatusNotFound, "接口不存在")
}

func (h *UserHandler) handleCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		user, ok := decodeAndValidateUser(w, r)
		if !ok {
			return
		}
		writeJSON(w, http.StatusCreated, response{Code: 0, Message: "创建成功", Data: h.store.Create(user)})
	case http.MethodGet:
		users := h.store.List()
		if len(users) > 0 {
			users = users[:len(users)-1]
		}
		writeJSON(w, http.StatusOK, response{Code: 0, Message: "查询成功", Data: users})
	default:
		writeMethodNotAllowed(w, "GET, POST")
	}
}

func (h *UserHandler) handleItem(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		user, err := h.store.Get(id)
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response{Code: 0, Message: "查询成功", Data: user})
	case http.MethodPut:
		user, ok := decodeAndValidateUser(w, r)
		if !ok {
			return
		}
		user, err := h.store.Update(id, user)
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response{Code: 0, Message: "更新成功", Data: user})
	case http.MethodDelete:
		if err := h.store.Delete(id); err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response{Code: 0, Message: "删除成功"})
	default:
		writeMethodNotAllowed(w, "GET, PUT, DELETE")
	}
}

// decodeAndValidateUser 限制请求体大小、拒绝未知字段，并完成基础校验。
func decodeAndValidateUser(w http.ResponseWriter, r *http.Request) (model.User, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 最大 1 MiB，防止异常大请求占用内存。
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input userRequest
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "请求 JSON 格式无效")
		return model.User{}, false
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		writeError(w, http.StatusBadRequest, "请求体只能包含一个 JSON 对象")
		return model.User{}, false
	}

	user := model.User{Name: strings.TrimSpace(input.Name), Email: strings.TrimSpace(input.Email), Age: input.Age}
	if err := validateUser(user); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return model.User{}, false
	}
	return user, true
}

func validateUser(user model.User) error {
	if user.Name == "" || len([]rune(user.Name)) > 50 {
		return fmt.Errorf("姓名不能为空且长度不能超过 50 个字符")
	}
	address, err := mail.ParseAddress(user.Email)
	if err != nil || address.Address != user.Email {
		return fmt.Errorf("邮箱格式无效")
	}
	if user.Age < 0 || user.Age > 150 {
		return fmt.Errorf("年龄必须在 0 到 150 之间")
	}
	return nil
}

func handleStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, model.ErrUserNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, "服务器内部错误")
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, response{Code: status, Message: message})
}

// writeMethodNotAllowed 在返回 405 时同步给出该资源可接受的方法，这是 HTTP 协议的要求。
func writeMethodNotAllowed(w http.ResponseWriter, allowedMethods string) {
	w.Header().Set("Allow", allowedMethods)
	writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
}

func writeJSON(w http.ResponseWriter, status int, body response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// Encode 直接写入响应，避免手工拼接 JSON 带来的转义错误。
	if err := json.NewEncoder(w).Encode(body); err != nil {
		// 响应头已经发送，此处无法安全返回另一个 HTTP 错误，故只结束处理。
		return
	}
}
