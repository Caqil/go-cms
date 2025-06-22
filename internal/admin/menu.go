package admin

import (
	"sort"

	"go-cms/internal/plugins"
)

type MenuManager struct {
	pluginManager *plugins.Manager
	baseMenuItems []plugins.AdminMenuItem
}

func NewMenuManager(pluginManager *plugins.Manager) *MenuManager {
	return &MenuManager{
		pluginManager: pluginManager,
		baseMenuItems: getBaseMenuItems(),
	}
}

func (m *MenuManager) GetFullMenu() []plugins.AdminMenuItem {
	// Start with base menu items
	allItems := make([]plugins.AdminMenuItem, len(m.baseMenuItems))
	copy(allItems, m.baseMenuItems)

	// Add plugin menu items
	pluginItems := m.pluginManager.GetAdminMenuItems()
	allItems = append(allItems, pluginItems...)

	// Sort by order
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].Order < allItems[j].Order
	})

	// Sort children within each parent
	for i := range allItems {
		if len(allItems[i].Children) > 0 {
			sort.Slice(allItems[i].Children, func(a, b int) bool {
				return allItems[i].Children[a].Order < allItems[i].Children[b].Order
			})
		}
	}

	return allItems
}

func (m *MenuManager) GetMenuByRole(role string) []plugins.AdminMenuItem {
	fullMenu := m.GetFullMenu()

	// Filter menu items based on user role
	var filteredMenu []plugins.AdminMenuItem

	for _, item := range fullMenu {
		if m.hasAccess(item, role) {
			// Filter children as well
			var filteredChildren []plugins.AdminMenuItem
			for _, child := range item.Children {
				if m.hasAccess(child, role) {
					filteredChildren = append(filteredChildren, child)
				}
			}
			item.Children = filteredChildren
			filteredMenu = append(filteredMenu, item)
		}
	}

	return filteredMenu
}

func (m *MenuManager) hasAccess(item plugins.AdminMenuItem, role string) bool {
	// For now, basic role checking
	// In a real implementation, you might have more complex permission logic

	switch role {
	case "super_admin":
		return true // Super admin has access to everything
	case "admin":
		// Admins can access most things except super admin specific items
		return !m.isSuperAdminOnly(item)
	default:
		return false // Regular users don't have admin access
	}
}

func (m *MenuManager) isSuperAdminOnly(item plugins.AdminMenuItem) bool {
	// Define which menu items are super admin only
	superAdminOnlyItems := map[string]bool{
		"user-management": true,
		"system-settings": true,
		"plugin-upload":   true, // Only super admins can upload plugins
	}

	return superAdminOnlyItems[item.ID]
}

func getBaseMenuItems() []plugins.AdminMenuItem {
	return []plugins.AdminMenuItem{
		{
			ID:    "dashboard",
			Title: "Dashboard",
			Icon:  "tachometer-alt",
			URL:   "/admin/dashboard",
			Order: 1,
		},
		{
			ID:    "content",
			Title: "Content",
			Icon:  "file-alt",
			URL:   "/admin/content",
			Order: 10,
			Children: []plugins.AdminMenuItem{
				{
					ID:    "content-all",
					Title: "All Content",
					URL:   "/admin/content/all",
					Order: 1,
				},
				{
					ID:    "content-create",
					Title: "Create New",
					URL:   "/admin/content/create",
					Order: 2,
				},
				{
					ID:    "content-categories",
					Title: "Categories",
					URL:   "/admin/content/categories",
					Order: 3,
				},
			},
		},
		{
			ID:    "media",
			Title: "Media",
			Icon:  "images",
			URL:   "/admin/media",
			Order: 15,
		},
		{
			ID:    "users",
			Title: "Users",
			Icon:  "users",
			URL:   "/admin/users",
			Order: 20,
			Children: []plugins.AdminMenuItem{
				{
					ID:    "users-all",
					Title: "All Users",
					URL:   "/admin/users/all",
					Order: 1,
				},
				{
					ID:    "users-roles",
					Title: "Roles & Permissions",
					URL:   "/admin/users/roles",
					Order: 2,
				},
			},
		},
		{
			ID:    "plugins",
			Title: "Plugins",
			Icon:  "puzzle-piece",
			URL:   "/admin/plugins",
			Order: 80,
			Children: []plugins.AdminMenuItem{
				{
					ID:    "plugins-installed",
					Title: "Installed Plugins",
					URL:   "/admin/plugins/installed",
					Order: 1,
				},
				{
					ID:    "plugin-upload",
					Title: "Upload Plugin",
					URL:   "/admin/plugins/upload",
					Order: 2,
				},
				{
					ID:    "plugin-marketplace",
					Title: "Plugin Marketplace",
					URL:   "/admin/plugins/marketplace",
					Order: 3,
				},
			},
		},
		{
			ID:    "themes",
			Title: "Appearance",
			Icon:  "palette",
			URL:   "/admin/themes",
			Order: 85,
			Children: []plugins.AdminMenuItem{
				{
					ID:    "themes-all",
					Title: "Themes",
					URL:   "/admin/themes/all",
					Order: 1,
				},
				{
					ID:    "themes-customize",
					Title: "Customize",
					URL:   "/admin/themes/customize",
					Order: 2,
				},
				{
					ID:    "themes-menus",
					Title: "Menus",
					URL:   "/admin/themes/menus",
					Order: 3,
				},
			},
		},
		{
			ID:    "settings",
			Title: "Settings",
			Icon:  "cog",
			URL:   "/admin/settings",
			Order: 90,
			Children: []plugins.AdminMenuItem{
				{
					ID:    "settings-general",
					Title: "General",
					URL:   "/admin/settings/general",
					Order: 1,
				},
				{
					ID:    "settings-security",
					Title: "Security",
					URL:   "/admin/settings/security",
					Order: 2,
				},
				{
					ID:    "settings-email",
					Title: "Email",
					URL:   "/admin/settings/email",
					Order: 3,
				},
			},
		},
		{
			ID:    "tools",
			Title: "Tools",
			Icon:  "tools",
			URL:   "/admin/tools",
			Order: 95,
			Children: []plugins.AdminMenuItem{
				{
					ID:    "tools-import",
					Title: "Import",
					URL:   "/admin/tools/import",
					Order: 1,
				},
				{
					ID:    "tools-export",
					Title: "Export",
					URL:   "/admin/tools/export",
					Order: 2,
				},
				{
					ID:    "tools-backup",
					Title: "Backup",
					URL:   "/admin/tools/backup",
					Order: 3,
				},
			},
		},
	}
}
