package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type FamilyRepo struct {
	db *sqlx.DB
}

func NewFamilyRepo(db *sqlx.DB) *FamilyRepo {
	return &FamilyRepo{db: db}
}

func (r *FamilyRepo) CreateFamily(ctx context.Context, family *models.Family) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO families (name, slug, logo_url, plan)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		family.Name, family.Slug, family.LogoURL, family.Plan,
	).Scan(&family.ID, &family.CreatedAt)
}

func (r *FamilyRepo) GetFamilyByID(ctx context.Context, id, familyID string) (*models.Family, error) {
	var f models.Family
	err := r.db.GetContext(ctx, &f,
		`SELECT id, name, slug, logo_url, plan, created_at FROM families WHERE id = $1 AND id = $2`,
		id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FamilyRepo) GetFamilyBySlug(ctx context.Context, slug string) (*models.Family, error) {
	var f models.Family
	err := r.db.GetContext(ctx, &f,
		`SELECT id, name, slug, logo_url, plan, created_at FROM families WHERE slug = $1`,
		slug,
	)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// GetFamilyBySlugOrName resolves a family by slug first, then by case-insensitive name.
// Name lookup is accepted for child login convenience; when multiple families share
// the same name, this returns an error so callers can safely require the slug.
func (r *FamilyRepo) GetFamilyBySlugOrName(ctx context.Context, identifier string) (*models.Family, error) {
	family, err := r.GetFamilyBySlug(ctx, identifier)
	if err == nil {
		return family, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	var matches []models.Family
	if err := r.db.SelectContext(ctx, &matches,
		`SELECT id, name, slug, logo_url, plan, created_at
		 FROM families
		 WHERE LOWER(name) = LOWER($1)
		 LIMIT 2`,
		identifier,
	); err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, sql.ErrNoRows
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple families matched this name")
	}

	return &matches[0], nil
}

func (r *FamilyRepo) UpdateFamily(ctx context.Context, id, familyID string, name string, logoURL *string) (*models.Family, error) {
	var f models.Family
	err := r.db.GetContext(ctx, &f,
		`UPDATE families SET name = $1, logo_url = $2
		 WHERE id = $3 AND id = $4
		 RETURNING id, name, slug, logo_url, plan, created_at`,
		name, logoURL, id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FamilyRepo) ListMembers(ctx context.Context, familyID string) ([]*models.User, error) {
	var users []*models.User
	err := r.db.SelectContext(ctx, &users,
		`SELECT id, family_id, role, name, username, email, password_hash, avatar_url,
		        theme, date_of_birth, game_limit_minutes, is_active, created_by, last_login_at, created_at
		 FROM users WHERE family_id = $1 ORDER BY created_at ASC`,
		familyID,
	)
	return users, err
}

func (r *FamilyRepo) CreateMember(ctx context.Context, user *models.User) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO users (family_id, role, name, username, email, password_hash, avatar_url,
		                    theme, date_of_birth, game_limit_minutes, is_active, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING id, created_at`,
		user.FamilyID, user.Role, user.Name, user.Username, user.Email, user.PasswordHash,
		user.AvatarURL, user.Theme, user.DateOfBirth, user.GameLimitMinutes, user.IsActive, user.CreatedBy,
	).Scan(&user.ID, &user.CreatedAt)
}

func (r *FamilyRepo) GetMemberByID(ctx context.Context, id, familyID string) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u,
		`SELECT id, family_id, role, name, username, email, password_hash, avatar_url,
		        theme, date_of_birth, game_limit_minutes, is_active, created_by, last_login_at, created_at
		 FROM users WHERE id = $1 AND family_id = $2`,
		id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *FamilyRepo) UpdateMember(ctx context.Context, id, familyID string, updates map[string]interface{}) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u,
		`UPDATE users
		 SET name = COALESCE($1, name),
		     theme = COALESCE($2, theme),
		     game_limit_minutes = COALESCE($3, game_limit_minutes),
		     date_of_birth = COALESCE($4, date_of_birth)
		 WHERE id = $5 AND family_id = $6
		 RETURNING id, family_id, role, name, username, email, password_hash, avatar_url,
		           theme, date_of_birth, game_limit_minutes, is_active, created_by, last_login_at, created_at`,
		updates["name"], updates["theme"], updates["game_limit_minutes"], updates["date_of_birth"],
		id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *FamilyRepo) DeactivateMember(ctx context.Context, id, familyID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_active = FALSE WHERE id = $1 AND family_id = $2`,
		id, familyID,
	)
	return err
}

func (r *FamilyRepo) GetPermissions(ctx context.Context, userID, familyID string) ([]string, error) {
	var raw []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT permissions FROM family_access_control WHERE grantee_id = $1 AND family_id = $2`,
		userID, familyID,
	).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var perms []string
	if err := json.Unmarshal(raw, &perms); err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *FamilyRepo) SetPermissions(ctx context.Context, granteeID, familyID string, grantorID uuid.UUID, perms []string) (*models.FamilyAccessControl, error) {
	permsJSON, err := json.Marshal(perms)
	if err != nil {
		return nil, err
	}

	var fac models.FamilyAccessControl
	var rawPerms []byte
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO family_access_control (family_id, grantor_id, grantee_id, permissions)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (grantor_id, grantee_id)
		 DO UPDATE SET permissions = EXCLUDED.permissions
		 RETURNING id, family_id, grantor_id, grantee_id, permissions, created_at`,
		familyID, grantorID, granteeID, permsJSON,
	)
	err = row.Scan(&fac.ID, &fac.FamilyID, &fac.GrantorID, &fac.GranteeID, &rawPerms, &fac.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rawPerms, &fac.Permissions); err != nil {
		return nil, err
	}
	return &fac, nil
}

func (r *FamilyRepo) RevokePermissions(ctx context.Context, granteeID, familyID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM family_access_control WHERE grantee_id = $1 AND family_id = $2`,
		granteeID, familyID,
	)
	return err
}

func (r *FamilyRepo) ListAccessControl(ctx context.Context, familyID string) ([]*models.FamilyAccessControl, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, family_id, grantor_id, grantee_id, permissions, created_at
		 FROM family_access_control WHERE family_id = $1`,
		familyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.FamilyAccessControl
	for rows.Next() {
		var fac models.FamilyAccessControl
		var rawPerms []byte
		if err := rows.Scan(&fac.ID, &fac.FamilyID, &fac.GrantorID, &fac.GranteeID, &rawPerms, &fac.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rawPerms, &fac.Permissions); err != nil {
			return nil, err
		}
		results = append(results, &fac)
	}
	return results, rows.Err()
}

func (r *FamilyRepo) UpdateLogoURL(ctx context.Context, familyID, logoURL string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE families SET logo_url = $1 WHERE id = $2`,
		logoURL, familyID,
	)
	return err
}

func (r *FamilyRepo) GetRantCount(ctx context.Context, childID, familyID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*)
		 FROM rants r
		 JOIN users u ON u.id = r.user_id
		 WHERE r.user_id = $1 AND u.family_id = $2`,
		childID, familyID,
	).Scan(&count)
	return count, err
}
