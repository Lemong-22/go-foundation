package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lemong-22/go-foundation/internal/course"
	"github.com/spf13/cobra"
)

// Flag variabel untuk sub-command lesson.
var (
	lessonCourseIDFlag  string
	lessonTitleFlag     string
	lessonSlugFlag      string
	lessonOrderFlag     int
	lessonIDFlag        string
	outputFlagLesson    string // "json" atau "" (default human-readable)
)

// newLessonIDgenerator bikin ID sederhana berbasis timestamp.
// Format: LES-<unix-nano>. Sama gaya dengan CRSM-* di course.go.
func newLessonID() string {
	return fmt.Sprintf("LES-%d", time.Now().UnixNano())
}

// slugifyLesson: rapikan spasi jadi dash, fallback kalau kosong.
// Duplikasi kecil dari course.go slugify — di-scope terpisah agar tiap CLI
// file berdiri sendiri dan gak ada import cycle di package main.
func slugifyLesson(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	return strings.ToLower(s)
}

func init() {
	// 1. Sub-command utama "lesson".
	lessonCmd := &cobra.Command{
		Use:   "lesson",
		Short: "Manajemen data lesson (child of course)",
	}

	// 2. create — bikin lesson baru di bawah course tertentu.
	createLessonCmd := &cobra.Command{
		Use:   "create",
		Short: "Bikin lesson baru di bawah sebuah course",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := newRepo()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			// Validasi parent course exist dulu — kalau gak ada, FK Postgres bakal
			// nolak dengan error cryptic. Lebih baik fail-fast dengan pesan jelas.
			if _, err := repo.FindByID(ctx, lessonCourseIDFlag); err != nil {
				if errors.Is(err, course.ErrNotFound) {
					fmt.Fprintf(os.Stderr, "course %q tidak ditemukan — bikin course dulu sebelum tambah lesson\n",
						lessonCourseIDFlag)
				} else {
					fmt.Fprintf(os.Stderr, "gagal verifikasi course: %v\n", err)
				}
				os.Exit(1)
			}

			now := time.Now().UTC()
			newLesson := course.Lesson{
				ID:         newLessonID(),
				CourseID:   lessonCourseIDFlag,
				Title:      lessonTitleFlag,
				Slug:       slugifyLesson(lessonSlugFlag),
				Content:    "", // belum ada flag-nya, default kosong
				OrderIndex: lessonOrderFlag,
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			if err := repo.SaveLesson(ctx, &newLesson); err != nil {
				fmt.Fprintf(os.Stderr, "gagal save lesson: %v\n", err)
				os.Exit(1)
			}

			if outputFlagLesson == "json" {
				json.NewEncoder(os.Stdout).Encode(newLesson)
			} else {
				fmt.Printf("Mantap! Lesson sukses dibuat dengan ID: %s\n", newLesson.ID)
			}
			os.Exit(0)
			return nil
		},
	}
	createLessonCmd.Flags().StringVar(&lessonCourseIDFlag, "course-id", "",
		"ID course induk (e.g. CRSM-1234567890)")
	createLessonCmd.Flags().StringVar(&lessonTitleFlag, "title", "", "Judul lesson")
	createLessonCmd.Flags().StringVar(&lessonSlugFlag, "slug", "", "Slug url lesson")
	createLessonCmd.Flags().IntVar(&lessonOrderFlag, "order", 0,
		"Urutan lesson dalam course (0 = pertama)")
	createLessonCmd.Flags().StringVar(&outputFlagLesson, "output", "", "Format output: json")
	_ = createLessonCmd.MarkFlagRequired("course-id")
	_ = createLessonCmd.MarkFlagRequired("title")
	_ = createLessonCmd.MarkFlagRequired("slug")

	// 3. list — tampilkan semua lesson untuk course tertentu.
	listLessonCmd := &cobra.Command{
		Use:   "list",
		Short: "Tampilkan semua lesson milik sebuah course",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := newRepo()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			items, err := repo.FindLessonsByCourseID(ctx, lessonCourseIDFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "gagal FindLessonsByCourseID: %v\n", err)
				os.Exit(1)
			}

			if outputFlagLesson == "json" {
				json.NewEncoder(os.Stdout).Encode(items)
				os.Exit(0)
			}

			if len(items) == 0 {
				fmt.Printf("Course %q belum punya lesson. Kosong melompong!\n", lessonCourseIDFlag)
				os.Exit(0)
			}
			fmt.Printf("Lessons untuk course %s:\n", lessonCourseIDFlag)
			for _, l := range items {
				fmt.Printf("  [%d] [%s] %s (Slug: %s)\n",
					l.OrderIndex, l.ID, l.Title, l.Slug)
			}
			os.Exit(0)
			return nil
		},
	}
	listLessonCmd.Flags().StringVar(&lessonCourseIDFlag, "course-id", "",
		"ID course induk (e.g. CRSM-1234567890)")
	listLessonCmd.Flags().StringVar(&outputFlagLesson, "output", "", "Format output: json")
	_ = listLessonCmd.MarkFlagRequired("course-id")

	// 4. delete — hapus lesson berdasarkan id.
	deleteLessonCmd := &cobra.Command{
		Use:   "delete",
		Short: "Hapus lesson by id",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := newRepo()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			if err := repo.DeleteLesson(ctx, lessonIDFlag); err != nil {
				fmt.Fprintf(os.Stderr, "DeleteLesson(%q): %v\n", lessonIDFlag, err)
				os.Exit(1)
			}
			fmt.Printf("Berhasil hapus lesson %s\n", lessonIDFlag)
			os.Exit(0)
			return nil
		},
	}
	deleteLessonCmd.Flags().StringVar(&lessonIDFlag, "id", "",
		"ID lesson (e.g. LES-1234567890)")
	_ = deleteLessonCmd.MarkFlagRequired("id")

	// Gabungin semua sub-command ke root.
	lessonCmd.AddCommand(createLessonCmd, listLessonCmd, deleteLessonCmd)
	rootCmd.AddCommand(lessonCmd)
}
