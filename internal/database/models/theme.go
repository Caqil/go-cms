package models

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ThemeMetadata represents theme information stored in the database
type ThemeMetadata struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name            string             `bson:"name" json:"name" binding:"required"`
	Version         string             `bson:"version" json:"version"`
	Description     string             `bson:"description" json:"description"`
	Author          string             `bson:"author" json:"author"`
	AuthorURI       string             `bson:"author_uri,omitempty" json:"author_uri,omitempty"`
	Website         string             `bson:"website,omitempty" json:"website,omitempty"`
	Screenshot      string             `bson:"screenshot,omitempty" json:"screenshot,omitempty"`
	Tags            []string           `bson:"tags,omitempty" json:"tags,omitempty"`
	MinVersion      string             `bson:"min_version,omitempty" json:"min_version,omitempty"`
	RequiredPlugins []string           `bson:"required_plugins,omitempty" json:"required_plugins,omitempty"`
	Path            string             `bson:"path" json:"path"`
	IsActive        bool               `bson:"is_active" json:"is_active"`
	Assets          ThemeAssets        `bson:"assets" json:"assets"`
	Templates       []ThemeTemplate    `bson:"templates,omitempty" json:"templates,omitempty"`
	Customization   ThemeCustomization `bson:"customization" json:"customization"`
	InstalledAt     time.Time          `bson:"installed_at" json:"installed_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
	InstalledBy     string             `bson:"installed_by,omitempty" json:"installed_by,omitempty"`
	Status          string             `bson:"status" json:"status"` // active, inactive, broken, updating
}

// ThemeAssets represents theme asset files
type ThemeAssets struct {
	CSS        []string          `bson:"css,omitempty" json:"css,omitempty"`
	JS         []string          `bson:"js,omitempty" json:"js,omitempty"`
	Images     []string          `bson:"images,omitempty" json:"images,omitempty"`
	Fonts      []string          `bson:"fonts,omitempty" json:"fonts,omitempty"`
	Other      map[string]string `bson:"other,omitempty" json:"other,omitempty"`
	Screenshot string            `bson:"screenshot,omitempty" json:"screenshot,omitempty"`
}

// ThemeTemplate represents theme template files
type ThemeTemplate struct {
	Name        string            `bson:"name" json:"name"`
	File        string            `bson:"file" json:"file"`
	Description string            `bson:"description,omitempty" json:"description,omitempty"`
	Type        string            `bson:"type" json:"type"` // page, post, archive, single, index, etc.
	IsDefault   bool              `bson:"is_default" json:"is_default"`
	Variables   map[string]string `bson:"variables,omitempty" json:"variables,omitempty"`
	Regions     []TemplateRegion  `bson:"regions,omitempty" json:"regions,omitempty"`
}

// TemplateRegion represents customizable regions in templates
type TemplateRegion struct {
	ID          string `bson:"id" json:"id"`
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	Type        string `bson:"type" json:"type"` // content, sidebar, header, footer, etc.
	Required    bool   `bson:"required" json:"required"`
}

// ThemeCustomization represents user customizations for a theme
type ThemeCustomization struct {
	Colors      map[string]string      `bson:"colors,omitempty" json:"colors,omitempty"`
	Fonts       map[string]string      `bson:"fonts,omitempty" json:"fonts,omitempty"`
	Layout      map[string]interface{} `bson:"layout,omitempty" json:"layout,omitempty"`
	Spacing     map[string]string      `bson:"spacing,omitempty" json:"spacing,omitempty"`
	Typography  TypographySettings     `bson:"typography,omitempty" json:"typography,omitempty"`
	CustomCSS   string                 `bson:"custom_css,omitempty" json:"custom_css,omitempty"`
	CustomJS    string                 `bson:"custom_js,omitempty" json:"custom_js,omitempty"`
	HeaderCode  string                 `bson:"header_code,omitempty" json:"header_code,omitempty"`
	FooterCode  string                 `bson:"footer_code,omitempty" json:"footer_code,omitempty"`
	Logo        LogoSettings           `bson:"logo,omitempty" json:"logo,omitempty"`
	Favicon     string                 `bson:"favicon,omitempty" json:"favicon,omitempty"`
	SocialMedia SocialMediaSettings    `bson:"social_media,omitempty" json:"social_media,omitempty"`
	SEO         SEOSettings            `bson:"seo,omitempty" json:"seo,omitempty"`
	Analytics   AnalyticsSettings      `bson:"analytics,omitempty" json:"analytics,omitempty"`
}

// TypographySettings represents typography customization
type TypographySettings struct {
	PrimaryFont   FontSettings `bson:"primary_font,omitempty" json:"primary_font,omitempty"`
	SecondaryFont FontSettings `bson:"secondary_font,omitempty" json:"secondary_font,omitempty"`
	HeadingFont   FontSettings `bson:"heading_font,omitempty" json:"heading_font,omitempty"`
	BaseFontSize  string       `bson:"base_font_size,omitempty" json:"base_font_size,omitempty"`
	LineHeight    string       `bson:"line_height,omitempty" json:"line_height,omitempty"`
}

// FontSettings represents font configuration
type FontSettings struct {
	Family     string `bson:"family,omitempty" json:"family,omitempty"`
	Weight     string `bson:"weight,omitempty" json:"weight,omitempty"`
	Style      string `bson:"style,omitempty" json:"style,omitempty"`
	Size       string `bson:"size,omitempty" json:"size,omitempty"`
	GoogleFont bool   `bson:"google_font,omitempty" json:"google_font,omitempty"`
	FontURL    string `bson:"font_url,omitempty" json:"font_url,omitempty"`
}

// LogoSettings represents logo configuration
type LogoSettings struct {
	URL      string `bson:"url,omitempty" json:"url,omitempty"`
	Alt      string `bson:"alt,omitempty" json:"alt,omitempty"`
	Width    string `bson:"width,omitempty" json:"width,omitempty"`
	Height   string `bson:"height,omitempty" json:"height,omitempty"`
	Position string `bson:"position,omitempty" json:"position,omitempty"`   // left, center, right
	DarkMode string `bson:"dark_mode,omitempty" json:"dark_mode,omitempty"` // URL for dark mode logo
}

// SocialMediaSettings represents social media integration
type SocialMediaSettings struct {
	Facebook     string `bson:"facebook,omitempty" json:"facebook,omitempty"`
	Twitter      string `bson:"twitter,omitempty" json:"twitter,omitempty"`
	Instagram    string `bson:"instagram,omitempty" json:"instagram,omitempty"`
	LinkedIn     string `bson:"linkedin,omitempty" json:"linkedin,omitempty"`
	YouTube      string `bson:"youtube,omitempty" json:"youtube,omitempty"`
	GitHub       string `bson:"github,omitempty" json:"github,omitempty"`
	TikTok       string `bson:"tiktok,omitempty" json:"tiktok,omitempty"`
	Pinterest    string `bson:"pinterest,omitempty" json:"pinterest,omitempty"`
	ShowInHeader bool   `bson:"show_in_header" json:"show_in_header"`
	ShowInFooter bool   `bson:"show_in_footer" json:"show_in_footer"`
}

// SEOSettings represents SEO configuration
type SEOSettings struct {
	SiteTitle       string                 `bson:"site_title,omitempty" json:"site_title,omitempty"`
	SiteDescription string                 `bson:"site_description,omitempty" json:"site_description,omitempty"`
	SiteKeywords    string                 `bson:"site_keywords,omitempty" json:"site_keywords,omitempty"`
	OGImage         string                 `bson:"og_image,omitempty" json:"og_image,omitempty"`
	TwitterCard     string                 `bson:"twitter_card,omitempty" json:"twitter_card,omitempty"`
	CanonicalURL    string                 `bson:"canonical_url,omitempty" json:"canonical_url,omitempty"`
	RobotsContent   string                 `bson:"robots_content,omitempty" json:"robots_content,omitempty"`
	MetaTags        map[string]string      `bson:"meta_tags,omitempty" json:"meta_tags,omitempty"`
	Schema          map[string]interface{} `bson:"schema,omitempty" json:"schema,omitempty"`
}

// AnalyticsSettings represents analytics integration
type AnalyticsSettings struct {
	GoogleAnalytics    string `bson:"google_analytics,omitempty" json:"google_analytics,omitempty"`
	GoogleTagManager   string `bson:"google_tag_manager,omitempty" json:"google_tag_manager,omitempty"`
	FacebookPixel      string `bson:"facebook_pixel,omitempty" json:"facebook_pixel,omitempty"`
	HotjarSiteID       string `bson:"hotjar_site_id,omitempty" json:"hotjar_site_id,omitempty"`
	MatomoSiteID       string `bson:"matomo_site_id,omitempty" json:"matomo_site_id,omitempty"`
	MatomoURL          string `bson:"matomo_url,omitempty" json:"matomo_url,omitempty"`
	CustomTrackingCode string `bson:"custom_tracking_code,omitempty" json:"custom_tracking_code,omitempty"`
}

// ThemeBackup represents theme backup information
type ThemeBackup struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ThemeName   string             `bson:"theme_name" json:"theme_name"`
	BackupName  string             `bson:"backup_name" json:"backup_name"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Data        ThemeCustomization `bson:"data" json:"data"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	CreatedBy   string             `bson:"created_by" json:"created_by"`
	Size        int64              `bson:"size" json:"size"`
	IsAutomatic bool               `bson:"is_automatic" json:"is_automatic"`
}

// ThemeConfig represents theme configuration options
type ThemeConfig struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ThemeName    string             `bson:"theme_name" json:"theme_name"`
	ConfigKey    string             `bson:"config_key" json:"config_key"`
	ConfigValue  interface{}        `bson:"config_value" json:"config_value"`
	ConfigType   string             `bson:"config_type" json:"config_type"` // string, number, boolean, object, array
	IsRequired   bool               `bson:"is_required" json:"is_required"`
	DefaultValue interface{}        `bson:"default_value,omitempty" json:"default_value,omitempty"`
	Description  string             `bson:"description,omitempty" json:"description,omitempty"`
	Category     string             `bson:"category,omitempty" json:"category,omitempty"`
	Order        int                `bson:"order" json:"order"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	UpdatedBy    string             `bson:"updated_by" json:"updated_by"`
}

