package settings

// AboutHero is the about page hero section.
type AboutHero struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	ImageURL string `json:"image_url,omitempty"`
}

// AboutStory is the company story section.
type AboutStory struct {
	Title           string `json:"title"`
	ContentHTML     string `json:"content_html,omitempty"`
	ContentMarkdown string `json:"content_markdown,omitempty"`
}

// AboutTextBlock is a titled text section such as mission or vision.
type AboutTextBlock struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// AboutMilestone is a timeline entry on the about page.
type AboutMilestone struct {
	Year        string `json:"year"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// AboutTeamMember is a team member card on the about page.
type AboutTeamMember struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	PhotoURL string `json:"photo_url,omitempty"`
}

// AboutStats holds about page statistics.
type AboutStats struct {
	YearsExperience   int   `json:"years_experience"`
	HappyCustomers    int64 `json:"happy_customers"`
	ProductsCount     int64 `json:"products_count"`
	CompletedProjects int   `json:"completed_projects"`
}

// About holds CMS content for the about page.
type About struct {
	Hero       AboutHero         `json:"hero"`
	Story      AboutStory        `json:"story"`
	Mission    AboutTextBlock    `json:"mission"`
	Vision     AboutTextBlock    `json:"vision"`
	Milestones []AboutMilestone  `json:"milestones"`
	Team       []AboutTeamMember `json:"team"`
	Stats      AboutStats        `json:"stats"`
}

// PublicContact extends contact info exposed on the about page.
type PublicContact struct {
	Phone        string `json:"phone,omitempty"`
	Mobile       string `json:"mobile,omitempty"`
	Email        string `json:"email,omitempty"`
	Address      string `json:"address,omitempty"`
	WorkingHours string `json:"working_hours,omitempty"`
}

// PublicSocial extends social links exposed on the about page.
type PublicSocial struct {
	Instagram string `json:"instagram,omitempty"`
	WhatsApp  string `json:"whatsapp,omitempty"`
	Telegram  string `json:"telegram,omitempty"`
	Facebook  string `json:"facebook,omitempty"`
	Twitter   string `json:"twitter,omitempty"`
	LinkedIn  string `json:"linkedin,omitempty"`
	YouTube   string `json:"youtube,omitempty"`
	TikTok    string `json:"tiktok,omitempty"`
}

// AboutSEO is SEO metadata for the about page.
type AboutSEO struct {
	MetaTitle       string `json:"meta_title,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
}
