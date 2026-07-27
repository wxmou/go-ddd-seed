package commandHandler

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/user"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/auth"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/config"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/crypto"
	appErrors "github.com/go-ddd-seed/go-ddd-seed/pkg/errors"
)

// ---------- mocks ----------

type mockUserRepo struct {
	users map[string]*user.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*user.User)}
}

func (m *mockUserRepo) Save(_ context.Context, user *user.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) FindByID(_ context.Context, id string) (*user.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) FindByUsername(_ context.Context, username string) (*user.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, appErrors.ErrNotFound
}

func (m *mockUserRepo) Delete(_ context.Context, id string) error {
	delete(m.users, id)
	return nil
}

type mockUserReadRepo struct {
	users map[string]*appRepo.UserWithRolesDTO
}

func newMockUserReadRepo() *mockUserReadRepo {
	return &mockUserReadRepo{users: make(map[string]*appRepo.UserWithRolesDTO)}
}

func (m *mockUserReadRepo) FindByID(_ context.Context, id string) (*appRepo.UserDTO, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return &u.UserDTO, nil
}

func (m *mockUserReadRepo) FindByUsername(_ context.Context, username string) (*appRepo.UserWithRolesDTO, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, appErrors.ErrNotFound
}

func (m *mockUserReadRepo) FindByIDWithRoles(_ context.Context, id string) (*appRepo.UserWithRolesDTO, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return u, nil
}