// ThemeInstallRequest represents theme installation request
type ThemeInstallRequest struct {
	Name        string `json:"name" binding:"required"`
	Source      string `json:"source"` // file, url, marketplace
	URL         string `json:"url,omitempty"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
}

// ThemeUpdateRequest represents theme update request
type ThemeUpdateRequest struct {
	Version     string                 `json:"version,omitempty"`
	Description string                 `json:"description,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// ThemeCustomizationRequest represents theme customization update request
type ThemeCustomizationRequest struct {
	Colors      map[string]string      `json:"colors,omitempty"`
	Fonts       map[string]string      `json:"fonts,omitempty"`
	Layout      map[string]interface{} `json:"layout,omitempty"`
	Spacing     map[string]string      `json:"spacing,omitempty"`
	Typography  TypographySettings     `json:"typography,omitempty"`
	CustomCSS   string                 `json:"custom_css,omitempty"`
	CustomJS    string                 `json:"custom_js,omitempty"`
	HeaderCode  string                 `json:"header_code,omitempty"`
	FooterCode  string                 `json:"footer_code,omitempty"`
	Logo        LogoSettings           `json:"logo,omitempty"`
	Favicon     string                 `json:"favicon,omitempty"`
	SocialMedia SocialMediaSettings    `json:"social_media,omitempty"`
	SEO         SEOSettings            `json:"seo,omitempty"`
	Analytics   AnalyticsSettings      `json:"analytics,omitempty"`
}

// Helper methods for ThemeMetadata

// IsInstalled checks if theme is properly installed
func (t *ThemeMetadata) IsInstalled() bool {
	return t.Path != "" && t.InstalledAt.After(time.Time{})
}

// IsUpdateAvailable checks if theme has updates (placeholder logic)
func (t *ThemeMetadata) IsUpdateAvailable() bool {
	// In a real implementation, you would check against a marketplace or repository
	return false
}

// GetThemeDir returns the theme directory path
func (t *ThemeMetadata) GetThemeDir() string {
	return t.Path
}

// GetAssetPath returns the full path to a specific asset
func (t *ThemeMetadata) GetAssetPath(assetType, filename string) string {
	switch assetType {
	case "css":
		return filepath.Join(t.Path, "assets", "css", filename)
	case "js":
		return filepath.Join(t.Path, "assets", "js", filename)
	case "images":
		return filepath.Join(t.Path, "assets", "images", filename)
	case "fonts":
		return filepath.Join(t.Path, "assets", "fonts", filename)
	default:
		return filepath.Join(t.Path, "assets", filename)
	}
}

// Validate validates theme metadata
func (t *ThemeMetadata) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("theme name is required")
	}
	if t.Version == "" {
		return fmt.Errorf("theme version is required")
	}
	if t.Path == "" {
		return fmt.Errorf("theme path is required")
	}
	return nil
}

