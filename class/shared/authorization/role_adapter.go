package authorization

import (
	"database/sql"
	"fmt"
	"strings"

	appErrors "class-backend/core/app/shared/errors"

	"github.com/Blank-Xu/sql-adapter"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// RoleOnlyPostgresAdapter extends sql-adapter to only persist role assignments (g records)
// Policies (p records) are managed in memory and not persisted to database
type RoleOnlyPostgresAdapter struct {
	*sqladapter.Adapter
	db *sql.DB
}

// NewRoleOnlyPostgresAdapter creates a new adapter that only persists role assignments
func NewRoleOnlyPostgresAdapter(db *sql.DB) (*RoleOnlyPostgresAdapter, *appErrors.InfrastructureError) {
	// Create base SQL adapter with custom table name
	baseAdapter, err := sqladapter.NewAdapter(db, "postgres", "casbin_rule")
	if err != nil {
		return nil, appErrors.NewInfrastructureError("failed to create base SQL adapter", err)
	}

	return &RoleOnlyPostgresAdapter{
		Adapter: baseAdapter,
		db:      db,
	}, nil
}

// LoadPolicy loads only role assignments (g records) from database
// Policies (p records) are intentionally skipped as they are managed in memory
func (a *RoleOnlyPostgresAdapter) LoadPolicy(model model.Model) error {
	// Only load grouping policies (role assignments)
	// Skip regular policies (p, p2, etc.) as they are defined in code

	// Query only for grouping policies (ptype starting with 'g')
	query := "SELECT ptype, v0, v1, v2, v3, v4, v5 FROM casbin_rule WHERE ptype LIKE 'g%'"

	rows, err := a.db.Query(query)
	if err != nil {
		return appErrors.NewInfrastructureError("failed to query role assignments from database", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ptype, v0, v1, v2, v3, v4, v5 sql.NullString
		if err := rows.Scan(&ptype, &v0, &v1, &v2, &v3, &v4, &v5); err != nil {
			return appErrors.NewInfrastructureError("failed to scan role assignment row", err)
		}

		// Build rule from non-null values
		var rule []string
		for _, v := range []sql.NullString{v0, v1, v2, v3, v4, v5} {
			if v.Valid && v.String != "" {
				rule = append(rule, v.String)
			}
		}

		if len(rule) > 0 {
			// Add to model - this will be a grouping policy
			persist.LoadPolicyLine(ptype.String, model)
			if len(rule) > 0 {
				model[ptype.String][ptype.String].Policy = append(model[ptype.String][ptype.String].Policy, rule)
			}
		}
	}

	if err := rows.Err(); err != nil {
		return appErrors.NewInfrastructureError("error iterating role assignments result set", err)
	}

	return nil
}

// SavePolicy saves only role assignments (g records) to database
// Policies (p records) are intentionally skipped
func (a *RoleOnlyPostgresAdapter) SavePolicy(model model.Model) error {
	// Clear existing role assignments
	if _, err := a.db.Exec("DELETE FROM casbin_rule WHERE ptype LIKE 'g%'"); err != nil {
		return appErrors.NewInfrastructureError("failed to clear existing role assignments", err)
	}

	// Save only grouping policies (role assignments)
	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			if err := a.insertRoleAssignment(a.db, ptype, rule); err != nil {
				return appErrors.NewInfrastructureError(
					fmt.Sprintf("failed to save role assignment %s %v", ptype, rule),
					err)
			}
		}
	}

	return nil
}

// AddPolicy adds a policy rule - only processes role assignments
func (a *RoleOnlyPostgresAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	// Only handle grouping policies (role assignments)
	if sec != "g" {
		return nil // Silently ignore non-grouping policies
	}

	if err := a.insertRoleAssignment(a.db, ptype, rule); err != nil {
		return err.Unwrap()
	}
	return nil
}

// RemovePolicy removes a policy rule - only processes role assignments
func (a *RoleOnlyPostgresAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	// Only handle grouping policies (role assignments)
	if sec != "g" {
		return nil // Silently ignore non-grouping policies
	}

	// Build WHERE conditions for the rule
	conditions := []string{"ptype = $1"}
	args := []interface{}{ptype}

	for i, value := range rule {
		if i < 6 { // v0-v5
			conditions = append(conditions, fmt.Sprintf("v%d = $%d", i, i+2))
			args = append(args, value)
		}
	}

	query := fmt.Sprintf("DELETE FROM casbin_rule WHERE %s", strings.Join(conditions, " AND "))
	_, err := a.db.Exec(query, args...)

	if err != nil {
		return fmt.Errorf("failed to remove role assignment from database: %w", err)
	}

	return nil
}

// insertRoleAssignment inserts a single role assignment into the database
func (a *RoleOnlyPostgresAdapter) insertRoleAssignment(db *sql.DB, ptype string, rule []string) *appErrors.InfrastructureError {
	// Prepare values (up to 6 values: v0-v5)
	values := make([]interface{}, 7) // ptype + v0-v5
	values[0] = ptype

	for i := 0; i < 6; i++ {
		if i < len(rule) {
			values[i+1] = rule[i]
		} else {
			values[i+1] = ""
		}
	}

	query := `
		INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := db.Exec(query, values...)
	if err != nil {
		return appErrors.NewInfrastructureError("failed to insert role assignment into database", err)
	}

	return nil
}

// GetDB returns the underlying database connection
func (a *RoleOnlyPostgresAdapter) GetDB() *sql.DB {
	return a.db
}
