"use client";

import { useState } from "react";
import { Button } from "@entropy-course/ui/components/button";
import { Card } from "@entropy-course/ui/components/card";
import type { QuizQuestion } from "@/lib/mock-data";

interface Props {
  questions: QuizQuestion[];
}

export function QuizModule({ questions }: Props) {
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [submitted, setSubmitted] = useState(false);

  const handleSelect = (questionId: string, answer: string) => {
    if (submitted) return;
    setAnswers((prev) => ({ ...prev, [questionId]: answer }));
  };

  const handleSubmit = () => setSubmitted(true);

  const score = questions.reduce((acc, q) => {
    const userAnswer = answers[q.id];
    const isCorrect = Array.isArray(q.correctAnswer)
      ? q.correctAnswer.includes(userAnswer ?? "")
      : userAnswer === q.correctAnswer;
    return acc + (isCorrect ? 1 : 0);
  }, 0);

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold flex items-center gap-2">
        <span className="text-yellow-500">❓</span> Quiz
      </h2>
      <div className="space-y-4">
        {questions.map((q, qi) => {
          const userAnswer = answers[q.id];
          const isCorrect = Array.isArray(q.correctAnswer)
            ? q.correctAnswer.includes(userAnswer ?? "")
            : userAnswer === q.correctAnswer;
          const showResult = submitted;

          return (
            <Card key={q.id} className="p-4 space-y-3">
              <p className="font-medium text-sm whitespace-pre-wrap">
                {qi + 1}. {q.question}
              </p>
              <div className="space-y-2">
                {(q.options ?? []).map((option) => {
                  const isSelected = answers[q.id] === option;
                  const isCorrectOption = Array.isArray(q.correctAnswer)
                    ? q.correctAnswer.includes(option)
                    : option === q.correctAnswer;

                  let bgClass = "bg-muted/50 hover:bg-muted border-transparent";
                  if (showResult) {
                    if (isCorrectOption)
                      bgClass =
                        "bg-green-100 dark:bg-green-900 border-green-500 text-green-800 dark:text-green-200";
                    else if (isSelected && !isCorrectOption)
                      bgClass =
                        "bg-red-100 dark:bg-red-900 border-red-500 text-red-800 dark:text-red-200";
                  } else if (isSelected) {
                    bgClass = "bg-primary/10 border-primary";
                  }

                  return (
                    <button
                      key={option}
                      onClick={() => handleSelect(q.id, option)}
                      className={`w-full text-left px-3 py-2 rounded-lg border text-sm transition-colors ${bgClass}`}
                      disabled={submitted}
                    >
                      {option}
                    </button>
                  );
                })}
              </div>
              {showResult && (
                <div
                  className={`p-3 rounded-lg text-sm ${
                    isCorrect
                      ? "bg-green-50 dark:bg-green-950 text-green-700 dark:text-green-300"
                      : "bg-red-50 dark:bg-red-950 text-red-700 dark:text-red-300"
                  }`}
                >
                  <p className="font-medium mb-1">
                    {isCorrect ? "✓ Correct!" : "✗ Incorrect"}
                  </p>
                  <p className="text-muted-foreground">{q.explanation}</p>
                </div>
              )}
            </Card>
          );
        })}
      </div>
      {!submitted ? (
        <Button
          onClick={handleSubmit}
          disabled={Object.keys(answers).length < questions.length}
        >
          Submit Answers
        </Button>
      ) : (
        <div className="p-4 rounded-xl bg-primary/10 border border-primary/20">
          <p className="font-medium">
            You scored{" "}
            <span className="text-primary text-lg">{score}</span> out of{" "}
            <span>{questions.length}</span>
          </p>
        </div>
      )}
    </div>
  );
}