// Helper methods for ThemeCustomization

// GetColor returns a color value with fallback
func (tc *ThemeCustomization) GetColor(key, fallback string) string {
	if tc.Colors != nil {
		if color, exists := tc.Colors[key]; exists && color != "" {
			return color
		}
	}
	return fallback
}

// SetColor sets a color value
func (tc *ThemeCustomization) SetColor(key, value string) {
	if tc.Colors == nil {
		tc.Colors = make(map[string]string)
	}
	tc.Colors[key] = value
}

// GetFont returns a font value with fallback
func (tc *ThemeCustomization) GetFont(key, fallback string) string {
	if tc.Fonts != nil {
		if font, exists := tc.Fonts[key]; exists && font != "" {
			return font
		}
	}
	return fallback
}

// HasCustomCSS checks if custom CSS is defined
func (tc *ThemeCustomization) HasCustomCSS() bool {
	return tc.CustomCSS != ""
}

// HasCustomJS checks if custom JavaScript is defined
func (tc *ThemeCustomization) HasCustomJS() bool {
	return tc.CustomJS != ""
}

// GetAnalyticsCode returns formatted analytics code
func (tc *ThemeCustomization) GetAnalyticsCode() string {
	var code strings.Builder

	// Google Analytics
	if tc.Analytics.GoogleAnalytics != "" {
		code.WriteString(fmt.Sprintf(`
<!-- Google Analytics -->
<script async src="https://www.googletagmanager.com/gtag/js?id=%s"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());
  gtag('config', '%s');
</script>
`, tc.Analytics.GoogleAnalytics, tc.Analytics.GoogleAnalytics))
	}

	// Google Tag Manager
	if tc.Analytics.GoogleTagManager != "" {
		code.WriteString(fmt.Sprintf(`
<!-- Google Tag Manager -->
<script>(function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':
new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],
j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src=
'https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);
})(window,document,'script','dataLayer','%s');</script>
`, tc.Analytics.GoogleTagManager))
	}

	// Facebook Pixel
	if tc.Analytics.FacebookPixel != "" {
		code.WriteString(fmt.Sprintf(`
<!-- Facebook Pixel -->
<script>
  !function(f,b,e,v,n,t,s)
  {if(f.fbq)return;n=f.fbq=function(){n.callMethod?
  n.callMethod.apply(n,arguments):n.queue.push(arguments)};
  if(!f._fbq)f._fbq=n;n.push=n;n.loaded=!0;n.version='2.0';
  n.queue=[];t=b.createElement(e);t.async=!0;
  t.src=v;s=b.getElementsByTagName(e)[0];
  s.parentNode.insertBefore(t,s)}(window, document,'script',
  'https://connect.facebook.net/en_US/fbevents.js');
  fbq('init', '%s');
  fbq('track', 'PageView');
</script>
`, tc.Analytics.FacebookPixel))
	}

	// Custom tracking code
	if tc.Analytics.CustomTrackingCode != "" {
		code.WriteString(tc.Analytics.CustomTrackingCode)
	}

	return code.String()
}
