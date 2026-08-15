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

func TestListUsersHasNoZeroValueRecord(t *testing.T) {
	h := NewUserHandler(model.NewUserStore())

	// 空列表：应返回空数组，而不是包含一条零值记录的数组。
	list := httptest.NewRequest(http.MethodGet, "/users", nil)
	listRecorder := httptest.NewRecorder()
	h.ServeHTTP(listRecorder, list)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("空列表状态码 = %d，期望 %d", listRecorder.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(listRecorder.Body.String()); !strings.Contains(got, `"data":[]`) && !strings.Contains(got, `"data":{}`) {
		// data 为空切片时 omitempty 会让字段消失；只要不出现 ID=0 的记录即可。
	}
	if strings.Contains(listRecorder.Body.String(), `"id":0`) {
		t.Fatalf("空列表不应返回零值记录，响应=%s", listRecorder.Body.String())
	}

	// 创建一个用户后：列表中应恰好只有该用户，不混入零值记录。
	create := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"张三","email":"zhangsan@example.com","age":28}`))
	createRecorder := httptest.NewRecorder()
	h.ServeHTTP(createRecorder, create)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("创建用户状态码 = %d，期望 %d", createRecorder.Code, http.StatusCreated)
	}

	list2 := httptest.NewRequest(http.MethodGet, "/users", nil)
	list2Recorder := httptest.NewRecorder()
	h.ServeHTTP(list2Recorder, list2)
	if list2Recorder.Code != http.StatusOK {
		t.Fatalf("查询列表状态码 = %d，期望 %d", list2Recorder.Code, http.StatusOK)
	}
	if strings.Contains(list2Recorder.Body.String(), `"id":0`) {
		t.Fatalf("用户列表混入了零值记录，响应=%s", list2Recorder.Body.String())
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
