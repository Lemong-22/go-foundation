# Design System: [Project/Product Name]

**Project ID:** [Identifier]  
**Source:** `[design file / mockup path]`

> **Template origin:** Di-extract dari `vendor/entropy-course/design.md` (Stephen Antoni). Pattern: 5-section design doc yang fokus ke **aesthetic decisions + their rationale**, bukan pixel-level mockup.

---

## 1. Visual Theme & Atmosphere

[2-3 paragraf. JANGAN list fitur di sini. List di sini = ambiance, mood, feel, archetype. PERTANYAAN UNTUK DIJAWAB: "Kalau produk ini ruangan, gimana rasanya masuk pertama kali?"]

**Pertanyaan panduan:**
- Archetype: editorial / dashboard / gamified / clinical / playful?
- Density: compact (data-dense) atau spacious (breathing room)?
- Mood: warm (cream, paper) atau cool (blue, white)?
- Editorial feel: book-like / workbook / manual / guide?

**Template (extract dari entropy-course):**
> [Product] is a warm, editorial learning interface: quiet, compact, and study-focused without feeling sterile. The design reads like a modern [X] with [persistent element], [main content], and [secondary surfaces].
>
> The atmosphere is [tactile, academic, technical, etc.]. A [background color], [dividers], [accents], [display face], and [monospace label] create the feeling of a [X]. Density is [moderate / dense / light]: navigation is [compact / spacious], while [content type] gets enough space to breathe.
>
> The product should feel [adjective 1], [adjective 2], and [adjective 3]. Avoid decorative spectacle. New screens should prioritize [core value: learning progress, readable hierarchy, precise interaction states].

---

## 2. Color Palette & Roles

[List SEMUA warna dengan role spesifik-nya. JANGAN list "primary blue" — list "Primary action button background in active state".]

**Template:**
- **[Color Name] (#HEX):** [Role]. Use it for [where].
- **[Color Name] (#HEX):** [Role]. Use it for [where].

**Extract dari entropy-course untuk referensi role taxonomy:**
- Backgrounds: canvas, elevated surface, recessed/hover
- Borders: hairline, strong, dashed helper
- Text: primary, secondary, muted
- Action: primary accent, strong accent text, accent wash
- State: success, success-wash, error, error-wash, warning
- Code/syntax: function token, property token

---

## 3. Typography Rules

[Display face / body face / monospace. Size + line-height untuk setiap konteks. Kapan pakai weight medium vs regular.]

**Template:**
- Use **[Display Face]** as the display face for [where]. It should feel [adjective]. Display sizes around [N]px with line-height [N].
- Use **[Body Face]** as body and interface. Body sits at [N]px with line-height [N]. [Context] increases to [N]px with line-height [N].
- Use **[Monospace Face]** for [where]. Labels are small, usually [N]px to [N]px, uppercase, tracked out. This monospace layer gives [impression].

**Principle:** Display text uses slight tightening. Body remains steady. Code text visibly distinct through both monospace + parchment-backed chip.

---

## 4. Component Stylings

[Untuk setiap component penting, definisikan: dimensions, corners, padding, borders, hover/active state.]

**Template structure per component:**
- **[Component Name]:** [Shape + dimensions + radius]. [Padding]. [Border style]. [Background]. [Text style]. [Hover state]. [Active/selected state]. [Disabled state].

**Component checklist (extract dari entropy-course):**
- App Shell (sidebar width, main column cap)
- Brand Header (logo + name)
- Card variants (default, aside)
- Navigation (row height, icon tile, active state)
- Tabs (rule, underline, icon)
- Buttons (corner radius, padding, variants)
- Forms (input border, focus state, validation)
- Callouts (left rule, color, when to use)
- Code blocks (background, header, syntax colors)
- Modals / dialogs (overlay, card, corner radius)
- Empty states (illustration? text? action?)
- Loading states (skeleton vs spinner)

---

## 5. Layout Principles

[JANGAN list pixel di sini — itu masuk §4. Di sini: GRAMMAR. Pola-pola yang konsisten.]

**Template:**
- The interface is built for [use case: repeated sessions, quick checks, etc.]. Keep [X] persistent, [Y] visible, and [Z] close to [W].
- Desktop pages use [layout pattern: two-column / three-column / etc] with [sidebar width], [main cap], [padding sizes].
- [View type] views split into [columns] with [gap].
- Responsive behavior collapses [grid pattern] before content becomes cramped: [breakpoint A] → stack, [breakpoint B] → single column.
- Use whitespace as a quiet organizing tool. Vertical rhythm should be [compact / moderate / generous] in [navigation / cards / prose].
- New screens should preserve [core tokens: background, hierarchy, metadata, accent].

---

## Catatan Pakai Template Ini

1. **§1 Theme** — mood board. Kalau lo ga bisa jelasin "gimana rasanya" dalam 3 kalimat, lo belum cukup research. Lihat Dribbble, Behance, Mobbin, Refero untuk referensi sebelum nulis.
2. **§2 Color palette** — JANGAN list 20 warna random. Setiap warna harus punya ROLE. Kalau 2 warna punya role sama, pilih satu.
3. **§3 Typography** — 3 face max: display, body, monospace. Lebih dari itu = inkonsisten. Tentukan size scale (16/14/12/10 atau 18/16/14/12) SEBELUM implementasi.
4. **§4 Component** — untuk component baru, list SEBELUM coding. Ini "design system" bukan "design wishlist". Kalau tidak akan dipakai, jangan list.
5. **§5 Layout** — patterns, bukan pixels. Yang berulang-ulang. Breakpoint juga masuk sini.
6. **Iterate** — v1 design doc akan salah. Tulis, build, observe, update. Jangan tunggu "perfect" baru tulis.
7. **Reference real systems** — copy structure design doc dari project yang lo admire (Stripe, Linear, Vercel, dll). Template ini extract dari entropy-course yang bagus.
