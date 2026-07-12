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

// Flag variabel untuk menampung input user.
var (
	titleFlag      string
	slugFlag       string
	idFlag         string
	outputFlagCourse string // "json" atau "" (default human-readable)
)

// newRepo adalah helper yang bikin repository Course dari dbPool global.
// Return error kalau dbPool belum dibuka (DATABASE_URL gak di-set).
func newRepo() (course.CourseRepository, error) {
	if dbPool == nil {
		return nil, errors.New("DATABASE_URL belum di-set; tidak bisa pakai repository. " +
			"Export DATABASE_URL lalu jalankan ulang.")
	}
	return course.NewPostgresCourseRepository(dbPool), nil
}

// newIDgenerator bikin ID sederhana berbasis timestamp.
// Format: CRSM-<unix-nano>. Bukan UUID tapi cukup unik untuk CLI local.
func newID() string {
	return fmt.Sprintf("CRSM-%d", time.Now().UnixNano())
}

// slugify: rapikan spasi jadi dash, fallback kalau kosong.
func slugify(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	return strings.ToLower(s)
}

func init() {
	// 1. Sub-command utama "course".
	courseCmd := &cobra.Command{
		Use:   "course",
		Short: "Manajemen data course",
	}

	// 2. create — bikin course baru, simpan ke Postgres.
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Bikin course baru (persist ke Postgres)",
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

			now := time.Now().UTC()
			newCourse := course.Course{
				ID:          newID(),
				Title:       titleFlag,
				Slug:        slugify(slugFlag),
				Description: "", // belum ada flag-nya, default kosong
				Status:      "draft",
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if err := repo.Save(ctx, &newCourse); err != nil {
				fmt.Fprintf(os.Stderr, "gagal save course: %v\n", err)
				os.Exit(1)
			}

			if outputFlagCourse == "json" {
				json.NewEncoder(os.Stdout).Encode(newCourse)
			} else {
				fmt.Printf("Mantap! Course sukses dibuat dengan ID: %s\n", newCourse.ID)
			}
			os.Exit(0)
			return nil
		},
	}
	createCmd.Flags().StringVar(&titleFlag, "title", "", "Judul course")
	createCmd.Flags().StringVar(&slugFlag, "slug", "", "Slug url course")
	createCmd.Flags().StringVar(&outputFlagCourse, "output", "", "Format output: json")
	_ = createCmd.MarkFlagRequired("title")
	_ = createCmd.MarkFlagRequired("slug")

	// 3. list — tampilkan semua course dari Postgres.
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Tampilkan semua course",
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

			items, err := repo.FindAll(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "gagal FindAll: %v\n", err)
				os.Exit(1)
			}

			if outputFlagCourse == "json" {
				json.NewEncoder(os.Stdout).Encode(items)
				os.Exit(0)
			}

			if len(items) == 0 {
				fmt.Println("Belum ada course yang dibuat. Kosong melompong!")
				os.Exit(0)
			}
			for _, c := range items {
				fmt.Printf("[%s] %s (Slug: %s) - Status: %s\n",
					c.ID, c.Title, c.Slug, c.Status)
			}
			os.Exit(0)
			return nil
		},
	}
	listCmd.Flags().StringVar(&outputFlagCourse, "output", "", "Format output: json")

	// 4. find — ambil 1 course by id.
	findCmd := &cobra.Command{
		Use:   "find",
		Short: "Cari course by id",
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

			found, err := repo.FindByID(ctx, idFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "FindByID(%q): %v\n", idFlag, err)
				os.Exit(1)
			}

			if outputFlagCourse == "json" {
				json.NewEncoder(os.Stdout).Encode(found)
			} else {
				fmt.Printf("Found: [%s] %s\n  Slug:        %s\n  Status:      %s\n  Description: %s\n  CreatedAt:   %s\n  UpdatedAt:   %s\n",
					found.ID, found.Title, found.Slug, found.Status,
					found.Description, found.CreatedAt, found.UpdatedAt)
			}
			os.Exit(0)
			return nil
		},
	}
	findCmd.Flags().StringVar(&idFlag, "id", "", "ID course (e.g. CRSM-1234567890)")
	findCmd.Flags().StringVar(&outputFlagCourse, "output", "", "Format output: json")
	_ = findCmd.MarkFlagRequired("id")

	// 5. delete — hapus course by id.
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Hapus course by id",
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

			if err := repo.Delete(ctx, idFlag); err != nil {
				fmt.Fprintf(os.Stderr, "Delete(%q): %v\n", idFlag, err)
				os.Exit(1)
			}
			fmt.Printf("Berhasil hapus course %s\n", idFlag)
			os.Exit(0)
			return nil
		},
	}
	deleteCmd.Flags().StringVar(&idFlag, "id", "", "ID course (e.g. CRSM-1234567890)")
	_ = deleteCmd.MarkFlagRequired("id")

	// Gabungin semua sub-command ke root.
	courseCmd.AddCommand(createCmd, listCmd, findCmd, deleteCmd)
	rootCmd.AddCommand(courseCmd)
}
