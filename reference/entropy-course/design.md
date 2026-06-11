# Design System: Crashcourse JavaScript

**Project ID:** Not provided in source HTML.  
**Source:** `design/course.html`

## 1. Visual Theme & Atmosphere

Crashcourse JavaScript is a warm, editorial learning interface: quiet, compact, and study-focused without feeling sterile. The design reads like a modern course reader with a fixed curriculum rail, an article-like lesson body, and tool surfaces for quizzes, coding practice, and cheatsheets.

The atmosphere is tactile and academic. A warm paper canvas, hairline dividers, restrained terracotta accents, serif display titles, and monospace labels create the feeling of a carefully annotated workbook. Density is moderate: navigation is compact and scannable, while the lesson content gets enough space to breathe.

The product should feel calm, durable, and low-friction. Avoid decorative spectacle. New screens should prioritize learning progress, readable hierarchy, and precise interaction states.

## 2. Color Palette & Roles

- **Warm Paper Canvas (#FBF9F4):** The primary application background. Use it for the page shell, sidebar, and open content areas.
- **Clean Elevated White (#FFFFFF):** Raised panels, aside cards, quiz options, editor surfaces, and cheatsheet cards.
- **Recessed Parchment (#F2EEE5):** Hover rows, muted control backgrounds, editor headers, inactive tab icons, and inset utility surfaces.
- **Code Parchment (#F5F1E8):** Inline code chips and code-adjacent surfaces that need a softer technical feel.
- **Linen Hairline (#E7E1D2):** Default borders and separators.
- **Aged Taupe Border (#CBC2AE):** Stronger borders, inactive dots, button strokes, and dashed helper outlines.
- **Near-Black Charcoal (#1B1814):** Primary text, dark course card, primary buttons, selected markers, and dark console panels.
- **Muted Coffee Text (#6B6357):** Secondary body copy, inactive labels, summaries, helper text, and neutral metadata.
- **Dusty Warm Gray (#A8A091):** Subtle metadata, section labels, inactive numbering, breadcrumbs, and quiet timestamps.
- **Terracotta Action (#B44E2E):** Primary accent for active tabs, progress bars, active lesson dots, callout rules, and important instructional emphasis.
- **Burnt Terracotta (#852B09):** Strong accent text for active lesson labels, key terms, and emphasized inline concepts.
- **Terracotta Wash (#FFEDE6):** Selected lesson backgrounds, callouts, active outlines, and gentle accent fills.
- **Fern Success (#33854A):** Completed states, correct quiz answers, passing tests, and positive status dots.
- **Mint Success Wash (#E0FAE4):** Correct-answer and success-feedback backgrounds.
- **Clay Error (#BD413F):** Wrong quiz answers, failing tests, and error markers.
- **Blush Error Wash (#FFE8E4):** Wrong-answer and danger-feedback backgrounds.
- **Ochre Warning (#C8942D):** Medium difficulty and caution states.
- **Video Gradient Creams (#FFF5E8, #FFE8D4, #F8D4B3):** Concept-video placeholders and warm visual emphasis areas.
- **Syntax Blue (#006AAA) and Syntax Teal (#00858D):** Function and property token colors inside code examples.

## 3. Typography Rules

Use **Instrument Serif** as the display face for identity, lesson titles, card titles, prompts, and editorial headings. It should feel refined and bookish, with large type set at a tight line-height. The main lesson title sits around 44px, course card titles around 30px, section headings and quiz prompts around 26px, and cheatsheet titles around 22px.

Use **Geist** as the body and interface face. Body text sits at 15px with a relaxed 1.55 line-height; lesson prose increases to 16px with about 1.65 line-height for comfortable reading. Medium-weight Geist is used sparingly for tab labels, chapter titles, buttons, and important UI names.

Use **JetBrains Mono** for curriculum numbers, section labels, breadcrumbs, code, timings, status metadata, editor text, and test output. Labels are small, usually 10.5px to 12.5px, uppercase, and tracked out. This monospace layer gives the interface a precise courseware rhythm.

Display text uses slight tightening for elegance. Body copy remains steady and practical. Code text should always be visibly distinct through both monospace type and a parchment-backed chip or block.

## 4. Component Stylings

- **App Shell:** Use a two-column desktop layout with a 296px sticky sidebar and a flexible main pane capped around 1100px. The sidebar uses the warm paper background, a single linen right border, and compact vertical spacing.
- **Brand Header:** Pair a small 22px charcoal square mark with serif product text. The mark has softly squared 5px corners and reversed paper-colored lettering.
- **Course Card:** Use a charcoal panel with 8px corners, paper-colored text, compact padding, a subtle warm radial highlight, and a 4px pill progress track. This is the strongest visual anchor in the sidebar.
- **Curriculum Rows:** Chapter rows are low-height, text-first controls with 6px corners. Expanded lessons sit under a thin vertical linen guide. Active lessons use terracotta wash, burnt terracotta text, and a small terracotta status dot.
- **Tabs:** Tabs sit on a hairline bottom rule and use an active terracotta underline. Each tab includes a 22px softly squared icon tile; active icon tiles fill with terracotta and reverse to white.
- **Cards and Containers:** Default cards use elevated white, a 1px linen border, 14px corners, and whisper-soft shadowing. Aside cards are slightly plainer, with 8px corners and no visible shadow.
- **Buttons:** Buttons are compact, 6px-corner controls with 8px by 14px padding, medium text, and taupe borders. Primary buttons use charcoal fill; accent buttons use terracotta fill; ghost buttons stay transparent until hover.
- **Concept Media:** The video frame is a 16:9 rounded rectangle with 14px corners, a cream-to-apricot gradient, soft radial warmth, and a centered 56px circular charcoal play button.
- **Callouts:** Instructional callouts use terracotta wash, a 3px terracotta left rule, and rounded right corners only. They should feel like margin notes, not alerts.
- **Quiz Options:** Answer choices are white cards with 8px corners, linen borders, circular markers, and clear state fills: charcoal for selected, fern/mint for correct, clay/blush for wrong.
- **Practice Editor:** The coding surface uses a white editor body, a recessed parchment header, 14px outer corners, monospace input, and a charcoal test console attached below.
- **Cheatsheet Cards:** Cheatsheets use a six-column grid on desktop. Cards are elevated white with 14px corners and enough internal padding for compact tables, snippets, diagrams, and gotchas.

Depth should stay restrained. Prefer hairline borders and very soft shadows over heavy elevation. Motion is brief and functional: hover color changes around 120ms to 150ms, tab content fades upward over about 250ms, and active progress width changes over about 400ms.

## 5. Layout Principles

The interface is built for repeated study sessions. Keep navigation persistent, lesson context visible, and task-specific panels close to the learning content.

Desktop pages use a stable left rail and generous main padding of roughly 28px top, 44px sides, and 80px bottom. Lesson headers lead with a small monospace breadcrumb, a large serif title, and a muted summary no wider than about 620px.

Concept views split into a large content column and a 280px sticky aside with a 32px gap. Practice views use two equal columns with an 18px gap. Cheatsheet views use a six-column grid with 18px gaps and card spans for hierarchy.

Responsive behavior should collapse complex grids before content becomes cramped: concept aside stacks below the content around 1100px, practice columns stack around 1000px, and cheatsheet cards become full-width around 900px.

Use whitespace as a quiet organizing tool. Vertical rhythm should be compact in navigation, moderate in cards, and more generous in prose. New screens should preserve the warm paper background, editorial serif hierarchy, monospace metadata, and terracotta-only action emphasis.
