package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/Wei-Shaw/sub2api/migrations"
)

// 一次性修复工具：当已应用的迁移文件被修改导致 checksum 不匹配时，
// 将数据库 schema_migrations 记录的 checksum 更新为当前嵌入文件的值。
// 仅用于开发环境排查，生产环境应遵循不可变迁移原则。
func main() {
	only := ""
	if len(os.Args) > 1 {
		only = os.Args[1]
	}

	files, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		fmt.Printf("列出迁移失败: %v\n", err)
		os.Exit(1)
	}
	sort.Strings(files)

	dsn := "host=127.0.0.1 port=5432 user=sub2api password=sub2api dbname=sub2api sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fixed := 0
	for _, name := range files {
		if only != "" && !strings.Contains(name, only) {
			continue
		}
		contentBytes, err := migrations.FS.ReadFile(name)
		if err != nil {
			continue
		}
		content := strings.TrimSpace(string(contentBytes))
		if content == "" {
			continue
		}
		sum := sha256.Sum256([]byte(content))
		newChecksum := hex.EncodeToString(sum[:])

		var old string
		err = db.QueryRowContext(ctx, "SELECT checksum FROM schema_migrations WHERE filename = $1", name).Scan(&old)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			fmt.Printf("查询 %s 失败: %v\n", name, err)
			continue
		}
		if old == newChecksum {
			continue
		}
		_, err = db.ExecContext(ctx, "UPDATE schema_migrations SET checksum = $1 WHERE filename = $2", newChecksum, name)
		if err != nil {
			fmt.Printf("更新 %s 失败: %v\n", name, err)
			continue
		}
		fmt.Printf("已修复 %s（db=%s -> file=%s）\n", name, old, newChecksum)
		fixed++
	}
	fmt.Printf("完成，共修复 %d 个迁移 checksum\n", fixed)
}
