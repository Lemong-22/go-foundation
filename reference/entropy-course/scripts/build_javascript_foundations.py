#!/usr/bin/env python3
"""Build the 'JavaScript Foundations' v1 course package from structured data.
Emits a readable source tree + a course.zip ready for `course-cli import`."""
import os, shutil, zipfile, yaml

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SRC  = os.path.join(REPO, "courses/javascript-foundations")
ZIP  = os.path.join(REPO, "courses/javascript-foundations.zip")

# --- literal block style for multi-line strings (human-friendly YAML) ---
class _Lit(str): pass
def _lit_rep(dumper, data):
    return dumper.represent_scalar('tag:yaml.org,2002:str', data, style='|')
yaml.add_representer(_Lit, _lit_rep, Dumper=yaml.SafeDumper)
def wrap(o):
    if isinstance(o, dict):  return {k: wrap(v) for k, v in o.items()}
    if isinstance(o, list):  return [wrap(v) for v in o]
    if isinstance(o, str) and "\n" in o: return _Lit(o)
    return o
def dump(o):
    return yaml.safe_dump(wrap(o), sort_keys=False, allow_unicode=True, width=4096)

# ------------------------------------------------------------------ content
COURSE = {
    "title": "JavaScript Foundations",
    "slug": "javascript-foundations",
    "description": "A short, hands-on introduction to JavaScript for complete beginners: how to run code, variables and data types, and functions with control flow.",
    "status": "published",
}

LESSONS = [
    {
        "file": "01-getting-started.md",
        "front": {
            "title": "Getting Started with JavaScript",
            "order": 0,
            "blocks": [
                {"kind": "text", "markdown":
"""## What is JavaScript?

JavaScript (JS) is the programming language of the web. It runs in two main places:

- **In the browser**, to make web pages interactive.
- **On your computer or a server**, through a runtime called **Node.js**.

You write plain text instructions, and the JavaScript engine runs them top to bottom.

In this short course you'll learn just enough to read and write small programs with confidence: running code, storing values in variables, and making decisions with functions and control flow.
"""},
                {"kind": "video", "video_provider": "youtube",
                 "video_locator": "DHjqpvDnNGE", "video_caption": "JavaScript in 100 Seconds"},
                {"kind": "text", "markdown":
"""## Your first line of code

The classic way to make a program *say something* is `console.log`, which prints a value to the output:

```js
console.log("Hello, world!");
```

Run that and you'll see:

```text
Hello, world!
```

`console.log` is your most useful tool while learning: print things to see what your program is doing.
"""},
                {"kind": "quiz", "quiz_ref": "getting-started-quiz"},
            ],
        },
    },
    {
        "file": "02-variables-and-types.md",
        "front": {
            "title": "Variables and Data Types",
            "order": 1,
            "blocks": [
                {"kind": "text", "markdown":
"""## Storing values in variables

A **variable** is a named box that holds a value. In modern JavaScript you create one with `let` (a value that can change) or `const` (a value that should not be reassigned):

```js
let score = 0;        // can change later
const name = "Ada";   // should stay the same
score = 10;           // fine
// name = "Bob";      // error: cannot reassign a const
```

Prefer `const` by default, and reach for `let` only when you genuinely need to reassign.
"""},
                {"kind": "text", "markdown":
"""## The basic data types

Every value in JavaScript has a **type**. The ones you'll meet first are:

| Type | Example | Meaning |
| --- | --- | --- |
| `string` | `"hello"` | text, in quotes |
| `number` | `42`, `3.14` | integers and decimals |
| `boolean` | `true`, `false` | yes/no values |
| `undefined` | `undefined` | a variable with no value yet |
| `null` | `null` | a deliberate "nothing" |

You can check a value's type with `typeof`:

```js
console.log(typeof "hello"); // "string"
console.log(typeof 42);      // "number"
console.log(typeof true);    // "boolean"
```
"""},
                {"kind": "video", "video_provider": "youtube",
                 "video_locator": "9emXNzqCKyg", "video_caption": "Variables and data types"},
                {"kind": "quiz", "quiz_ref": "variables-quiz"},
                {"kind": "practice", "practice_ref": "count-to-five"},
            ],
        },
    },
    {
        "file": "03-functions-and-flow.md",
        "front": {
            "title": "Functions and Control Flow",
            "order": 2,
            "blocks": [
                {"kind": "text", "markdown":
"""## Functions: reusable instructions

A **function** is a named block of code you can run whenever you want. It can take **inputs** (parameters) and `return` a result:

```js
function greet(personName) {
  return "Hello, " + personName + "!";
}

console.log(greet("Ada")); // "Hello, Ada!"
```

Define a function once, then call it as many times as you like.
"""},
                {"kind": "text", "markdown":
"""## Making decisions and repeating work

**`if` / `else`** lets your program choose between paths:

```js
const hour = 9;
if (hour < 12) {
  console.log("Good morning");
} else {
  console.log("Good afternoon");
}
```

**Loops** repeat work. A `for` loop runs a counter through a range:

```js
for (let i = 1; i <= 3; i++) {
  console.log(i); // prints 1, then 2, then 3
}
```

With variables, functions, conditionals, and loops you can already build real little programs, like the practice below.
"""},
                {"kind": "quiz", "quiz_ref": "functions-quiz"},
                {"kind": "practice", "practice_ref": "fizzbuzz-lite"},
            ],
        },
    },
]

