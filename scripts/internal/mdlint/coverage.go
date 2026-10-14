package mdlint

// allRules is the implemented rule set. It is a deliberate subset of
// markdownlint's ~50 rules, chosen to cover the violations this repository
// actually has (measured 2026-09-25) plus the structural and accessibility rules
// that keep new documents honest. `plaesy validate markdown --list-rules` prints
// this set, and coverage.go records what is deliberately absent so nobody reads
// a clean run as parity with markdownlint.
var allRules = []ruleSpec{
	{id: "MD001", aliases: []string{"heading-increment"}, options: func() any { return &MD001Options{} }, check: checkMD001, fixable: false},
	{id: "MD003", aliases: []string{"heading-style"}, options: func() any { return &MD003Options{} }, check: checkMD003, fixable: false},
	{id: "MD004", aliases: []string{"ul-style"}, options: func() any { return &MD004Options{} }, check: checkMD004, fixable: false},
	{id: "MD007", aliases: []string{"ul-indent"}, options: func() any { return &MD007Options{} }, check: checkMD007, fixable: false},
	{id: "MD009", aliases: []string{"no-trailing-spaces"}, options: func() any { return &MD009Options{} }, check: checkMD009, fixable: true},
	{id: "MD010", aliases: []string{"no-hard-tabs"}, options: func() any { return &MD010Options{} }, check: checkMD010, fixable: true},
	{id: "MD012", aliases: []string{"no-multiple-blanks"}, options: func() any { return &MD012Options{} }, check: checkMD012, fixable: true},
	{id: "MD013", aliases: []string{"line-length"}, options: func() any { return &MD013Options{} }, check: checkMD013, fixable: false},
	{id: "MD022", aliases: []string{"blanks-around-headings"}, options: func() any { return &MD022Options{} }, check: checkMD022, fixable: true},
	{id: "MD024", aliases: []string{"no-duplicate-heading"}, options: func() any { return &MD024Options{} }, check: checkMD024, fixable: false},
	{id: "MD025", aliases: []string{"single-title", "single-h1"}, options: func() any { return &MD025Options{} }, check: checkMD025, fixable: false},
	{id: "MD026", aliases: []string{"no-trailing-punctuation"}, options: func() any { return &MD026Options{} }, check: checkMD026, fixable: false},
	{id: "MD029", aliases: []string{"ol-prefix"}, options: func() any { return &MD029Options{} }, check: checkMD029, fixable: false},
	{id: "MD031", aliases: []string{"blanks-around-fences"}, options: func() any { return &MD031Options{} }, check: checkMD031, fixable: true},
	{id: "MD032", aliases: []string{"blanks-around-lists"}, options: func() any { return &MD032Options{} }, check: checkMD032, fixable: true},
	{id: "MD036", aliases: []string{"no-emphasis-as-heading"}, options: func() any { return &MD036Options{} }, check: checkMD036, fixable: false},
	{id: "MD040", aliases: []string{"fenced-code-language"}, options: func() any { return &MD040Options{} }, check: checkMD040, fixable: false},
	{id: "MD041", aliases: []string{"first-line-heading"}, options: func() any { return &MD041Options{} }, check: checkMD041, fixable: false},
	{id: "MD046", aliases: []string{"code-block-style"}, options: func() any { return &MD046Options{} }, check: checkMD046, fixable: false},
	{id: "MD047", aliases: []string{"single-trailing-newline"}, options: func() any { return &MD047Options{} }, check: checkMD047, fixable: true},
	{id: "MD048", aliases: []string{"code-fence-style"}, options: func() any { return &MD048Options{} }, check: checkMD048, fixable: false},
	{id: "MD058", aliases: []string{"blanks-around-tables"}, options: func() any { return &MD058Options{} }, check: checkMD058, fixable: true},
	{id: GHAEmptyAltText, aliases: []string{"no-empty-alt-text"}, check: checkGHA001, fixable: false},
	{id: GHADefaultAltText, aliases: []string{"no-default-alt-text"}, check: checkGHA002, fixable: false},
	{id: GHAGenericLinkText, aliases: []string{"no-generic-link-text"}, check: checkGHA003, fixable: false},
}

// MD025Options is declared in rules_text.go with the other option structs; MD041
// reuses the same front_matter_title setting, so it gets a distinct type to keep
// its own decode target.
type MD041Options struct {
	FrontMatterTitle string `json:"front_matter_title"`
}

// RuleIDs returns the IDs of every implemented rule.
func RuleIDs() []string { return implementedRuleIDs() }

// RuleDescription returns a short description of an implemented rule, for
// `plaesy validate markdown --list-rules`.
func RuleDescription(id string) string {
	for i := range allRules {
		if allRules[i].id == id {
			return ruleDescriptions[allRules[i].id]
		}
	}
	return ""
}

// ruleDescriptions documents each rule for the listing. Keeping the text here
// means --list-rules never has to guess or omit a rule.
var ruleDescriptions = map[string]string{
	"MD001":  "Heading levels increment by one at a time",
	"MD003":  "Heading style is consistent (atx or setext)",
	"MD004":  "Unordered list marker style is consistent",
	"MD007":  "Unordered list indentation is a multiple of the configured width",
	"MD009":  "No trailing spaces (br_spaces tolerated)",
	"MD010":  "No hard tabs",
	"MD012":  "No more than one consecutive blank line",
	"MD013":  "Line length limit",
	"MD022":  "Headings are surrounded by blank lines",
	"MD024":  "No duplicate heading at the same level (siblings_only)",
	"MD025":  "Exactly one top-level heading",
	"MD026":  "No trailing punctuation in a heading",
	"MD029":  "Ordered list prefix style",
	"MD031":  "Fenced code blocks are surrounded by blank lines",
	"MD032":  "Lists are surrounded by blank lines",
	"MD036":  "Emphasis is not used instead of a heading",
	"MD040":  "Fenced code blocks declare a language",
	"MD041":  "First content line is a top-level heading",
	"MD046":  "Code block style is fenced",
	"MD047":  "File ends with exactly one newline",
	"MD048":  "Code fence style (backtick or tilde)",
	"MD058":  "Tables are surrounded by blank lines",
	"GHA001": "Images have alt text (GitHub ruleset)",
	"GHA002": "Alt text is not just the image file name (GitHub ruleset)",
	"GHA003": "Link text describes the destination (GitHub ruleset)",
}
