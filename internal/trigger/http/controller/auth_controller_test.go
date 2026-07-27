package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/go-ddd-seed/go-ddd-seed/internal/application/commandHandler"
	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/user"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/auth"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/config"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/crypto"
	appErrors "github.com/go-ddd-seed/go-ddd-seed/pkg/errors"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/middleware"
)

// ---------- mocks ----------

type ctrlMockUserRepo struct {
	users map[string]*user.User
}

func newCtrlMockUserRepo() *ctrlMockUserRepo {
	return &ctrlMockUserRepo{users: make(map[string]*user.User)}
}

func (m *ctrlMockUserRepo) Save(_ context.Context, user *user.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *ctrlMockUserRepo) FindByID(_ context.Context, id string) (*user.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return u, nil
}

func (m *ctrlMockUserRepo) FindByUsername(_ context.Context, username string) (*user.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, appErrors.ErrNotFound
}

func (m *ctrlMockUserRepo) Delete(_ context.Context, id string) error {
	delete(m.users, id)
	return nil
}

type ctrlMockUserReadRepo struct {
	users map[string]*appRepo.UserWithRolesDTO
}

func newCtrlMockUserReadRepo() *ctrlMockUserReadRepo {
	return &ctrlMockUserReadRepo{users: make(map[string]*appRepo.UserWithRolesDTO)}
}

func (m *ctrlMockUserReadRepo) FindByID(_ context.Context, id string) (*appRepo.UserDTO, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return &u.UserDTO, nil
}

func (m *ctrlMockUserReadRepo) FindByUsername(_ context.Context, username string) (*appRepo.UserWithRolesDTO, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, appErrors.ErrNotFound
}

func (m *ctrlMockUserReadRepo) FindByIDWithRoles(_ context.Context, id string) (*appRepo.UserWithRolesDTO, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return u, nil
}

