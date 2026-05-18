package desktop

import "github.com/wailsapp/wails/v3/pkg/application"

var (
	lightWindowBackground = application.NewRGBA(236, 237, 240, 255)
	darkWindowBackground  = application.NewRGBA(30, 31, 34, 255)
	lightWindowText       = application.NewRGBA(40, 42, 46, 255)
	darkWindowText        = application.NewRGBA(230, 231, 233, 255)
)

// windowBackgroundColour 返回桌面窗口背景色
func windowBackgroundColour(appearance string, systemDark bool) application.RGBA {
	if isDarkWindowAppearance(appearance, systemDark) {
		return darkWindowBackground
	}
	return lightWindowBackground
}

// macWindowChrome 构造 macOS 窗口外观配置
func macWindowChrome(appearance string) application.MacWindow {
	return application.MacWindow{
		TitleBar: application.MacTitleBar{
			AppearsTransparent:   false,
			HideTitle:            true,
			HideToolbarSeparator: true,
		},
		Appearance: macAppearance(appearance),
	}
}

// windowsWindowChrome 构造 Windows 窗口外观配置
func windowsWindowChrome(appearance string) application.WindowsWindow {
	return application.WindowsWindow{
		Theme:       windowsTheme(appearance),
		CustomTheme: windowsThemeSettings(),
	}
}

// macAppearance 将偏好设置转换为 macOS 外观模式
func macAppearance(appearance string) application.MacAppearanceType {
	switch appearance {
	case "dark":
		return application.NSAppearanceNameDarkAqua
	case "light":
		return application.NSAppearanceNameAqua
	default:
		return application.DefaultAppearance
	}
}

// windowsTheme 将偏好设置转换为 Windows 主题模式
func windowsTheme(appearance string) application.Theme {
	switch appearance {
	case "dark":
		return application.Dark
	case "light":
		return application.Light
	default:
		return application.SystemDefault
	}
}

// windowsThemeSettings 构造 Windows 主题配置
func windowsThemeSettings() application.ThemeSettings {
	return application.ThemeSettings{
		LightModeActive:   windowTheme(lightWindowBackground, lightWindowText),
		LightModeInactive: windowTheme(lightWindowBackground, lightWindowText),
		DarkModeActive:    windowTheme(darkWindowBackground, darkWindowText),
		DarkModeInactive:  windowTheme(darkWindowBackground, darkWindowText),
		LightModeMenuBar:  menuBarTheme(lightWindowBackground, lightWindowText),
		DarkModeMenuBar:   menuBarTheme(darkWindowBackground, darkWindowText),
	}
}

// windowTheme 返回当前窗口主题配置
func windowTheme(bg application.RGBA, text application.RGBA) *application.WindowTheme {
	return &application.WindowTheme{
		BorderColour:   application.NewRGBPtr(bg.Red, bg.Green, bg.Blue),
		TitleBarColour: application.NewRGBPtr(bg.Red, bg.Green, bg.Blue),
		TitleTextColour: application.NewRGBPtr(
			text.Red,
			text.Green,
			text.Blue,
		),
	}
}

// menuBarTheme 返回菜单栏主题配置
func menuBarTheme(bg application.RGBA, text application.RGBA) *application.MenuBarTheme {
	return &application.MenuBarTheme{
		Default:  textTheme(text, bg),
		Hover:    textTheme(text, bg),
		Selected: textTheme(text, bg),
	}
}

// textTheme 根据明暗模式返回文本主题色
func textTheme(text application.RGBA, bg application.RGBA) *application.TextTheme {
	return &application.TextTheme{
		Text:       application.NewRGBPtr(text.Red, text.Green, text.Blue),
		Background: application.NewRGBPtr(bg.Red, bg.Green, bg.Blue),
	}
}

// isDarkWindowAppearance 判断当前外观是否为深色模式
func isDarkWindowAppearance(appearance string, systemDark bool) bool {
	return appearance == "dark" || (appearance == "system" && systemDark)
}
