package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-project-2/internal/model"
)

func TestUserCRUD(t *testing.T) {
	h := NewUserHandler(model.NewUserStore())

	create := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"张三","email":"zhangsan@example.com","age":28}`))
	createRecorder := httptest.NewRecorder()
	h.ServeHTTP(createRecorder, create)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("创建用户状态码 = %d，期望 %d", createRecorder.Code, http.StatusCreated)
	}

	get := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	getRecorder := httptest.NewRecorder()
	h.ServeHTTP(getRecorder, get)
	if getRecorder.Code != http.StatusOK || !strings.Contains(getRecorder.Body.String(), "张三") {
		t.Fatalf("查询用户失败：状态码=%d，响应=%s", getRecorder.Code, getRecorder.Body.String())
	}

	update := httptest.NewRequest(http.MethodPut, "/users/1", strings.NewReader(`{"name":"李四","email":"lisi@example.com","age":30}`))
	updateRecorder := httptest.NewRecorder()
	h.ServeHTTP(updateRecorder, update)
	if updateRecorder.Code != http.StatusOK || !strings.Contains(updateRecorder.Body.String(), "李四") {
		t.Fatalf("更新用户失败：状态码=%d，响应=%s", updateRecorder.Code, updateRecorder.Body.String())
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	deleteRecorder := httptest.NewRecorder()
	h.ServeHTTP(deleteRecorder, deleteRequest)
	if deleteRecorder.Code != http.StatusOK {
		t.Fatalf("删除用户状态码 = %d，期望 %d", deleteRecorder.Code, http.StatusOK)
	}
}

func TestListUsersContainsOnlyStoredUsers(t *testing.T) {
	h := NewUserHandler(model.NewUserStore())

	for _, body := range []string{
		`{"name":"张三","email":"zhangsan@example.com","age":28}`,
		`{"name":"李四","email":"lisi@example.com","age":30}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		recorder := httptest.NewRecorder()
		h.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("创建用户状态码 = %d，期望 %d", recorder.Code, http.StatusCreated)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("查询用户状态码 = %d，期望 %d", recorder.Code, http.StatusOK)
	}

	var body struct {
		Data []model.User `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("解析响应失败：%v", err)
	}
	if len(body.Data) != 2 || body.Data[0].ID != 1 || body.Data[1].ID != 2 {
		t.Fatalf("用户列表 = %#v，期望仅包含 ID 为 1、2 的两条用户", body.Data)
	}
}

func TestListUsersReturnsEmptyArray(t *testing.T) {
	h := NewUserHandler(model.NewUserStore())
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("查询空列表状态码 = %d，期望 %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `"data":[]`) {
		t.Fatalf("空列表响应 = %s，期望 data 为空数组", recorder.Body.String())
	}
}

func TestCreateUserValidation(t *testing.T) {
	h := NewUserHandler(model.NewUserStore())
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"","email":"bad-email","age":-1}`))
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("无效参数状态码 = %d，期望 %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestRejectUnknownFieldAndMultipleJSONValues(t *testing.T) {
	h := NewUserHandler(model.NewUserStore())

	for _, body := range []string{
		`{"name":"张三","email":"zhangsan@example.com","age":28,"role":"admin"}`,
		`{"name":"张三","email":"zhangsan@example.com","age":28} {}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		recorder := httptest.NewRecorder()
		h.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("无效请求状态码 = %d，期望 %d，响应=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
	}
}

func TestMethodNotAllowedIncludesAllowHeader(t *testing.T) {
	h := NewUserHandler(model.NewUserStore())
	req := httptest.NewRequest(http.MethodPatch, "/users/1", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("状态码 = %d，期望 %d", recorder.Code, http.StatusMethodNotAllowed)
	}
	if allow := recorder.Header().Get("Allow"); allow != "GET, PUT, DELETE" {
		t.Fatalf("Allow = %q，期望 %q", allow, "GET, PUT, DELETE")
	}
}
