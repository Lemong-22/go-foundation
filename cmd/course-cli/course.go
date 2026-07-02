package main

import (
	"fmt"
	"time"

	"github.com/lemong-22/go-foundation/internal/course"
	"github.com/spf13/cobra"
)

// Ini adalah "Database Palsu" kita di RAM
var dbInMemory []course.Course

// Flag variabel untuk menampung input user
var titleFlag string
var slugFlag string

func init() {
	// 1. Bikin sub-command utama bernama "course"
	courseCmd := &cobra.Command{
		Use:   "course",
		Short: "Manajemen data course",
	}

	// 2. Bikin aksi "create"
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Bikin course baru",
		Run: func(cmd *cobra.Command, args []string) {
			// Bikin objek course baru dari input flag
			newCourse := course.Course{
				ID:        fmt.Sprintf("CRSM-%d", len(dbInMemory)+1), // ID palsu urutan
				Title:     titleFlag,
				Slug:      slugFlag,
				Status:    "draft",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Simpan ke database RAM (append kayak push di JS)
			dbInMemory = append(dbInMemory, newCourse)
			fmt.Printf("Mantap! Course sukses dibuat dengan ID: %s\n", newCourse.ID)
		},
	}

	// Daftarin input wajib (--title dan --slug) ke command create
	createCmd.Flags().StringVar(&titleFlag, "title", "", "Judul course")
	createCmd.Flags().StringVar(&slugFlag, "slug", "", "Slug url course")
	createCmd.MarkFlagRequired("title")
	createCmd.MarkFlagRequired("slug")

	// 3. Bikin aksi "list"
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Tampilkan semua course",
		Run: func(cmd *cobra.Command, args []string) {
			if len(dbInMemory) == 0 {
				fmt.Println("Belum ada course yang dibuat. Kosong melompong!")
				return
			}

			for _, c := range dbInMemory {
				fmt.Printf("[%s] %s (Slug: %s) - Status: %s\n", c.ID, c.Title, c.Slug, c.Status)
			}
		},
	}

	// Gabungin semua potongan perintah ke root utama
	courseCmd.AddCommand(createCmd, listCmd)
	rootCmd.AddCommand(courseCmd)
}
