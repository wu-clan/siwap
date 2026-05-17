package desktop

import "github.com/wailsapp/wails/v3/pkg/application"

var (
	lightWindowBackground = application.NewRGBA(236, 237, 240, 255)
	darkWindowBackground  = application.NewRGBA(30, 31, 34, 255)
	lightWindowText       = application.NewRGBA(40, 42, 46, 255)
	darkWindowText        = application.NewRGBA(230, 231, 233, 255)
)

func windowBackgroundColour(appearance string, systemDark bool) application.RGBA {
	if isDarkWindowAppearance(appearance, systemDark) {
		return darkWindowBackground
	}
	return lightWindowBackground
}

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

func windowsWindowChrome(appearance string) application.WindowsWindow {
	return application.WindowsWindow{
		Theme:       windowsTheme(appearance),
		CustomTheme: windowsThemeSettings(),
	}
}

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

func menuBarTheme(bg application.RGBA, text application.RGBA) *application.MenuBarTheme {
	return &application.MenuBarTheme{
		Default:  textTheme(text, bg),
		Hover:    textTheme(text, bg),
		Selected: textTheme(text, bg),
	}
}

func textTheme(text application.RGBA, bg application.RGBA) *application.TextTheme {
	return &application.TextTheme{
		Text:       application.NewRGBPtr(text.Red, text.Green, text.Blue),
		Background: application.NewRGBPtr(bg.Red, bg.Green, bg.Blue),
	}
}

func isDarkWindowAppearance(appearance string, systemDark bool) bool {
	return appearance == "dark" || (appearance == "system" && systemDark)
}
