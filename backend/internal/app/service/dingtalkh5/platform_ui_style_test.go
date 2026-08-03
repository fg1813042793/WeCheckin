package dingtalkh5

import (
	"os"
	"strings"
	"testing"
)

func TestDingTalkH5UsesDingTalkPlatformLayoutStyle(t *testing.T) {
	appShellSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/AppShell.vue")
	if err != nil {
		t.Fatalf("read app shell: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read h5 styles: %v", err)
	}
	combined := string(appShellSrc) + string(styleSrc)
	for _, snippet := range []string{
		`class="app-topbar"`,
		`grid-template-rows: 56px minmax(0, 1fr);`,
		`grid-template-areas:`,
		`grid-area: topbar;`,
		`grid-area: sidebar;`,
		`grid-area: main;`,
		`background: #fff;`,
		`class="sidebar-menu-caption"`,
		`class="nav-children flat"`,
		`class="desktop-user-pill avatar-only"`,
		`class="layout-switch-btn icon-only"`,
		`.desktop-user-pill.avatar-only {`,
		`background: transparent;`,
		`box-shadow: none;`,
		`border-radius: 4px;`,
		`height: 32px;`,
		`background: #f2f3f5;`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("DingTalk platform UI style must include %q", snippet)
		}
	}
	for _, forbidden := range []string{
		`background: linear-gradient(180deg, #ffffff 0%, #fbfdff 100%);`,
		`class="brand-line sidebar-brand"`,
		`class="desktop-user-main"`,
		`class="desktop-user-caret"`,
		`class="sidebar-layout-btn icon-only"`,
		`@click="setMenuLayout('top')"`,
		`<text>{{ menuLayout === 'side' ? '顶部菜单' : '左侧菜单' }}</text>`,
		`<text class="layout-label">顶部菜单</text>`,
		`box-shadow: 0 10px 30px rgba(31, 35, 41, 0.045);`,
		`border-radius: 16px;`,
	} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("DingTalk platform UI style should avoid legacy soft-card style %q", forbidden)
		}
	}
}

func TestDingTalkH5TopMenuAlignsMenuItemsToLeft(t *testing.T) {
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read h5 styles: %v", err)
	}
	style := string(styleSrc)
	for _, snippet := range []string{
		`.top-nav-list {`,
		`justify-content: flex-start;`,
		`.top-nav-item {`,
		`.top-submenu-item {`,
		`margin: 0;`,
		`.top-submenu {`,
	} {
		if !strings.Contains(style, snippet) {
			t.Fatalf("top menu alignment style must include %q", snippet)
		}
	}
	if strings.Contains(style, `.top-nav-item {
  min-height: 56px;
  padding: 0 14px;`) {
		t.Fatalf("top nav item must reset uni button margin before padding")
	}
	if strings.Contains(style, `.top-submenu-item {
  height: 34px;
  min-height: 34px;
  padding: 0 14px;`) {
		t.Fatalf("top submenu item must reset uni button margin before padding")
	}
}

func TestDingTalkH5PcOpenedTabsUsePolishedChrome(t *testing.T) {
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read h5 styles: %v", err)
	}
	style := string(styleSrc)
	start := strings.Index(style, `.desktop-page-tabs {`)
	if start < 0 {
		t.Fatalf("pc opened tabs style section is missing")
	}
	end := strings.Index(style[start:], `.desktop-topbar {`)
	if end < 0 {
		t.Fatalf("pc opened tabs style section must appear before desktop topbar styles")
	}
	tabsStyle := style[start : start+end]
	for _, snippet := range []string{
		`.desktop-page-tabs {`,
		`background: #f7f9fc;`,
		`box-shadow: inset 0 -1px 0 #e5eaf3;`,
		`.desktop-page-tab + .desktop-page-tab::before {`,
		`background: #dfe5ef;`,
		`border-radius: 6px 6px 0 0;`,
		`border-bottom-color: #fff;`,
		`box-shadow: 0 -1px 0 rgba(22, 119, 255, 0.04);`,
		`.desktop-page-tab-close::after {`,
		`display: none;`,
	} {
		if !strings.Contains(tabsStyle, snippet) {
			t.Fatalf("pc opened tabs polished style must include %q", snippet)
		}
	}
	for _, legacy := range []string{
		`gap: 2px;`,
		`background: #f5f7fb;`,
		`border-radius: 3px;`,
	} {
		if strings.Contains(tabsStyle, legacy) {
			t.Fatalf("pc opened tabs polished style should remove rough legacy style %q", legacy)
		}
	}
}

func TestDingTalkH5SidebarUsesCompactWidth(t *testing.T) {
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read h5 styles: %v", err)
	}
	style := string(styleSrc)
	for _, snippet := range []string{
		`grid-template-columns: 196px minmax(0, 1fr);`,
		`grid-template-columns: 188px minmax(0, 1fr);`,
		`padding: 16px 8px;`,
		`padding: 0 8px 8px;`,
		`.nav-item {
  width: 100%;
  height: 34px;
  padding: 0 10px;`,
		`margin: 2px 0 6px 32px;`,
	} {
		if !strings.Contains(style, snippet) {
			t.Fatalf("compact sidebar style must include %q", snippet)
		}
	}
	for _, legacy := range []string{
		`grid-template-columns: 220px minmax(0, 1fr);`,
		`grid-template-columns: 216px minmax(0, 1fr);`,
		`padding: 16px 10px;`,
		`padding: 0 10px 8px;`,
		`.nav-item {
  width: 100%;
  height: 34px;
  padding: 0 12px;`,
		`margin: 2px 0 6px 36px;`,
	} {
		if strings.Contains(style, legacy) {
			t.Fatalf("compact sidebar style should not keep legacy spacing %q", legacy)
		}
	}
}

func TestDingTalkH5ProfileMenuClosesWhenClickingOutside(t *testing.T) {
	appShellSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/AppShell.vue")
	if err != nil {
		t.Fatalf("read app shell: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read h5 styles: %v", err)
	}
	combined := string(appShellSrc) + string(styleSrc)
	for _, snippet := range []string{
		`class="desktop-profile-backdrop"`,
		`@click="closeProfileMenu"`,
		`function closeProfileMenu()`,
		`.desktop-profile-backdrop {`,
		`position: fixed;`,
		`inset: 0;`,
		`z-index: 38;`,
		`.desktop-profile-dropdown {`,
		`z-index: 40;`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("profile menu outside-click behavior must include %q", snippet)
		}
	}
}
