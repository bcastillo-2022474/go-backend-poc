package authorization

import (
	"database/sql"
	"fmt"
	"log"

	appErrors "class-backend/core/app/shared/errors"

	"github.com/casbin/casbin/v2"
	_ "github.com/lib/pq"
)

// CasbinService provides authorization functionality using Casbin
type CasbinService struct {
	enforcer     *casbin.Enforcer
	adapter      *RoleOnlyPostgresAdapter
	policyLoader *PolicyLoader
}

func NewCasbinService(db *sql.DB, modelPath, policiesPath string, tenants []string) (*CasbinService, *appErrors.InfrastructureError) {
	adapter, err := NewRoleOnlyPostgresAdapter(db)
	if err != nil {
		return nil, err
	}

	enforcer, normalErr := casbin.NewEnforcer(modelPath, adapter)
	if normalErr != nil {
		return nil, appErrors.NewInfrastructureError("failed to create Casbin enforcer", normalErr)
	}

	policyLoader := NewPolicyLoader()
	if err := policyLoader.LoadFromFile(policiesPath); err != nil {
		return nil, err
	}

	if err := policyLoader.ValidateYAMLConfig(); err != nil {
		return nil, err
	}

	if err := policyLoader.LoadPoliciesIntoEnforcer(enforcer, tenants); err != nil {
		return nil, err
	}

	enforcer.EnableAutoSave(true)

	service := &CasbinService{
		enforcer:     enforcer,
		adapter:      adapter,
		policyLoader: policyLoader,
	}

	log.Printf("CasbinService initialized with %d roles for %d tenants",
		len(policyLoader.GetRoles()), len(tenants))

	return service, nil
}

func (c *CasbinService) CanDo(userID, resource, action, tenantID string) (bool, *appErrors.InfrastructureError) {
	if userID == "" || resource == "" || action == "" || tenantID == "" {
		return false, appErrors.NewInfrastructureError(
			fmt.Sprintf("authorization parameters cannot be empty: userID=%s, resource=%s, action=%s, tenantID=%s", userID, resource, action, tenantID),
			nil,
		)
	}

	allowed, err := c.enforcer.Enforce(userID, resource, action, tenantID)
	if err != nil {
		log.Printf("authorization error for user %s: %v", userID, err)
		return false, appErrors.NewInfrastructureError(fmt.Sprintf("failed to enforce authorization for user %s", userID), err)
	}
	return allowed, nil
}

func (c *CasbinService) AssignRole(userID, role, tenantID string) *appErrors.InfrastructureError {
	if userID == "" || role == "" || tenantID == "" {
		return appErrors.NewInfrastructureError(
			fmt.Sprintf("role assignment parameters cannot be empty: userID=%s, role=%s, tenantID=%s", userID, role, tenantID),
			nil,
		)
	}

	// Check if role exists in available roles
	availableRoles := c.policyLoader.GetRoles()
	roleExists := false
	for _, availableRole := range availableRoles {
		if availableRole == role {
			roleExists = true
			break
		}
	}
	if !roleExists {
		return appErrors.NewInfrastructureError(
			fmt.Sprintf("role %s is not available in the system", role),
			nil,
		)
	}

	added, err := c.enforcer.AddGroupingPolicy(userID, role, tenantID)
	if err != nil {
		return appErrors.NewInfrastructureError(
			fmt.Sprintf("failed to assign role %s to user %s in tenant %s", role, userID, tenantID),
			err)
	}

	if added {
		log.Printf("role assigned: user=%s, role=%s, tenant=%s", userID, role, tenantID)
	} else {
		log.Printf("role assignment skipped (already exists): user=%s, role=%s, tenant=%s", userID, role, tenantID)
	}
	return nil
}

