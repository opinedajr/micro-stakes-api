package auth

import (
	"context"
	"testing"

	"github.com/opinedajr/micro-stakes-api/internal/infrastructure/database"
	"github.com/stretchr/testify/assert"
)

func setupPostgresUserRepository(t *testing.T) (*User, UserRepository, func()) {
	t.Helper()

	ctx := context.Background()
	sqliteDB := database.NewSQLiteDatabase(t)
	db, err := sqliteDB.Connect(ctx)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = sqliteDB.Migrate(&User{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	cleanup := func() {
		sqliteDB.Close()
	}

	user := &User{
		FullName:        "John Doe",
		Email:           "john.doe@example.com",
		IdentityID:      "keycloak-user-123",
		IdentityAdapter: IdentityAdapterKeycloak,
	}

	return user, NewPostgresUserRepository(db), cleanup
}

func TestPostgresUserRepository_CreateUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		user          *User
		prepDB        func(*testing.T, UserRepository)
		expectError   bool
		errorContains string
	}{
		{
			name: "success - valid user",
			user: &User{
				FullName:        "Jane Smith",
				Email:           "jane.smith@example.com",
				IdentityID:      "keycloak-user-456",
				IdentityAdapter: IdentityAdapterKeycloak,
			},
			prepDB:      func(t *testing.T, r UserRepository) {},
			expectError: false,
		},
		{
			name: "error - duplicate email",
			user: &User{
				FullName:        "John Duplicate",
				Email:           "duplicate@example.com",
				IdentityID:      "keycloak-user-789",
				IdentityAdapter: IdentityAdapterKeycloak,
			},
			prepDB: func(t *testing.T, r UserRepository) {
				existingUser := &User{
					FullName:        "Existing User",
					Email:           "duplicate@example.com",
					IdentityID:      "keycloak-user-001",
					IdentityAdapter: IdentityAdapterKeycloak,
				}
				err := r.CreateUser(ctx, existingUser)
				assert.NoError(t, err)
			},
			expectError:   true,
			errorContains: "database error",
		},
		{
			name: "error - database connection closed",
			user: &User{
				FullName:        "Connection Error User",
				Email:           "conn.error@example.com",
				IdentityID:      "keycloak-user-002",
				IdentityAdapter: IdentityAdapterKeycloak,
			},
			prepDB:        func(t *testing.T, r UserRepository) {},
			expectError:   true,
			errorContains: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			sqliteDB := database.NewSQLiteDatabase(t)
			db, err := sqliteDB.Connect(ctx)
			assert.NoError(t, err)
			err = sqliteDB.Migrate(&User{})
			assert.NoError(t, err)

			if tt.name == "error - database connection closed" {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			}

			repo := NewPostgresUserRepository(db)
			tt.prepDB(t, repo)

			err = repo.CreateUser(ctx, tt.user)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.user.ID)
				assert.NotZero(t, tt.user.CreatedAt)
				assert.NotZero(t, tt.user.UpdatedAt)
			}
		})
	}
}

func TestPostgresUserRepository_FindByEmail(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		email       string
		prepDB      func(*testing.T, UserRepository)
		expectError bool
		errorIs     error
		validate    func(*testing.T, *User)
	}{
		{
			name:  "success - user found",
			email: "found@example.com",
			prepDB: func(t *testing.T, r UserRepository) {
				user := &User{
					FullName:        "Found User",
					Email:           "found@example.com",
					IdentityID:      "keycloak-user-111",
					IdentityAdapter: IdentityAdapterKeycloak,
				}
				err := r.CreateUser(ctx, user)
				assert.NoError(t, err)
			},
			expectError: false,
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, "found@example.com", u.Email)
				assert.Equal(t, "Found User", u.FullName)
				assert.NotZero(t, u.ID)
			},
		},
		{
			name:        "success - user not found",
			email:       "notfound@example.com",
			prepDB:      func(t *testing.T, r UserRepository) {},
			expectError: true,
			errorIs:     ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			sqliteDB := database.NewSQLiteDatabase(t)
			db, err := sqliteDB.Connect(ctx)
			assert.NoError(t, err)
			err = sqliteDB.Migrate(&User{})
			assert.NoError(t, err)

			repo := NewPostgresUserRepository(db)
			tt.prepDB(t, repo)

			user, err := repo.FindByEmail(ctx, tt.email)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if tt.validate != nil {
					tt.validate(t, user)
				}
			}
		})
	}
}

func TestPostgresUserRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	var createdUserID uint

	sqliteDB := database.NewSQLiteDatabase(t)
	db, err := sqliteDB.Connect(ctx)
	assert.NoError(t, err)
	err = sqliteDB.Migrate(&User{})
	assert.NoError(t, err)

	repo := NewPostgresUserRepository(db)

	user := &User{
		FullName:        "ID Search User",
		Email:           "idsearch@example.com",
		IdentityID:      "keycloak-user-222",
		IdentityAdapter: IdentityAdapterKeycloak,
	}
	err = repo.CreateUser(ctx, user)
	assert.NoError(t, err)
	createdUserID = user.ID

	tests := []struct {
		name        string
		id          uint
		expectError bool
		errorIs     error
		validate    func(*testing.T, *User)
	}{
		{
			name:        "success - user found by ID",
			id:          createdUserID,
			expectError: false,
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, createdUserID, u.ID)
				assert.Equal(t, "idsearch@example.com", u.Email)
			},
		},
		{
			name:        "success - user not found by ID",
			id:          99999,
			expectError: true,
			errorIs:     ErrUserNotFound,
		},
		{
			name:        "success - user not found with ID 0",
			id:          0,
			expectError: true,
			errorIs:     ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.FindByID(ctx, tt.id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if tt.validate != nil {
					tt.validate(t, user)
				}
			}
		})
	}
}

func TestPostgresUserRepository_FindByIdentityID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		identityID   string
		adapter      IdentityAdapter
		prepDB       func(*testing.T, UserRepository)
		expectError  bool
		errorIs      error
		validateUser func(*testing.T, *User)
	}{
		{
			name:       "success - user found by identity ID",
			identityID: "keycloak-user-333",
			adapter:    IdentityAdapterKeycloak,
			prepDB: func(t *testing.T, r UserRepository) {
				user := &User{
					FullName:        "Identity Search User",
					Email:           "identity@example.com",
					IdentityID:      "keycloak-user-333",
					IdentityAdapter: IdentityAdapterKeycloak,
				}
				err := r.CreateUser(ctx, user)
				assert.NoError(t, err)
			},
			expectError: false,
			validateUser: func(t *testing.T, u *User) {
				assert.Equal(t, "keycloak-user-333", u.IdentityID)
				assert.Equal(t, IdentityAdapterKeycloak, u.IdentityAdapter)
			},
		},
		{
			name:        "success - user not found by identity ID",
			identityID:  "nonexistent-keycloak-user",
			adapter:     IdentityAdapterKeycloak,
			prepDB:      func(t *testing.T, r UserRepository) {},
			expectError: true,
			errorIs:     ErrUserNotFound,
		},
		{
			name:       "success - user not found with wrong adapter",
			identityID: "keycloak-user-333",
			adapter:    "unknown_adapter",
			prepDB: func(t *testing.T, r UserRepository) {
				user := &User{
					FullName:        "Identity Search User",
					Email:           "identity@example.com",
					IdentityID:      "keycloak-user-333",
					IdentityAdapter: IdentityAdapterKeycloak,
				}
				err := r.CreateUser(ctx, user)
				assert.NoError(t, err)
			},
			expectError: true,
			errorIs:     ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			sqliteDB := database.NewSQLiteDatabase(t)
			db, err := sqliteDB.Connect(ctx)
			assert.NoError(t, err)
			err = sqliteDB.Migrate(&User{})
			assert.NoError(t, err)

			repo := NewPostgresUserRepository(db)
			tt.prepDB(t, repo)

			user, err := repo.FindByIdentityID(ctx, tt.identityID, tt.adapter)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if tt.validateUser != nil {
					tt.validateUser(t, user)
				}
			}
		})
	}
}
