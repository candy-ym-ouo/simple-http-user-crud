package handler

import (
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

func TestCreateUserValidation(t *testing.T) {
	h := NewUserHandler(model.NewUserStore())
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"","email":"bad-email","age":-1}`))
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("无效参数状态码 = %d，期望 %d", recorder.Code, http.StatusBadRequest)
	}
}

// TestRejectDisplayNameEmail 确保带显示名的邮箱文本不能通过校验，
// 接口只允许纯邮箱地址（如 zhangsan@example.com）被创建或更新。
func TestRejectDisplayNameEmail(t *testing.T) {
	h := NewUserHandler(model.NewUserStore())

	// 先创建一个合法用户，供后续更新路径使用。
	create := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"张三","email":"zhangsan@example.com","age":28}`))
	createRecorder := httptest.NewRecorder()
	h.ServeHTTP(createRecorder, create)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("创建用户状态码 = %d，期望 %d", createRecorder.Code, http.StatusCreated)
	}

	for _, tc := range []struct {
		name string
		body string
	}{
		{`带引号显示名`, `{"name":"张三","email":"\"张三\" <zhangsan@example.com>","age":28}`},
		{`裸显示名`, `{"name":"张三","email":"张三 <zhangsan@example.com>","age":28}`},
		{`尖括号包裹`, `{"name":"张三","email":"<zhangsan@example.com>","age":28}`},
		{`带注释`, `{"name":"张三","email":"zhangsan@example.com (注释)","age":28}`},
	} {
		t.Run("创建/"+tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tc.body))
			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, req)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("创建状态码 = %d，期望 %d，响应=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), "邮箱格式无效") {
				t.Fatalf("错误消息不匹配，响应=%s", recorder.Body.String())
			}
		})

		t.Run("更新/"+tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/users/1", strings.NewReader(tc.body))
			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, req)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("更新状态码 = %d，期望 %d，响应=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
		})
	}

	// 合法的纯邮箱地址仍应通过校验。
	validReq := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"李四","email":"lisi@example.com","age":30}`))
	validRecorder := httptest.NewRecorder()
	h.ServeHTTP(validRecorder, validReq)
	if validRecorder.Code != http.StatusCreated || !strings.Contains(validRecorder.Body.String(), "lisi@example.com") {
		t.Fatalf("合法邮箱被错误拒绝：状态码=%d，响应=%s", validRecorder.Code, validRecorder.Body.String())
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
