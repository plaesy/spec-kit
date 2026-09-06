# Markdown Document Creation Instructions

## Purpose
Consistent, well-formatted, accessible markdown documents per Markdown Guide standards.

## Core Principles

### 1. Structure First
- Clear document outline; hierarchical headings (H1 -> H6); logical content flow
- Table of contents for long documents

### 2. Consistency Standards
- Consistent formatting patterns, spacing, indentation
- Consistent link/reference styles and naming conventions

### 3. Accessibility Focus
- Descriptive alt text for images; semantic heading structure
- Sufficient color contrast; meaningful link text

## Basic Syntax Requirements

### Headings
```markdown
# Main Title (H1) - Use only once per document
## Major Section (H2)
### Subsection (H3)
#### Sub-subsection (H4)
##### Minor heading (H5)
###### Smallest heading (H6)
```

**Rules:**
- Space after `#`; only one H1 per document
- Don't skip heading levels; keep headings concise/descriptive

### Paragraphs and Line Breaks
```markdown
This is a paragraph. Separate paragraphs with blank lines.

This is another paragraph. For line breaks within
paragraphs, use two spaces at the end of the line.
```

**Rules:**
- Blank lines between paragraphs, no indenting
- Two spaces + enter for line breaks

### Emphasis
```markdown
*Italic text* or _italic text_
**Bold text** or __bold text__
***Bold and italic*** or ___bold and italic___
```

**Rules:**
- Asterisks (*) preferred for compatibility
- Use emphasis sparingly - don't overuse bold/italic

### Lists
```markdown
Unordered lists:
- First item
- Second item
  - Nested item
  - Another nested item
- Third item

Ordered lists:
1. First step
2. Second step
   1. Sub-step
   2. Another sub-step
3. Third step
```

**Rules:**
- Consistent delimiters (`-` preferred for unordered)
- 2-4 space indent for nested items; proper numbering for ordered lists

### Links and References
```markdown
[Link text](https://example.com "Optional title")

Reference-style links:
[Link text][reference-id]

[reference-id]: https://example.com "Optional title"
```

**Rules:**
- Descriptive link text (never "click here")
- Optional titles for context; reference-style for repeated links

### Images
```markdown
![Descriptive alt text](image-url.jpg "Optional caption")

Linked images:
[![Alt text](image-url.jpg)](link-url.com)
```

**Rules:**
- Always meaningful alt text
- Relative paths for local images; reasonable file sizes

### Code
```markdown
Inline `code snippets` use backticks.

```language
// Fenced code blocks with syntax highlighting
function example() {
    return "Hello, World!";
}
```
```

**Rules:**
- Language specifiers for syntax highlighting
- Relevant, concise code examples; test snippets for accuracy

## Extended Syntax Features

### Tables
```markdown
| Header 1 | Header 2 | Header 3 |
|----------|:--------:|---------:|
| Left     | Center   | Right    |
| Aligned  | Aligned  | Aligned  |
```

**Rules:**
- Colons for alignment (`:---`, `:---:`, `---:`)
- Keep tables simple/readable; consider alternatives for complex data

### Task Lists
```markdown
- [x] Completed task
- [ ] Incomplete task
- [ ] Another incomplete task
```

### Blockquotes
```markdown
> This is a blockquote.
>
> It can span multiple paragraphs.
>
> > Nested blockquotes are also possible.
```

### Footnotes
```markdown
Here's a sentence with a footnote[^1].

[^1]: This is the footnote content.
```

## Document Structure Template

```markdown
# Document Title

Brief description of the document's purpose and scope.

## Table of Contents
- [Section 1](#section-1)
- [Section 2](#section-2)
- [Conclusion](#conclusion)

## Section 1

Content goes here...

### Subsection 1.1

More detailed content...

## Section 2

Content continues...

## Conclusion

Summary and next steps...

## References

- [Reference 1](link)
- [Reference 2](link)

---

*Last updated: [Date]*
*Author: [Name]*
```

## Quality Checklist

### Before Publishing
- [ ] Document has clear title and purpose
- [ ] Heading hierarchy is logical and complete
- [ ] All links are functional and relevant
- [ ] Images have descriptive alt text
- [ ] Code examples are tested and accurate
- [ ] Spelling and grammar are correct
- [ ] Document renders correctly in target environment

### Content Standards
- [ ] Information is accurate and up-to-date
- [ ] Language is clear and concise
- [ ] Examples are relevant and helpful
- [ ] Document serves its intended audience
- [ ] All claims are supported or referenced

### Technical Standards
- [ ] Markdown syntax is valid
- [ ] No broken internal/external links
- [ ] Images load correctly
- [ ] Code blocks have proper syntax highlighting
- [ ] Tables are properly formatted
- [ ] Special characters are properly escaped

## Common Mistakes to Avoid
- **Inconsistent formatting**: mixed emphasis styles, irregular spacing/indentation, inconsistent lists
- **Poor structure**: skipped heading levels, missing ToC on long docs, unclear section organization
- **Accessibility issues**: missing/poor alt text, non-descriptive link text, poor heading hierarchy
- **Technical errors**: broken links/images, invalid markdown syntax, untested code examples

## Best Practices Summary
1. Plan before writing - outline structure and key points
2. Write for your audience - appropriate language/detail level
3. Keep it simple - avoid over-formatting
4. Test everything - links, code, rendering
5. Review and revise - clarity, accuracy, completeness
6. Stay updated - keep content current

## Tools and Resources

- [Markdown Guide](https://www.markdownguide.org) - Comprehensive reference
- [CommonMark Spec](https://commonmark.org) - Standard specification
- Markdown editors: Typora, Mark Text, VS Code with extensions
- Online validators: Markdownlint, Markdown validation tools

---

*Follow these instructions to create professional, accessible, and maintainable markdown documents that serve their intended purpose effectively.*