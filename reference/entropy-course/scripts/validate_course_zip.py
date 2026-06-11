#!/usr/bin/env python3
"""Static validator for a v1 course import zip (or unpacked source dir).

Replicates the rules in docs/import-format.md so a package can be checked
*before* running `course-cli import plan/apply` against a database. It does not
replace the Go importer's domain validation, but it catches every layout, schema,
reference, and enum error documented in the import contract.

Usage:
    python3 scripts/validate_course_zip.py courses/javascript-foundations.zip
    python3 scripts/validate_course_zip.py courses/javascript-foundations        # source dir
"""
import sys, os, re, io, zipfile, posixpath
try:
    import yaml
except ImportError:
    sys.exit("PyYAML required: pip install pyyaml --break-system-packages")

SLUG_RE   = re.compile(r"^[a-z0-9]+(-[a-z0-9]+)*$")
YT_RE     = re.compile(r"^[A-Za-z0-9_-]{11}$")
URL_RE    = re.compile(r"^https?://", re.I)
PROVIDERS = {"url", "youtube", "mux"}
LANGS     = {"javascript", "typescript", "golang", "rust"}
BLOCK_KINDS = {"text", "video", "quiz", "practice"}

errors, warnings = [], []
def err(m):  errors.append(m)
def warn(m): warnings.append(m)

# ---- allowed key sets (unknown keys are parse errors in the importer) ----
K_COURSE   = {"title", "slug", "description", "status"}
K_LESSON   = {"title", "order", "blocks"}
K_BLOCK    = {"kind", "position", "markdown", "video_provider", "video_locator",
              "video_caption", "quiz_ref", "practice_ref"}
K_QUIZ     = {"slug", "title", "pass_threshold", "questions"}
K_QUESTION = {"type", "position", "prompt", "options", "correct_indices", "explanation"}
K_PRACTICE = {"slug", "title", "language", "prompt", "starter_code", "solution", "test_cases"}
K_PTEST    = {"position", "name", "stdin", "expected_stdout"}
K_TEST     = {"slug", "title", "time_limit_minutes", "pass_threshold", "solution", "items"}
K_TSOL     = {"zip_provider", "zip_locator", "video_provider", "video_locator", "video_caption"}
K_ITEM     = {"kind", "position", "prompt", "choice_type", "options", "correct_indices",
              "explanation", "coding_prompt", "language", "starter_code", "solution", "test_cases"}

def check_keys(where, d, allowed):
    if not isinstance(d, dict):
        err(f"{where}: expected a mapping"); return
    extra = set(d) - allowed
    if extra: err(f"{where}: unknown key(s) {sorted(extra)} (importer rejects unknown YAML keys)")

# ---- load files (from zip or dir) into {normpath: bytes} ----
def load(target):
    files = {}
    if os.path.isdir(target):
        for root, _, fns in os.walk(target):
            for fn in fns:
                full = os.path.join(root, fn)
                rel = posixpath.normpath(os.path.relpath(full, target).replace(os.sep, "/"))
                files[rel] = open(full, "rb").read()
    else:
        with zipfile.ZipFile(target) as z:
            for info in z.infolist():
                if info.is_dir(): continue
                raw = info.filename.replace("\\", "/")
                if raw.startswith("/") or ".." in raw.split("/"):
                    err(f"unsafe file path: {info.filename}"); continue
                norm = posixpath.normpath(raw)
                if norm in files: err(f"duplicate normalized file path: {norm}")
                files[norm] = z.read(info)
    return files

def parse_yaml(path, data):
    try:
        return yaml.safe_load(data.decode("utf-8"))
    except Exception as e:
        err(f"{path}: malformed YAML ({e})"); return None

def split_frontmatter(path, text):
    if not text.startswith("---"):
        err(f"{path}: missing lesson frontmatter"); return None, ""
    m = re.match(r"^---\s*\n(.*?)\n---\s*(?:\n(.*))?$", text, re.S)
    if not m:
        err(f"{path}: unterminated lesson frontmatter"); return None, ""
    fm = parse_yaml(path, m.group(1).encode())
    return fm, (m.group(2) or "")

def num_in_unit(where, v):
    if v is None: return
    if not isinstance(v, (int, float)) or isinstance(v, bool) or not (0 <= v <= 1):
        err(f"{where}: pass_threshold must be a number in [0,1], got {v!r}")