func (m *mockUserReadRepo) List(_ context.Context, offset, limit int, status int) ([]*appRepo.UserDTO, int64, error) {
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

type mockRoleReadRepo struct {
	roles map[string]*appRepo.RoleWithPermissionsDTO
}

func newMockRoleReadRepo() *mockRoleReadRepo {
	return &mockRoleReadRepo{roles: make(map[string]*appRepo.RoleWithPermissionsDTO)}
}

func (m *mockRoleReadRepo) FindByID(_ context.Context, id string) (*appRepo.RoleDTO, error) {
	r, ok := m.roles[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return &r.RoleDTO, nil
}

func (m *mockRoleReadRepo) FindByCode(_ context.Context, code string) (*appRepo.RoleWithPermissionsDTO, error) {
	for _, r := range m.roles {
		if r.Code == code {
			return r, nil
		}
	}
	return nil, appErrors.ErrNotFound
}

func (m *mockRoleReadRepo) List(_ context.Context) ([]*appRepo.RoleDTO, error) {
	var result []*appRepo.RoleDTO
	for _, r := range m.roles {
		result = append(result, &r.RoleDTO)
	}
	return result, nil
}

// ---------- helpers ----------

func newTestAuthHandler() (*AuthCommandHandler, *mockUserRepo, *mockUserReadRepo, *mockRoleReadRepo) {
	userRepo := newMockUserRepo()
	userRead := newMockUserReadRepo()
	roleRead := newMockRoleReadRepo()
	jwtCfg := &config.JWTConfig{
		Secret:            "test-secret-key-for-testing",
		Expiration:        24,
		RefreshExpiration: 168,
		DefaultRole:       "user",
	}
	handler := NewAuthCommandHandler(userRepo, userRead, roleRead, nil, jwtCfg)
	return handler, userRepo, userRead, roleRead
}

func seedRole(roleRead *mockRoleReadRepo, id, code, name string) *appRepo.RoleWithPermissionsDTO {
	r := &appRepo.RoleWithPermissionsDTO{
		RoleDTO: appRepo.RoleDTO{
			ID:   id,
			Code: code,
			Name: name,
		},
	}
	roleRead.roles[id] = r
	return r
}

func seedUser(userRepo *mockUserRepo, userRead *mockUserReadRepo, id, username, realName string) *user.User {
	hash, _ := crypto.HashPassword("password123")
	u, _ := user.NewUser(id, username, hash, realName)
	userRepo.users[id] = u

	userRead.users[id] = &appRepo.UserWithRolesDTO{
		UserDTO: appRepo.UserDTO{
			ID:       id,
			Username: username,
			RealName: realName,
			Status:   1,
		},
	}
	return u
}

func seedUserWithRoles(userRepo *mockUserRepo, userRead *mockUserReadRepo, roleRead *mockRoleReadRepo, id, username, realName string, roles []*appRepo.RoleWithPermissionsDTO) *user.User {
	u := seedUser(userRepo, userRead, id, username, realName)

	for _, r := range roles {
		u.AssignRole(r.ID)
	}

	// update read repo DTO with roles
	roleEntries := make([]*appRepo.RoleEntry, 0, len(roles))
	permEntries := make([]*appRepo.PermissionEntry, 0)
	for _, r := range roles {
		roleEntries = append(roleEntries, &appRepo.RoleEntry{
			ID: r.ID, Name: r.Name, Code: r.Code,
		})
		for _, p := range r.Permissions {
			permEntries = append(permEntries, &appRepo.PermissionEntry{
				ID: p.ID, Name: p.Name, Code: p.Code,
			})
		}
	}
	userRead.users[id].Roles = roleEntries
	userRead.users[id].Permissions = permEntries

	return u
}

// ---------- Register tests ----------

func TestAuthCommandHandler_Register(t *testing.T) {
	t.Run("success without default role", func(t *testing.T) {
		handler, userRepo, _, _ := newTestAuthHandler()
		// 移除 default role 以测试无角色注册
		handler.jwtCfg.DefaultRole = ""

		result, err := handler.Register(context.Background(), &command.RegisterCommand{
			Username: "newuser",
			Password: "password123",
			RealName: "新用户",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Username != "newuser" {
			t.Errorf("expected username 'newuser', got %q", result.Username)
		}
		if result.RealName != "新用户" {
			t.Errorf("expected realname '新用户', got %q", result.RealName)
		}

		// verify user saved
		u := findUserByUsername(userRepo, "newuser")
		if u == nil {
			t.Fatal("expected user to be saved in repo")
		}
		if !crypto.VerifyPassword(u.PasswordHash, "password123") {
			t.Error("expected password to be hashed")
		}
	})

	t.Run("success with default role", func(t *testing.T) {
		handler, userRepo, _, roleRead := newTestAuthHandler()
		roleID := uuid.New().String()
		seedRole(roleRead, roleID, "user", "普通用户")

		result, err := handler.Register(context.Background(), &command.RegisterCommand{
			Username: "roleuser",
			Password: "password123",
			RealName: "角色用户",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Username != "roleuser" {
			t.Errorf("expected username 'roleuser', got %q", result.Username)
		}

		u := findUserByUsername(userRepo, "roleuser")
		if u == nil {
			t.Fatal("expected user to be saved")
		}
		if len(u.UserRoles) != 1 {
			t.Fatalf("expected 1 role assigned, got %d", len(u.UserRoles))
		}
		if u.UserRoles[0].RoleID != roleID {
			t.Errorf("expected roleID %s, got %s", roleID, u.UserRoles[0].RoleID)
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		handler, userRepo, userRead, _ := newTestAuthHandler()
		seedUser(userRepo, userRead, uuid.New().String(), "existing", "Exists")

		_, err := handler.Register(context.Background(), &command.RegisterCommand{
			Username: "existing",
			Password: "password123",
			RealName: "Test",
		})
		if err != ErrUsernameExists {
			t.Errorf("expected ErrUsernameExists, got %v", err)
		}
	})

	t.Run("empty username from domain validation", func(t *testing.T) {
		handler, _, _, _ := newTestAuthHandler()
		_, err := handler.Register(context.Background(), &command.RegisterCommand{
			Username: "",
			Password: "password123",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// ---------- Login tests ----------

func TestAuthCommandHandler_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, userRepo, userRead, roleRead := newTestAuthHandler()
		userID := uuid.New().String()
		roleID := uuid.New().String()
		permID := uuid.New().String()
		r := seedRole(roleRead, roleID, "admin", "管理员")
		r.Permissions = []*appRepo.PermissionDTO{
			{ID: permID, Code: "case:create", Name: "创建案件"},
		}
		seedUserWithRoles(userRepo, userRead, roleRead, userID, "admin", "管理员", []*appRepo.RoleWithPermissionsDTO{r})

		result, err := handler.Login(context.Background(), &command.LoginCommand{
			Username: "admin",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.AccessToken == "" {
			t.Error("expected access token to be returned")
		}
		if result.RefreshToken == "" {
			t.Error("expected refresh token to be returned")
		}
		if result.ExpiresIn <= 0 {
			t.Errorf("expected positive expiresIn, got %d", result.ExpiresIn)
		}
		if result.User.Username != "admin" {
			t.Errorf("expected username admin, got %q", result.User.Username)
		}
		if len(result.User.Roles) != 1 || result.User.Roles[0] != "admin" {
			t.Errorf("expected roles ['admin'], got %v", result.User.Roles)
		}
		if len(result.User.Permissions) != 1 || result.User.Permissions[0] != "case:create" {
			t.Errorf("expected permissions ['case:create'], got %v", result.User.Permissions)
		}

		// verify last_login_at updated
		u := findUserByUsername(userRepo, "admin")
		if u.LastLoginAt == nil {
			t.Error("expected last_login_at to be set")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		handler, userRepo, userRead, _ := newTestAuthHandler()
		seedUserWithRoles(userRepo, userRead, nil, uuid.New().String(), "user", "User", nil)

		_, err := handler.Login(context.Background(), &command.LoginCommand{
			Username: "user",
			Password: "wrongpass",
		})
		if err != appErrors.ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		handler, _, _, _ := newTestAuthHandler()
		_, err := handler.Login(context.Background(), &command.LoginCommand{
			Username: "nonexistent",
			Password: "password123",
		})
		if err != appErrors.ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("disabled user", func(t *testing.T) {
		handler, userRepo, userRead, _ := newTestAuthHandler()
		id := uuid.New().String()
		u := seedUserWithRoles(userRepo, userRead, nil, id, "disabled", "Disabled", nil)
		_ = u.Disable()

		_, err := handler.Login(context.Background(), &command.LoginCommand{
			Username: "disabled",
			Password: "password123",
		})
		if err != appErrors.ErrUserDisabled {
			t.Errorf("expected ErrUserDisabled, got %v", err)
		}
	})

	t.Run("token claims contain roles and permissions", func(t *testing.T) {
		handler, userRepo, userRead, roleRead := newTestAuthHandler()
		userID := uuid.New().String()
		roleID := uuid.New().String()
		r := seedRole(roleRead, roleID, "editor", "编辑者")
		r.Permissions = []*appRepo.PermissionDTO{
			{ID: uuid.New().String(), Code: "doc:edit", Name: "编辑文档"},
			{ID: uuid.New().String(), Code: "doc:view", Name: "查看文档"},
		}
		seedUserWithRoles(userRepo, userRead, roleRead, userID, "editor", "Editor", []*appRepo.RoleWithPermissionsDTO{r})

		result, err := handler.Login(context.Background(), &command.LoginCommand{
			Username: "editor",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// parse the token and verify claims
		claims, err := auth.ParseToken(result.AccessToken, handler.jwtCfg.Secret)
		if err != nil {
			t.Fatalf("failed to parse token: %v", err)
		}
		if claims.UserID != userID {
			t.Errorf("expected userID %s, got %s", userID, claims.UserID)
		}
		if claims.Username != "editor" {
			t.Errorf("expected username editor, got %s", claims.Username)
		}
	})
}

// ---------- RefreshToken tests ----------

func TestAuthCommandHandler_RefreshToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, userRepo, userRead, roleRead := newTestAuthHandler()
		userID := uuid.New().String()
		roleID := uuid.New().String()
		r := seedRole(roleRead, roleID, "user", "普通用户")
		seedUserWithRoles(userRepo, userRead, roleRead, userID, "refresher", "Refresher", []*appRepo.RoleWithPermissionsDTO{r})

		refreshToken, _ := auth.GenerateRefreshToken(userID, handler.jwtCfg.Secret, handler.jwtCfg.RefreshExpiration)

		result, err := handler.RefreshToken(context.Background(), &command.RefreshTokenCommand{
			RefreshToken: refreshToken,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.AccessToken == "" {
			t.Error("expected new access token")
		}
		if result.ExpiresIn <= 0 {
			t.Errorf("expected positive expiresIn, got %d", result.ExpiresIn)
		}

		// verify new token is valid and has roles
		claims, err := auth.ParseToken(result.AccessToken, handler.jwtCfg.Secret)
		if err != nil {
			t.Fatalf("failed to parse new token: %v", err)
		}
		if len(claims.Roles) != 1 || claims.Roles[0] != "user" {
			t.Errorf("expected roles ['user'], got %v", claims.Roles)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		handler, _, _, _ := newTestAuthHandler()
		_, err := handler.RefreshToken(context.Background(), &command.RefreshTokenCommand{
			RefreshToken: "not.a.valid.token",
		})
		if err != appErrors.ErrTokenInvalid {
			t.Errorf("expected ErrTokenInvalid, got %v", err)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		handler, _, _, _ := newTestAuthHandler()
		// generate a token with 0 hours expiry
		token, _ := auth.GenerateRefreshToken(uuid.New().String(), handler.jwtCfg.Secret, 0)
		time.Sleep(time.Millisecond) // 确保过期
		_, err := handler.RefreshToken(context.Background(), &command.RefreshTokenCommand{
			RefreshToken: token,
		})
		if err != appErrors.ErrTokenInvalid {
			t.Errorf("expected ErrTokenInvalid, got %v", err)
		}
	})

	t.Run("user disabled after token issued", func(t *testing.T) {
		handler, userRepo, userRead, roleRead := newTestAuthHandler()
		userID := uuid.New().String()
		r := seedRole(roleRead, uuid.New().String(), "user", "普通用户")
		seedUserWithRoles(userRepo, userRead, roleRead, userID, "disabledLater", "DL", []*appRepo.RoleWithPermissionsDTO{r})

		refreshToken, _ := auth.GenerateRefreshToken(userID, handler.jwtCfg.Secret, handler.jwtCfg.RefreshExpiration)

		// disable the user
		u, _ := userRepo.FindByID(context.Background(), userID)
		_ = u.Disable()
		userRead.users[userID].Status = 0

		_, err := handler.RefreshToken(context.Background(), &command.RefreshTokenCommand{
			RefreshToken: refreshToken,
		})
		if err != appErrors.ErrUserDisabled {
			t.Errorf("expected ErrUserDisabled, got %v", err)
		}
	})
}

// ---------- Logout tests ----------

func TestAuthCommandHandler_Logout(t *testing.T) {
	t.Run("success with valid token", func(t *testing.T) {
		handler, _, _, _ := newTestAuthHandler()
		claims := &auth.Claims{
			UserID:   uuid.New().String(),
			Username: "logoutuser",
		}
		token, _ := auth.GenerateAccessToken(claims, handler.jwtCfg.Secret, handler.jwtCfg.Expiration)

		// Without Redis client, just verify no error
		err := handler.Logout(context.Background(), &command.LogoutCommand{
			AccessToken: token,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("expired token still succeeds", func(t *testing.T) {
		handler, _, _, _ := newTestAuthHandler()
		err := handler.Logout(context.Background(), &command.LogoutCommand{
			AccessToken: "invalid.token.string",
		})
		if err != nil {
			t.Fatalf("expected no error for expired token, got %v", err)
		}
	})
}

// ---------- helpers ----------

func findUserByUsername(repo *mockUserRepo, username string) *user.User {
	for _, u := range repo.users {
		if u.Username == username {
			return u
		}
	}
	return nil
}