QUIZZES = [
    {"slug": "getting-started-quiz", "title": "Getting Started Quiz", "pass_threshold": 0.7,
     "questions": [
        {"type": "single", "prompt": "What does console.log do?",
         "options": ["Prints a value to the output", "Deletes a file", "Creates a new variable", "Stops the program"],
         "correct_indices": [0], "explanation": "console.log prints the value you give it to the output, which is how you inspect what your program is doing."},
        {"type": "single", "prompt": "Name one place JavaScript can run.",
         "options": ["Only inside spreadsheets", "In the browser (or in Node.js)", "Only on Apple devices", "Nowhere without the internet"],
         "correct_indices": [1], "explanation": "JavaScript runs in the browser and, via the Node.js runtime, directly on your machine or a server."},
     ]},
    {"slug": "variables-quiz", "title": "Variables and Types Quiz", "pass_threshold": 0.7,
     "questions": [
        {"type": "single", "prompt": "Which keyword declares a value that should NOT be reassigned?",
         "options": ["var", "let", "const", "static"],
         "correct_indices": [2], "explanation": "const declares a binding you do not intend to reassign. Use let when a value genuinely needs to change."},
        {"type": "multiple", "prompt": "Which of these are JavaScript data types you met in this lesson? (Select all that apply.)",
         "options": ["string", "number", "boolean", "spreadsheet"],
         "correct_indices": [0, 1, 2], "explanation": "string, number, and boolean are core JavaScript types. 'spreadsheet' is not a type."},
        {"type": "single", "prompt": "What does typeof 42 return?",
         "options": ['"string"', '"number"', '"boolean"', '"digit"'],
         "correct_indices": [1], "explanation": "42 is a number, so typeof 42 evaluates to the string \"number\"."},
     ]},
    {"slug": "functions-quiz", "title": "Functions and Control Flow Quiz", "pass_threshold": 0.7,
     "questions": [
        {"type": "single", "prompt": "What does the return keyword do inside a function?",
         "options": ["Prints text to the screen", "Sends a value back to whoever called the function", "Restarts the program", "Declares a variable"],
         "correct_indices": [1], "explanation": "return hands a value back to the code that called the function, so it can be used or stored."},
        {"type": "single", "prompt": "How many times does this loop run: for (let i = 1; i <= 3; i++) ?",
         "options": ["2 times", "3 times", "4 times", "Forever"],
         "correct_indices": [1], "explanation": "i takes the values 1, 2, and 3, so the loop body runs three times."},
     ]},
]

