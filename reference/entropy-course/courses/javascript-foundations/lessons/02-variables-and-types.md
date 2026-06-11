---
title: Variables and Data Types
order: 1
blocks:
- kind: text
  markdown: |
    ## Storing values in variables

    A **variable** is a named box that holds a value. In modern JavaScript you create one with `let` (a value that can change) or `const` (a value that should not be reassigned):

    ```js
    let score = 0;        // can change later
    const name = "Ada";   // should stay the same
    score = 10;           // fine
    // name = "Bob";      // error: cannot reassign a const
    ```

    Prefer `const` by default, and reach for `let` only when you genuinely need to reassign.
- kind: text
  markdown: |
    ## The basic data types

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
- kind: video
  video_provider: youtube
  video_locator: 9emXNzqCKyg
  video_caption: Variables and data types
- kind: quiz
  quiz_ref: variables-quiz
- kind: practice
  practice_ref: count-to-five
---