def check_choice(where, qtype, options, idx, require_correct=True):
    if not isinstance(options, list) or len(options) < 1:
        err(f"{where}: needs at least one option"); options = options or []
    if len(options) < 2:
        warn(f"{where}: only {len(options)} option(s); domain validation usually wants >=2")
    if idx is None:
        if require_correct: err(f"{where}: correct_indices required for a valid apply")
        return
    if not isinstance(idx, list) or not idx:
        err(f"{where}: correct_indices must be a non-empty list"); return
    if any((not isinstance(i, int)) or i < 0 or i >= len(options) for i in idx):
        err(f"{where}: correct_indices {idx} out of range for {len(options)} options")
    if len(set(idx)) != len(idx):
        err(f"{where}: correct_indices {idx} contains duplicates")
    if qtype == "single" and len(idx) != 1:
        err(f"{where}: a 'single' choice must have exactly one correct index, got {idx}")

# ============================================================ run
def main():
    if len(sys.argv) != 2:
        sys.exit("usage: validate_course_zip.py <course.zip | source-dir>")
    target = sys.argv[1]
    if not os.path.exists(target): sys.exit(f"not found: {target}")
    files = load(target)

    # ---- layout / allowlist ----
    if "format_version.txt" not in files: err("missing required file: format_version.txt")
    if "course.yaml" not in files:        err("missing required file: course.yaml")
    for p in files:
        top = p.split("/")[0]
        if p in ("format_version.txt", "course.yaml"): continue
        if p.endswith(".yml"): err(f"{p}: YAML files must use .yaml, not .yml")
        if top == "lessons":     ok = p.endswith(".md")
        elif top in ("quizzes", "practices", "tests"): ok = p.endswith(".yaml")
        else: ok = False
        if not ok: err(f"unexpected file path: {p}")

    # ---- format_version ----
    if "format_version.txt" in files:
        fv = files["format_version.txt"].decode("utf-8").strip()
        if fv == "": err("format_version.txt: empty content")
        elif fv != "1": err(f"unsupported import format: {fv!r} (expected '1')")

    # ---- course.yaml ----
    if "course.yaml" in files:
        c = parse_yaml("course.yaml", files["course.yaml"]) or {}
        check_keys("course.yaml", c, K_COURSE)
        if not str(c.get("title", "")).strip(): err("course.yaml: title is required")
        slug = c.get("slug", "")
        if not SLUG_RE.match(str(slug)): err(f"course.yaml: invalid slug {slug!r} (lowercase letters, numbers, single hyphens)")
        if c.get("status") not in ("draft", "published"): err(f"course.yaml: status must be draft|published, got {c.get('status')!r}")

    # ---- collect quiz/practice slugs for ref integrity ----
    quiz_slugs, practice_slugs, test_slugs = set(), set(), set()

    def each(prefix):
        return sorted(p for p in files if p.startswith(prefix))

    # ---- quizzes ----
    for p in each("quizzes/"):
        q = parse_yaml(p, files[p]);
        if q is None: continue
        check_keys(p, q, K_QUIZ)
        s = q.get("slug")
        if not s: err(f"{p}: quiz slug required")
        elif s in quiz_slugs: err(f"{p}: duplicate quiz slug {s!r}")
        else: quiz_slugs.add(s)
        if not str(q.get("title", "")).strip(): err(f"{p}: quiz title required")
        num_in_unit(p, q.get("pass_threshold"))
        for i, qu in enumerate(q.get("questions") or []):
            w = f"{p} q[{i}]"
            check_keys(w, qu, K_QUESTION)
            if qu.get("type") not in ("single", "multiple"): err(f"{w}: type must be single|multiple")
            if not str(qu.get("prompt", "")).strip(): err(f"{w}: prompt required")
            check_choice(w, qu.get("type"), qu.get("options"), qu.get("correct_indices"))

    # ---- practices ----
    for p in each("practices/"):
        pr = parse_yaml(p, files[p])
        if pr is None: continue
        check_keys(p, pr, K_PRACTICE)
        s = pr.get("slug")
        if not s: err(f"{p}: practice slug required")
        elif s in practice_slugs: err(f"{p}: duplicate practice slug {s!r}")
        else: practice_slugs.add(s)
        if not str(pr.get("title", "")).strip(): err(f"{p}: practice title required")
        if pr.get("language") not in LANGS: err(f"{p}: language must be one of {sorted(LANGS)}, got {pr.get('language')!r}")
        if not str(pr.get("prompt", "")).strip(): err(f"{p}: prompt required")
        for j, tc in enumerate(pr.get("test_cases") or []):
            w = f"{p} test_case[{j}]"
            check_keys(w, tc, K_PTEST)
            if not str(tc.get("expected_stdout", "")): err(f"{w}: expected_stdout must not be empty")

    # ---- tests ----
    for p in each("tests/"):
        t = parse_yaml(p, files[p])
        if t is None: continue
        check_keys(p, t, K_TEST)
        s = t.get("slug")
        if not s: err(f"{p}: test slug required")
        elif s in test_slugs: err(f"{p}: duplicate test slug {s!r}")
        else: test_slugs.add(s)
        if not str(t.get("title", "")).strip(): err(f"{p}: test title required")
        tl = t.get("time_limit_minutes")
        if tl is not None and (not isinstance(tl, int) or isinstance(tl, bool) or tl <= 0):
            err(f"{p}: time_limit_minutes must be a positive integer, got {tl!r}")
        num_in_unit(p, t.get("pass_threshold"))
        sol = t.get("solution")
        if sol is not None:
            check_keys(f"{p} solution", sol, K_TSOL)
            if sol.get("zip_provider") not in PROVIDERS: err(f"{p} solution: zip_provider must be {sorted(PROVIDERS)}")
            if not sol.get("zip_locator"): err(f"{p} solution: zip_locator required when solution present")
            if sol.get("video_provider") not in PROVIDERS: err(f"{p} solution: video_provider must be {sorted(PROVIDERS)}")
            if not sol.get("video_locator"): err(f"{p} solution: video_locator required when solution present")
        for i, it in enumerate(t.get("items") or []):
            w = f"{p} item[{i}]"
            check_keys(w, it, K_ITEM)
            if it.get("kind") == "choice":
                if not str(it.get("prompt", "")).strip(): err(f"{w}: choice prompt required")
                if it.get("choice_type") not in ("single", "multiple"): err(f"{w}: choice_type must be single|multiple")
                check_choice(w, it.get("choice_type"), it.get("options"), it.get("correct_indices"))
            elif it.get("kind") == "coding":
                if not str(it.get("coding_prompt", "")).strip(): err(f"{w}: coding_prompt required")
                if it.get("language") not in LANGS: err(f"{w}: language must be {sorted(LANGS)}, got {it.get('language')!r}")
                for j, tc in enumerate(it.get("test_cases") or []):
                    ww = f"{w} test_case[{j}]"
                    check_keys(ww, tc, K_PTEST)
                    if not str(tc.get("expected_stdout", "")): err(f"{ww}: expected_stdout must not be empty")
            else:
                err(f"{w}: unknown test item kind {it.get('kind')!r} (choice|coding)")

    # ---- lessons (parsed last so quiz/practice slugs are known) ----
    lesson_files = each("lessons/")
    if not lesson_files: warn("no lessons/*.md found")
    for p in lesson_files:
        fm, body = split_frontmatter(p, files[p].decode("utf-8"))
        if fm is None: continue
        check_keys(p, fm, K_LESSON)
        if not str(fm.get("title", "")).strip(): err(f"{p}: lesson title required")
        blocks = fm.get("blocks")
        if not isinstance(blocks, list) or len(blocks) < 1:
            err(f"{p}: blocks must contain at least one block"); blocks = []
        body_used = False
        for i, b in enumerate(blocks):
            w = f"{p} block[{i}]"
            check_keys(w, b, K_BLOCK)
            kind = b.get("kind")
            if kind not in BLOCK_KINDS:
                err(f"{w}: unknown block kind {kind!r} ({sorted(BLOCK_KINDS)})"); continue
            if kind == "text":
                if not str(b.get("markdown", "")).strip():
                    if body.strip() and not body_used: body_used = True
                    else: err(f"{w}: text block needs markdown (or a non-empty lesson body)")
            elif kind == "video":
                vp = b.get("video_provider")
                if vp not in PROVIDERS: err(f"{w}: video_provider must be {sorted(PROVIDERS)}, got {vp!r}")
                loc = b.get("video_locator")
                if not loc: err(f"{w}: video_locator required")
                elif vp == "youtube" and not YT_RE.match(str(loc)): err(f"{w}: youtube locator {loc!r} should be an 11-char video id")
                elif vp == "url" and not URL_RE.match(str(loc)): err(f"{w}: url locator must be a http(s) URL")
            elif kind == "quiz":
                ref = b.get("quiz_ref")
                if not ref: err(f"{w}: quiz block requires quiz_ref")
                elif ref not in quiz_slugs: err(f"{w}: quiz_ref {ref!r} does not match any quizzes/*.yaml slug")
            elif kind == "practice":
                ref = b.get("practice_ref")
                if not ref: err(f"{w}: practice block requires practice_ref")
                elif ref not in practice_slugs: err(f"{w}: practice_ref {ref!r} does not match any practices/*.yaml slug")

    # ---- report ----
    print(f"Validated: {target}")
    print(f"  files: {len(files)} | quizzes: {len(quiz_slugs)} | practices: {len(practice_slugs)} | tests: {len(test_slugs)} | lessons: {len(lesson_files)}")
    for wmsg in warnings: print("  WARN:", wmsg)
    if errors:
        print(f"\nFAIL — {len(errors)} error(s):")
        for e in errors: print("  -", e)
        sys.exit(1)
    print("\nPASS — package conforms to the v1 import contract.")

if __name__ == "__main__":
    main()