PRACTICES = [
    {"slug": "count-to-five", "title": "Count to Five", "language": "javascript",
     "prompt": "Print the numbers 1 through 5, each on its own line, using a for loop and console.log.",
     "starter_code": "// Print 1, 2, 3, 4, 5 — one number per line.\n// Hint: a for loop that counts from 1 to 5.\n",
     "solution": "for (let i = 1; i <= 5; i++) {\n  console.log(i);\n}\n",
     "test_cases": [
        {"name": "prints one through five", "expected_stdout": "1\n2\n3\n4\n5\n"},
     ]},
    {"slug": "fizzbuzz-lite", "title": "FizzBuzz (1 to 15)", "language": "javascript",
     "prompt": "For each number from 1 to 15, print 'Fizz' if it is divisible by 3, 'Buzz' if divisible by 5, 'FizzBuzz' if divisible by both, otherwise the number itself. One result per line.",
     "starter_code": "// For 1..15: Fizz (÷3), Buzz (÷5), FizzBuzz (both), else the number.\nfor (let i = 1; i <= 15; i++) {\n  // your logic here\n}\n",
     "solution": "for (let i = 1; i <= 15; i++) {\n  if (i % 15 === 0) console.log(\"FizzBuzz\");\n  else if (i % 3 === 0) console.log(\"Fizz\");\n  else if (i % 5 === 0) console.log(\"Buzz\");\n  else console.log(i);\n}\n",
     "test_cases": [
        {"name": "classic fizzbuzz to 15",
         "expected_stdout": "1\n2\nFizz\n4\nBuzz\nFizz\n7\n8\nFizz\nBuzz\n11\nFizz\n13\n14\nFizzBuzz\n"},
     ]},
]

TESTS = [
    {"slug": "foundations-checkpoint", "title": "Foundations Checkpoint",
     "time_limit_minutes": 20, "pass_threshold": 0.7,
     "items": [
        {"kind": "choice", "prompt": "Which keyword declares a block-scoped variable that cannot be reassigned?",
         "choice_type": "single", "options": ["var", "let", "const", "function"],
         "correct_indices": [2], "explanation": "const creates a block-scoped binding you cannot reassign."},
        {"kind": "choice", "prompt": "Which of these are primitive data types in JavaScript? (Select all that apply.)",
         "choice_type": "multiple", "options": ["string", "number", "array", "boolean"],
         "correct_indices": [0, 1, 3], "explanation": "string, number, and boolean are primitives. An array is an object, not a primitive."},
        {"kind": "coding", "coding_prompt": "Print the even numbers from 2 to 10, each on its own line.",
         "language": "javascript",
         "starter_code": "// Print 2, 4, 6, 8, 10 — one per line.\n",
         "solution": "for (let i = 2; i <= 10; i += 2) {\n  console.log(i);\n}\n",
         "test_cases": [
            {"name": "evens 2..10", "expected_stdout": "2\n4\n6\n8\n10\n"},
         ]},
     ]},
]

# ------------------------------------------------------------------ emit
if os.path.exists(SRC): shutil.rmtree(SRC, ignore_errors=True)
os.makedirs(os.path.join(SRC, "lessons"))
os.makedirs(os.path.join(SRC, "quizzes"))
os.makedirs(os.path.join(SRC, "practices"))
os.makedirs(os.path.join(SRC, "tests"))

def w(path, text):
    with open(os.path.join(SRC, path), "w") as f: f.write(text)

w("format_version.txt", "1\n")
w("course.yaml", dump(COURSE))
for L in LESSONS:
    w(os.path.join("lessons", L["file"]), "---\n" + dump(L["front"]) + "---\n")
for q in QUIZZES:
    w(os.path.join("quizzes", q["slug"] + ".yaml"), dump(q))
for p in PRACTICES:
    w(os.path.join("practices", p["slug"] + ".yaml"), dump(p))
for t in TESTS:
    w(os.path.join("tests", t["slug"] + ".yaml"), dump(t))

# zip (relative paths at archive root, sorted for determinism)
files = []
for root, _, fns in os.walk(SRC):
    for fn in fns:
        full = os.path.join(root, fn)
        files.append((full, os.path.relpath(full, SRC).replace(os.sep, "/")))
files.sort(key=lambda x: x[1])
if os.path.exists(ZIP): os.remove(ZIP)
with zipfile.ZipFile(ZIP, "w", zipfile.ZIP_DEFLATED) as z:
    for full, rel in files:
        z.write(full, rel)

print("Wrote source tree:", SRC)
print("Wrote zip:", ZIP)
print("Archive entries:")
for _, rel in files: print("  ", rel)