func (c *CasbinService) RemoveRole(userID, role, tenantID string) *appErrors.InfrastructureError {
	if userID == "" || role == "" || tenantID == "" {
		return appErrors.NewInfrastructureError(
			fmt.Sprintf("role removal parameters cannot be empty: userID=%s, role=%s, tenantID=%s", userID, role, tenantID),
			nil,
		)
	}

	removed, err := c.enforcer.RemoveGroupingPolicy(userID, role, tenantID)
	if err != nil {
		return appErrors.NewInfrastructureError(
			fmt.Sprintf("failed to remove role %s from user %s in tenant %s", role, userID, tenantID),
			err)
	}

	if removed {
		log.Printf("role removed: user=%s, role=%s, tenant=%s", userID, role, tenantID)
	} else {
		log.Printf("role removal skipped (not found): user=%s, role=%s, tenant=%s", userID, role, tenantID)
	}
	return nil
}

func (c *CasbinService) GetUserRoles(userID, tenantID string) ([]string, *appErrors.InfrastructureError) {
	if userID == "" || tenantID == "" {
		return nil, appErrors.NewInfrastructureError(
			fmt.Sprintf("user roles query parameters cannot be empty: userID=%s, tenantID=%s", userID, tenantID),
			nil,
		)
	}

	groupings, err := c.enforcer.GetGroupingPolicy()
	if err != nil {
		return nil, appErrors.NewInfrastructureError("failed to get grouping policies", err)
	}

	var roles []string
	for _, grouping := range groupings {
		if len(grouping) >= 3 && grouping[0] == userID && grouping[2] == tenantID {
			roles = append(roles, grouping[1])
		}
	}

	return roles, nil
}

// GetUserTenantsForRole returns all tenants where user has a specific role
func (c *CasbinService) GetUserTenantsForRole(userID, role string) ([]string, *appErrors.InfrastructureError) {
	if userID == "" || role == "" {
		return nil, appErrors.NewInfrastructureError(
			fmt.Sprintf("tenant query parameters cannot be empty: userID=%s, role=%s", userID, role),
			nil,
		)
	}

	groupings, err := c.enforcer.GetGroupingPolicy()
	if err != nil {
		return nil, appErrors.NewInfrastructureError("failed to get grouping policies", err)
	}

	var tenants []string
	for _, grouping := range groupings {
		if len(grouping) >= 3 && grouping[0] == userID && grouping[1] == role {
			tenants = append(tenants, grouping[2])
		}
	}

	return tenants, nil
}

func (c *CasbinService) HasRole(userID, role, tenantID string) (bool, *appErrors.InfrastructureError) {
	if userID == "" || role == "" || tenantID == "" {
		return false, appErrors.NewInfrastructureError(
			fmt.Sprintf("role check parameters cannot be empty: userID=%s, role=%s, tenantID=%s", userID, role, tenantID),
			nil,
		)
	}

	hasRole, err := c.enforcer.HasGroupingPolicy(userID, role, tenantID)
	if err != nil {
		return false, appErrors.NewInfrastructureError(
			fmt.Sprintf("failed to check if user %s has role %s in tenant %s", userID, role, tenantID),
			err)
	}

	return hasRole, nil
}

func (c *CasbinService) GetAvailableRoles() []string {
	return c.policyLoader.GetRoles()
}

// ReloadPolicies reloads policies from YAML for new tenants
func (c *CasbinService) ReloadPolicies(tenants []string) *appErrors.InfrastructureError {
	if len(tenants) == 0 {
		return appErrors.NewInfrastructureError(
			"tenants list cannot be empty for policy reload",
			nil,
		)
	}

	err := c.policyLoader.LoadPoliciesIntoEnforcer(c.enforcer, tenants)
	if err != nil {
		return err
	}

	log.Printf("policies reloaded successfully for %d tenants", len(tenants))
	return nil
}

func (c *CasbinService) PrintDebugInfo() {
	fmt.Println("\n=== Casbin Debug Info ===")
	if c.policyLoader != nil && c.enforcer != nil {
		c.policyLoader.PrintLoadedPolicies(c.enforcer)
	} else {
		fmt.Println("Error: Service not properly initialized")
	}
	fmt.Println("=========================")
}

func (c *CasbinService) GetEnforcer() *casbin.Enforcer {
	return c.enforcer
}

func (c *CasbinService) Close() error {
	log.Println("CasbinService closed")
	return nil
}
