---
title: Functions and Control Flow
order: 2
blocks:
- kind: text
  markdown: |
    ## Functions: reusable instructions

    A **function** is a named block of code you can run whenever you want. It can take **inputs** (parameters) and `return` a result:

    ```js
    function greet(personName) {
      return "Hello, " + personName + "!";
    }

    console.log(greet("Ada")); // "Hello, Ada!"
    ```

    Define a function once, then call it as many times as you like.
- kind: text
  markdown: |
    ## Making decisions and repeating work

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
- kind: quiz
  quiz_ref: functions-quiz
- kind: practice
  practice_ref: fizzbuzz-lite
---
