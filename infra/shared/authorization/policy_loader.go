package authorization

import (
	"fmt"
	"os"

	appErrors "github.com/nahualventure/class-backend/core/app/shared/errors"

	"github.com/casbin/casbin/v2"
	"gopkg.in/yaml.v3"
)

// PolicyConfig represents the structure of the policies.yaml file
type PolicyConfig struct {
	Roles map[string]RoleConfig `yaml:"roles"`
}

// RoleConfig represents a role and its permissions
type RoleConfig struct {
	Permissions map[string][]string `yaml:"permissions"`
}

// PolicyLoader handles loading and converting policies from YAML
type PolicyLoader struct {
	config *PolicyConfig
}

// NewPolicyLoader creates a new policy loader
func NewPolicyLoader() *PolicyLoader {
	return &PolicyLoader{}
}

func (p *PolicyLoader) LoadFromFile(filePath string) *appErrors.InfrastructureError {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return appErrors.NewInfrastructureError(fmt.Sprintf("failed to read policy file %s", filePath), err)
	}
	return p.LoadFromBytes(data)
}

// LoadFromBytes loads policy configuration from YAML bytes
func (p *PolicyLoader) LoadFromBytes(data []byte) *appErrors.InfrastructureError {
	config := &PolicyConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return appErrors.NewInfrastructureError("failed to parse YAML policy config", err)
	}

	p.config = config
	return nil
}

// LoadPoliciesIntoEnforcer loads policies into the Casbin enforcer for specified tenants
// Converts human-readable "all" keywords to Casbin "*" wildcards
func (p *PolicyLoader) LoadPoliciesIntoEnforcer(enforcer *casbin.Enforcer, tenants []string) *appErrors.InfrastructureError {
	if p.config == nil {
		return appErrors.NewInfrastructureError("policy config not loaded", nil)
	}

	// Clear existing policies (not role assignments)
	enforcer.ClearPolicy()

	// Generate policies for each role and tenant combination
	for roleName, roleConfig := range p.config.Roles {
		for _, tenantID := range tenants {
			if err := p.addRolePoliciesForTenant(enforcer, roleName, roleConfig, tenantID); err != nil {
				return err
			}
		}
	}

	return nil
}

// addRolePoliciesForTenant adds all policies for a specific role in a specific tenant
func (p *PolicyLoader) addRolePoliciesForTenant(enforcer *casbin.Enforcer, roleName string, roleConfig RoleConfig, tenantID string) *appErrors.InfrastructureError {
	for resource, actions := range roleConfig.Permissions {
		// Convert human-readable "all" to Casbin wildcard "*"
		casbinResource := p.convertToCasbinWildcard(resource)

		for _, action := range actions {
			// Convert human-readable "all" to Casbin wildcard "*"
			casbinAction := p.convertToCasbinWildcard(action)

			// Add policy: role, resource, action, tenant
			if _, err := enforcer.AddPolicy(roleName, casbinResource, casbinAction, tenantID); err != nil {
				return appErrors.NewInfrastructureError(
					fmt.Sprintf("failed to add policy [%s, %s, %s, %s]", roleName, casbinResource, casbinAction, tenantID),
					err)
			}
		}
	}

	return nil
}

// convertToCasbinWildcard converts human-readable "all" to Casbin wildcard "*"
func (p *PolicyLoader) convertToCasbinWildcard(value string) string {
	if value == "all" {
		return "*"
	}
	return value
}

// GetConfig returns the loaded policy configuration
func (p *PolicyLoader) GetConfig() *PolicyConfig {
	return p.config
}

// GetRoles returns all defined role names
func (p *PolicyLoader) GetRoles() []string {
	if p.config == nil {
		return nil
	}

	var roles []string
	for roleName := range p.config.Roles {
		roles = append(roles, roleName)
	}
	return roles
}

// ValidateYAMLConfig validates the loaded YAML configuration
func (p *PolicyLoader) ValidateYAMLConfig() *appErrors.InfrastructureError {
	if p.config == nil {
		return appErrors.NewInfrastructureError("no config loaded", nil)
	}

	if len(p.config.Roles) == 0 {
		return appErrors.NewInfrastructureError("no roles defined in config", nil)
	}

	// Validate each role has at least one permission
	for roleName, roleConfig := range p.config.Roles {
		if len(roleConfig.Permissions) == 0 {
			return appErrors.NewInfrastructureError(
				fmt.Sprintf("role '%s' has no permissions defined", roleName),
				nil)
		}

		// Validate each permission has at least one action
		for resource, actions := range roleConfig.Permissions {
			if len(actions) == 0 {
				return appErrors.NewInfrastructureError(
					fmt.Sprintf("role '%s' resource '%s' has no actions defined", roleName, resource),
					nil)
			}
		}
	}

	return nil
}

// PrintLoadedPolicies prints all loaded policies for debugging
func (p *PolicyLoader) PrintLoadedPolicies(enforcer *casbin.Enforcer) {
	fmt.Println("=== Loaded Policies ===")

	policies, err := enforcer.GetPolicy()
	if err != nil {
		fmt.Printf("error retrieving policies: %v\n", err)
	} else if len(policies) == 0 {
		fmt.Println("no policies found")
	} else {
		for _, policy := range policies {
			fmt.Printf("Policy: %v\n", policy)
		}
	}

	fmt.Println("\n=== Loaded Groupings ===")

	groupings, err := enforcer.GetGroupingPolicy()
	if err != nil {
		fmt.Printf("error retrieving groupings: %v\n", err)
	} else if len(groupings) == 0 {
		fmt.Println("no groupings found")
	} else {
		for _, grouping := range groupings {
			fmt.Printf("Grouping: %v\n", grouping)
		}
	}
}