func (m *ctrlMockUserReadRepo) List(_ context.Context, offset, limit int, status int) ([]*appRepo.UserDTO, int64, error) {
	var result []*appRepo.UserDTO
	for _, u := range m.users {
		if status < 0 || u.Status == status {
			result = append(result, &u.UserDTO)
		}
	}
	total := int64(len(result))
	if offset >= len(result) {
		return []*appRepo.UserDTO{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

type ctrlMockRoleReadRepo struct {
	roles map[string]*appRepo.RoleWithPermissionsDTO
}

func newCtrlMockRoleReadRepo() *ctrlMockRoleReadRepo {
	return &ctrlMockRoleReadRepo{roles: make(map[string]*appRepo.RoleWithPermissionsDTO)}
}

func (m *ctrlMockRoleReadRepo) FindByID(_ context.Context, id string) (*appRepo.RoleDTO, error) {
	r, ok := m.roles[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return &r.RoleDTO, nil
}

func (m *ctrlMockRoleReadRepo) FindByCode(_ context.Context, code string) (*appRepo.RoleWithPermissionsDTO, error) {
	for _, r := range m.roles {
		if r.Code == code {
			return r, nil
		}
	}
	return nil, appErrors.ErrNotFound
}

func (m *ctrlMockRoleReadRepo) List(_ context.Context) ([]*appRepo.RoleDTO, error) {
	var result []*appRepo.RoleDTO
	for _, r := range m.roles {
		result = append(result, &r.RoleDTO)
	}
	return result, nil
}

// ---------- test environment ----------

type authCtrlTestEnv struct {
	engine     *gin.Engine
	handler    *commandHandler.AuthCommandHandler
	userRepo   *ctrlMockUserRepo
	userRead   *ctrlMockUserReadRepo
	roleRead   *ctrlMockRoleReadRepo
	jwtCfg     *config.JWTConfig
}

func newAuthCtrlTestEnv() *authCtrlTestEnv {
	gin.SetMode(gin.TestMode)

	userRepo := newCtrlMockUserRepo()
	userRead := newCtrlMockUserReadRepo()
	roleRead := newCtrlMockRoleReadRepo()
	jwtCfg := &config.JWTConfig{
		Secret:            "test-secret-key",
		Expiration:        24,
		RefreshExpiration: 168,
		DefaultRole:       "user",
	}

	handler := commandHandler.NewAuthCommandHandler(userRepo, userRead, roleRead, nil, jwtCfg)
	ctrl := NewAuthController(handler)

	engine := gin.New()
	api := engine.Group("/api/v1")
	authGroup := api.Group("")
	authGroup.Use(middleware.Auth(jwtCfg.Secret, nil))

	ctrl.RegisterRoutes(api, authGroup)

	return &authCtrlTestEnv{
		engine:   engine,
		handler:  handler,
		userRepo: userRepo,
		userRead: userRead,
		roleRead: roleRead,
		jwtCfg:   jwtCfg,
	}
}

func (env *authCtrlTestEnv) doRequest(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.engine.ServeHTTP(w, req)
	return w
}

func (env *authCtrlTestEnv) doAuthRequest(method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	env.engine.ServeHTTP(w, req)
	return w
}

func (env *authCtrlTestEnv) seedUser(username, password, realName string) (string, string) {
	id := uuid.New().String()
	hash, _ := crypto.HashPassword(password)
	u, _ := user.NewUser(id, username, hash, realName)
	env.userRepo.users[id] = u
	env.userRead.users[id] = &appRepo.UserWithRolesDTO{
		UserDTO: appRepo.UserDTO{
			ID:       id,
			Username: username,
			RealName: realName,
			Status:   1,
		},
	}
	return id, username
}

func (env *authCtrlTestEnv) seedUserWithRoles(password string) (string, string) {
	// 创建角色
	roleID := uuid.New().String()
	r := &appRepo.RoleWithPermissionsDTO{
		RoleDTO: appRepo.RoleDTO{
			ID: roleID, Code: "admin", Name: "管理员",
		},
		Permissions: []*appRepo.PermissionDTO{
			{ID: uuid.New().String(), Code: "case:create", Name: "创建案件"},
			{ID: uuid.New().String(), Code: "case:view", Name: "查看案件"},
		},
	}
	env.roleRead.roles[roleID] = r

	// 创建用户并分配角色
	id := uuid.New().String()
	hash, _ := crypto.HashPassword(password)
	u, _ := user.NewUser(id, "adminuser", hash, "管理员用户")
	u.AssignRole(roleID)
	env.userRepo.users[id] = u
	env.userRead.users[id] = &appRepo.UserWithRolesDTO{
		UserDTO: appRepo.UserDTO{
			ID:       id,
			Username: "adminuser",
			RealName: "管理员用户",
			Status:   1,
		},
		Roles: []*appRepo.RoleEntry{
			{ID: roleID, Code: "admin", Name: "管理员"},
		},
		Permissions: []*appRepo.PermissionEntry{
			{ID: uuid.New().String(), Code: "case:create", Name: "创建案件"},
			{ID: uuid.New().String(), Code: "case:view", Name: "查看案件"},
		},
	}
	return id, "adminuser"
}

// ---------- tests ----------

func TestAuthController_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newAuthCtrlTestEnv()
		w := env.doRequest("POST", "/api/v1/auth/register",
			`{"username":"newguy","password":"password123","real_name":"新用户"}`)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		r := parseAuthResp(t, w)
		if r.Code != 0 {
			t.Errorf("expected code 0, got %d", r.Code)
		}
	})

	t.Run("bad request missing fields", func(t *testing.T) {
		env := newAuthCtrlTestEnv()
		w := env.doRequest("POST", "/api/v1/auth/register", `{}`)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		env := newAuthCtrlTestEnv()
		env.seedUser("dupuser", "password123", "")

		w := env.doRequest("POST", "/api/v1/auth/register",
			`{"username":"dupuser","password":"password123","real_name":"Dup"}`)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestAuthController_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newAuthCtrlTestEnv()
		env.seedUser("testuser", "password123", "测试")

		w := env.doRequest("POST", "/api/v1/auth/login",
			`{"username":"testuser","password":"password123"}`)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		r := parseAuthResp(t, w)
		if r.Code != 0 {
			t.Errorf("expected code 0, got %d", r.Code)
		}
		data := parseLoginData(t, w)
		if data.AccessToken == "" {
			t.Error("expected access_token")
		}
		if data.RefreshToken == "" {
			t.Error("expected refresh_token")
		}
		if data.User == nil {
			t.Error("expected user info")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		env := newAuthCtrlTestEnv()
		env.seedUser("testuser", "password123", "")

		w := env.doRequest("POST", "/api/v1/auth/login",
			`{"username":"testuser","password":"wrongpass"}`)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		env := newAuthCtrlTestEnv()
		w := env.doRequest("POST", "/api/v1/auth/login",
			`{"username":"nobody","password":"password123"}`)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
}

func TestAuthController_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newAuthCtrlTestEnv()
		id, _ := env.seedUser("refresher", "password123", "Refresh")

		// generate refresh token directly
		refreshToken, _ := auth.GenerateRefreshToken(id, env.jwtCfg.Secret, env.jwtCfg.RefreshExpiration)

		w := env.doRequest("POST", "/api/v1/auth/refresh",
			`{"refresh_token":"`+refreshToken+`"}`)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		r := parseAuthResp(t, w)
		if r.Code != 0 {
			t.Errorf("expected code 0, got %d", r.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		env := newAuthCtrlTestEnv()
		w := env.doRequest("POST", "/api/v1/auth/refresh",
			`{"refresh_token":"invalid.token.here"}`)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
}

func TestAuthController_Logout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newAuthCtrlTestEnv()
		claims := &auth.Claims{
			UserID:   uuid.New().String(),
			Username: "logouttest",
		}
		token, _ := auth.GenerateAccessToken(claims, env.jwtCfg.Secret, env.jwtCfg.Expiration)

		w := env.doAuthRequest("POST", "/api/v1/auth/logout", "", token)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("without token returns 401", func(t *testing.T) {
		env := newAuthCtrlTestEnv()
		w := env.doRequest("POST", "/api/v1/auth/logout", "")

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
}

func TestAuthController_Me(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newAuthCtrlTestEnv()
		uid, username := env.seedUserWithRoles("password123")

		claims := &auth.Claims{
			UserID:      uid,
			Username:    username,
			Roles:       []string{"admin"},
			Permissions: []string{"case:create", "case:view"},
		}
		token, _ := auth.GenerateAccessToken(claims, env.jwtCfg.Secret, env.jwtCfg.Expiration)

		w := env.doAuthRequest("GET", "/api/v1/auth/me", "", token)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		r := parseAuthResp(t, w)
		if r.Code != 0 {
			t.Errorf("expected code 0, got %d", r.Code)
		}
	})

	t.Run("unauthorized without token", func(t *testing.T) {
		env := newAuthCtrlTestEnv()
		w := env.doRequest("GET", "/api/v1/auth/me", "")

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
}

// ---------- response helpers ----------

type authApiResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type loginData struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresIn    int64           `json:"expires_in"`
	User         *loginUserData `json:"user"`
}

type loginUserData struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	RealName    string   `json:"real_name"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

func parseAuthResp(t *testing.T, w *httptest.ResponseRecorder) *authApiResp {
	t.Helper()
	var r authApiResp
	if err := json.Unmarshal(w.Body.Bytes(), &r); err != nil {
		t.Fatalf("failed to parse response: %v, body: %s", err, w.Body.String())
	}
	return &r
}

func parseLoginData(t *testing.T, w *httptest.ResponseRecorder) *loginData {
	t.Helper()
	var r authApiResp
	if err := json.Unmarshal(w.Body.Bytes(), &r); err != nil {
		t.Fatalf("failed to parse response: %v, body: %s", err, w.Body.String())
	}
	if r.Data == nil {
		t.Fatal("expected data in response")
	}
	var d loginData
	if err := json.Unmarshal(r.Data, &d); err != nil {
		t.Fatalf("failed to parse login data: %v", err)
	}
	return &d
}